import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { WebSocketSubject, webSocket } from 'rxjs/webSocket';
import type { User, Message, Opcode, GatewayPayloadStructure, Channel, EventName, Event } from '@satorijs/protocol'
import type { APIChannelCreateResp, APIChannelListResp, APIMessage } from '@/types';
import { nanoid } from 'nanoid'
import { groupBy } from 'lodash-es';
import { Emitter } from '@/utils/event';
import { useUserStore } from './user';

interface ChatState {
  subject: WebSocketSubject<any> | null;
  // user: User,
  channelTree: any,
}

const apiMap = new Map<string, any>();

type myEventName = EventName | 'message-created'; // 当前npm版本缺这个事件
export const chatEvent = new Emitter<{
  [key in myEventName]: (msg: Event) => void;
  // 'message-created': (msg: Event) => void;
}>();


export const useChatStore = defineStore({
  id: 'chat',
  state: (): ChatState => ({
    // user: { id: '1', },
    subject: null,
    channelTree: [],
  }),

  actions: {
    async init() {
      const u: User = {
        id: '',
      }
      const subject = webSocket('ws://localhost:3212/ws/seal');

      let isReady = false;

      // 发送协议握手
      // Opcode.IDENTIFY: 3
      const user = useUserStore();
      subject.next({ op: 3, body: {
        token: user.token,
      }});

      subject.subscribe({
        next: (msg: any) => {
          // Opcode.READY
          if (msg.op === 4) {
            console.log('svr ready', msg);
            isReady = true
            this.wsReady();
          } else if (msg.op === 0) {
            // Opcode.EVENT
            const e = msg as Event;
            this.eventDispatch(e);
          } else if (apiMap.get(msg.echo)) {
            apiMap.get(msg.echo).resolve(msg);
            apiMap.delete(msg.echo);
          }
        },
        error: err => console.log(err), // Called if at any point WebSocket API signals some kind of error.
        complete: () => console.log('complete') // Called when connection is closed (for whatever reason).
      });

      this.subject = subject;
    },

    async reinit() {
      if (!this.subject) {
        await this.init();
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

    async wsReady() {
      // const resp = await this.sendAPI('channel.create', { name: '测试频道3' }) as APIChannelCreateResp;
      // console.log(1111, resp);
      this.channelList();
    },
    
    async channelCreate(name: string) {
      const resp = await this.sendAPI('channel.create', { name: '测试频道' }) as APIChannelCreateResp;
    },

    async channelList() {
      const resp = await this.sendAPI('channel.list', {}) as APIChannelListResp;

      const groupedData = groupBy(resp.data, 'parent_id');
      const buildTree = (parentId: string): any => {
        const children = groupedData[parentId] || [];
        return children.map((child: Channel) => ({
          ...child,
          children: buildTree(child.id),
        }));
      };

      const tree = buildTree('');
      this.channelTree = tree;
      return tree;
    },

    async messageList(channelId: string) {
      const resp = await this.sendAPI('message.list', { channel_id: channelId });
      return resp;
    },

    async messageCreate(content: string) {
      const resp = await this.sendAPI('message.create', { channel_id: this.channelTree[0].id, content });
      console.log(1111, resp)
    },

    async eventDispatch(e: Event) {
      chatEvent.emit(e.type as any, e);
    }
  }
});
