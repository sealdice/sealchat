import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { WebSocketSubject, webSocket } from 'rxjs/webSocket';
import type { User, Message, Opcode, GatewayPayloadStructure, Channel, EventName, Event } from '@satorijs/protocol'
import type { APIChannelCreateResp, APIChannelListResp, APIMessage } from '@/types';
import { nanoid } from 'nanoid'
import { groupBy } from 'lodash-es';
import { Emitter } from '@/utils/event';
import { useUserStore } from './user';
import { urlBase } from './_config';
import { useMessage } from 'naive-ui';

interface ChatState {
  subject: WebSocketSubject<any> | null;
  // user: User,
  channelTree: Channel[],
  curChannel: Channel | null,
  connectState: 'connecting' | 'connected' | 'disconnected' | 'reconnecting',
  iReconnectAfterTime: number,
}

const apiMap = new Map<string, any>();
let _connectResolve: any = null;

type myEventName = EventName | 'message-created'; // 当前npm版本缺这个事件
export const chatEvent = new Emitter<{
  [key in myEventName]: (msg?: Event) => void;
  // 'message-created': (msg: Event) => void;
}>();

let pingLoopOn = false;

export const useChatStore = defineStore({
  id: 'chat',
  state: (): ChatState => ({
    // user: { id: '1', },
    subject: null,
    channelTree: [],
    curChannel: null,
    connectState: 'connecting',
    iReconnectAfterTime: 0,
  }),

  getters: {
    _lastChannel: (state) => {
      return localStorage.getItem('lastChannel') || '';
    }
  },

  actions: {
    async connect() {
      const u: User = {
        id: '',
      }
      this.connectState = 'connecting';

      // 'ws://localhost:3212/ws/seal'
      const subject = webSocket(`ws:${urlBase}/ws/seal`);

      let isReady = false;

      // 发送协议握手
      // Opcode.IDENTIFY: 3
      const user = useUserStore();
      subject.next({
        op: 3, body: {
          token: user.token,
        }
      });

      subject.subscribe({
        next: (msg: any) => {
          // Opcode.READY
          if (msg.op === 4) {
            console.log('svr ready', msg);
            isReady = true
            this.connectReady();
          } else if (msg.op === 0) {
            // Opcode.EVENT
            const e = msg as Event;
            this.eventDispatch(e);
          } else if (apiMap.get(msg.echo)) {
            apiMap.get(msg.echo).resolve(msg);
            apiMap.delete(msg.echo);
          }
        },
        error: err => {
          console.log('ws error', err);
          this.subject = null;
          this.connectState = 'disconnected';
          this.reconnectAfter(5, () => {
            try {
              err.target?.close();
              this.subject?.unsubscribe();
              console.log('try close');
            } catch (e) {
              console.log('unsubscribe error', e)
            }
          })
        }, // Called if at any point WebSocket API signals some kind of error.
        complete: () => console.log('complete') // Called when connection is closed (for whatever reason).
      });

      this.subject = subject;
    },

    async reconnectAfter(secs: number, beforeConnect?: Function) {
      setTimeout(async () => {
        this.connectState = 'reconnecting';
        // alert(`连接已断开，${secs} 秒后自动重连`);
        for (let i = secs; i > 0; i--) {
          this.iReconnectAfterTime = i;
          await new Promise(resolve => setTimeout(resolve, 1000));
        }
        if (beforeConnect) beforeConnect();
        this.connect();
      }, 500);
    },

    async connectReady() {
      this.connectState = 'connected';

      if (!pingLoopOn) {
        pingLoopOn = true;
        const user = useUserStore();
        setInterval(async () => {
          if (this.subject) {
            this.subject.next({
              op: 1, body: {
                token: user.token,
              }
            });
          }
        }, 10000)
      }

      await this.channelList();
      if (_connectResolve) {
        _connectResolve();
        _connectResolve = null;
      }
    },

    /** try to initialize */
    async tryInit() {
      if (!this.subject) {
        return new Promise((resolve) => {
          _connectResolve = resolve;
          this.connect();
        });
      }
    },

    async sendAPI(api: string, data: APIMessage): Promise<any> {
      const echo = nanoid();
      return new Promise((resolve, reject) => {
        apiMap.set(echo, { resolve, reject });
        this.subject?.next({ api, data, echo });
      })
    },

    async send(channelId: string, content: string) {
      let msg: APIMessage = {
        // api: 'message.create',
        channel_id: channelId,
        content: content
      }
      this.subject?.next(msg);
    },

    async channelCreate(name: string) {
      const resp = await this.sendAPI('channel.create', { name }) as APIChannelCreateResp;
    },

    async channelSwitchTo(id: string) {
      this.curChannel = this.channelTree.find(c => c.id === id) || this.curChannel;
      await this.sendAPI('channel.enter', { 'channel_id': id });
      localStorage.setItem('lastChannel', id);
      this.channelList();
    },

    async channelList() {
      const resp = await this.sendAPI('channel.list', {}) as APIChannelListResp;

      const curItem = resp.data.find(c => c.id === this.curChannel?.id);
      this.curChannel = curItem || this.curChannel;

      const groupedData = groupBy(resp.data, 'parentId');
      const buildTree = (parentId: string): any => {
        const children = groupedData[parentId] || [];
        return children.map((child: Channel) => ({
          ...child,
          children: buildTree(child.id),
        }));
      };

      const tree = buildTree('');
      this.channelTree = tree;

      if (!this.curChannel) {
        // 这是为了正确标记人数，有点屎但实现了
        const lastChannel = this._lastChannel;
        const c = this.channelTree.find(c => c.id === lastChannel);
        if (c) {
          this.channelSwitchTo(c.id);
        } else {
          this.channelSwitchTo(tree[0].id);
        }
      }

      return tree;
    },

    async channelRefresh() {
      setInterval(async () => {
        await this.channelList();
        if (!this.channelTree.find(c => c.id === this.curChannel?.id)) {
          this.curChannel = this.channelTree[0];
          chatEvent.emit('channel-deleted', undefined)
        }
      }, 20000);
    },

    async messageList(channelId: string, next?: string) {
      const resp = await this.sendAPI('message.list', { channel_id: channelId, next });
      return resp;
    },

    async messageCreate(content: string) {
      // const resp = await this.sendAPI('message.create', { channel_id: this.curChannel?.id, content });
      const resp = await this.sendAPI('qqq.x', { channel_id: this.curChannel?.id, content });
      console.log(1111, resp)
    },

    async eventDispatch(e: Event) {
      chatEvent.emit(e.type as any, e);
    }
  }
});
