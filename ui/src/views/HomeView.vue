<script setup lang="tsx">
import dayjs from 'dayjs';
import imgAvatar from '@/assets/head2.png'
import ChatItem from '@/components/chat-item.vue';
import { computed, ref, watch, h, onMounted, onBeforeMount, nextTick, type Component, inject } from 'vue'
import { VirtualList } from 'vue-tiny-virtual-list';
import { chatEvent, useChatStore } from '@/stores/chat';
import type { Event, Message } from '@satorijs/protocol'
import { useUserStore } from '@/stores/user';
import { ArrowBarToDown, Plus, Upload } from '@vicons/tabler'
import { NIcon, c, useDialog, useMessage } from 'naive-ui';
import VueScrollTo from 'vue-scrollto'
import UploadSupport from './upload.vue'
import { liveQuery } from "dexie";
import { useObservable } from "@vueuse/rxjs";
import { db, getSrc, type Thumb } from '@/models';
import { throttle } from 'lodash-es';

const uploadImages = useObservable<Thumb[]>(
  liveQuery(() => db.thumbs.toArray()) as any
)

const message = useMessage()
const dialog = useDialog()

const virtualListRef = ref<InstanceType<typeof VirtualList> | null>(null);
const uploadSupportRef = ref<any>(null);
const messagesListRef = ref<HTMLElement | null>(null);
const textInputRef = ref<any>(null);

const rows = ref<Message[]>([]);

const textToSend = ref('');
const send = () => {
  if (chat.connectState !== 'connected') {
    message.error('尚未连接，请稍等');
    return;
  }
  const t = textToSend.value;
  if (t.trim() === '') {
    message.error('不能发送空消息');
    return;
  }
  if (t.length > 10000) {
    message.error('消息过长，请分段发送');
    return;
  }
  chat.messageCreate(t);

  textToSend.value = '';
  scrollToBottom();
}

const toBottom = () => {
  scrollToBottom();
  showButton.value = false;
}

const chat = useChatStore();
const user = useUserStore();

const doUpload = () => {
  uploadSupportRef.value.openUpload();
}

const isMe = (item: Message) => {
  return user.info.id === item.user?.id;
}

const scrollToBottom = () => {
  // virtualListRef.value?.scrollToBottom();
  nextTick(() => {
    const elLst = messagesListRef.value;
    if (elLst) {
      elLst.scrollTop = elLst.scrollHeight;
    }
  });
}

onMounted(async () => {
  await chat.tryInit();
  chatEvent.off('message-created', '*');
  chatEvent.on('message-created', (e?: Event) => {
    if (e && e.message && e.channel?.id == chat.curChannel?.id) {
      rows.value.push(e.message);
      if (!showButton.value) {
        scrollToBottom();
      }
    }
  });

  chatEvent.off('channel-deleted', '*');
  chatEvent.on('channel-deleted', (e) => {
    if (!e) {
      // 当前频道没了，直接进行重载
      channelSelect(chat.channelTree[0].id);
    }
  })

  loadMessages();
})

const messagesNextFlag = ref("");

const loadMessages = async () => {
  const messages = await chat.messageList(chat.curChannel?.id || '');
  messagesNextFlag.value = messages.next || "";
  rows.value.push(...messages.data);
  // for (let i = 0; i < 5000; i++) {
  //   rows.value.push({
  //     id: `x${i}`,
  //     timestamp: 123,
  //     member: {
  //       nick: '海豹',
  //       avatar: 'https://avatars.githubusercontent.com/u/12621342?v=4'
  //     },
  //     content: '已经就绪' + Math.random() + "||||    " + i,
  //   });
  // }

  nextTick(() => {
    scrollToBottom();
    showButton.value = false;
  })
}

const showModal = ref(false);
const newChannelName = ref('');
const newChannel = async () => {
  if (!newChannelName.value.trim()) {
    message.error('频道名不能为空');
    return;
  }
  await chat.channelCreate(newChannelName.value);
  await chat.channelList();
}

const showButton = ref(false)
const onScroll = (evt: any) => {
  // 会打断输入，不要blur
  // if (textInputRef.value?.blur) {
  //   (textInputRef.value as any).blur()
  // }
  // console.log(222, messagesListRef.value?.scrollTop, messagesListRef.value?.scrollHeight)
  if (messagesListRef.value) {
    const elLst = messagesListRef.value;
    const offset = elLst.scrollHeight - (elLst.clientHeight + elLst.scrollTop);
    showButton.value = offset > 200;

    if (elLst.scrollTop === 0) {
      reachTop(evt);
    }
  }
  // const vl = virtualListRef.value;
  // showButton.value = vl.clientRef.itemRefEl.clientHeight - vl.getOffset() > vl.clientRef.itemRefEl.clientHeight / 2
}

const keyUp = function (e: KeyboardEvent) {
  if (e.key === 'Enter' && (!e.ctrlKey) && (!e.shiftKey)) {
    send();
    e.preventDefault();

    // if (textInputRef.value?.blur) {
    //   (textInputRef.value as any).blur()
    // }
  }
}

let reachTopLoading = false;
const reachTop = async (evt: any) => {
  if (reachTopLoading) return;
  console.log('reachTop', messagesNextFlag.value)
  if (messagesNextFlag.value) {
    reachTopLoading = true;
    const messages = await chat.messageList(chat.curChannel?.id || '', messagesNextFlag.value);
    messagesNextFlag.value = messages.next || "";
    reachTopLoading = false;

    let oldId = '';
    if (rows.value.length) {
      oldId = rows.value[0].id || '';
    }

    rows.value.unshift(...messages.data);

    nextTick(() => {
      // 注意: el会变，如果不在下一帧取的话
      const el = document.getElementById(oldId)
      VueScrollTo.scrollTo(el, {
        container: messagesListRef.value,
        duration: 0,
        offset: 0,
      })
    })
    // virtualListRef.value?.scrollToIndex(messages.data.length);
  }
}

const renderIcon = (icon: Component) => {
  return () => {
    return h(NIcon, null, {
      default: () => h(icon)
    })
  }
}

const chOptions = computed(() => {
  const lst = chat.channelTree.map(i => {
    return {
      label: `${i.name} (${(i as any).membersCount})`,
      key: i.id,
      icon: undefined as any,
      props: undefined as any,
    }
  })
  lst.push({ label: '新建', key: 'new', icon: renderIcon(Plus), props: { style: { 'font-weight': 'bold' } } })
  return lst;
})

const channelSelect = async (key: string) => {
  if (key === 'new') {
    showModal.value = true;
    // chat.channelCreate('测试频道');
    // message.info('暂不支持新建频道');
  } else {
    await chat.channelSwitchTo(key);
    rows.value = []
    showButton.value = false;
    // 具体不知道原因，但是必须在这个位置reset才行
    virtualListRef.value?.reset();
    loadMessages();
  }
}

const sendEmoji = throttle((i: Thumb) => {
  chat.messageCreate(`<img src="id:${i.id}" />`)
}, 1000)
</script>

<template>
  <main class=" h-screen">
    <n-layout-header style="height: 4rem; padding: 24px" bordered>
      <div class="flex justify-between items-center">
        <div>
          <span class="text-sm font-bold sm:text-xl">海豹尬聊</span>
          <!-- <n-button>登录</n-button>
      <n-button>切换房间</n-button> -->
          <span class="ml-4">
            <n-dropdown trigger="click" :options="chOptions" @select="channelSelect">
              <!-- <n-button>{{ chat.curChannel?.name || '加载中 ...' }}</n-button> -->
              <n-button text>{{ chat.curChannel?.name ? `${chat.curChannel?.name} (${(chat.curChannel as
                any).membersCount})`
                : '加载中 ...' }} ▼</n-button>
            </n-dropdown>
          </span>

        </div>
        <div class="space-x-8">
          <!-- ● -->
          <span v-if="chat.connectState === 'connecting'" class=" text-blue-500">连接中 ...</span>
          <span v-if="chat.connectState === 'connected'" class=" text-green-600">已连接</span>
          <span v-if="chat.connectState === 'disconnected'" class=" text-red-500">已断开</span>
          <span v-if="chat.connectState === 'reconnecting'" class=" text-orange-400">{{ chat.iReconnectAfterTime }}s
            后自动重连</span>
          <!-- 这个其实有问题，应该是群内昵称 -->
          <span>{{ user.info.nick }}</span>
        </div>
      </div>
    </n-layout-header>

    <n-layout has-sider position="absolute" style="margin-top: 4rem;">
      <!-- <n-layout-sider :collapsed="false" content-style="padding: 24px;" :native-scrollbar="false" bordered>
        <n-h2 v-for="i in channelList">{{ i.name }}</n-h2>
      </n-layout-sider> -->

      <n-layout>
        <div class="flex flex-col h-full justify-between">
          <div class="chat overflow-y-auto h-full px-4 pt-6" v-show="rows.length > 0" @scroll="onScroll"
            ref="messagesListRef">
            <!-- <VirtualList itemKey="id" :list="rows" :minSize="50" ref="virtualListRef" @scroll="onScroll"
              @toBottom="reachBottom" @toTop="reachTop"> -->
            <template v-for="itemData in rows" :key="itemData.id">
              <chat-item :avatar="imgAvatar" :username="itemData.member?.nick" :content="itemData.content"
                :is-rtl="isMe(itemData)" :item="itemData" />
            </template>

            <!-- <VirtualList itemKey="id" :list="rows" :minSize="50" ref="virtualListRef" @scroll="onScroll"
              @toBottom="reachBottom" @toTop="reachTop">
              <template #default="{ itemData }">
                <chat-item :avatar="imgAvatar" :username="itemData.member?.nick" :content="itemData.content"
                  :is-rtl="isMe(itemData)" :createdAt="itemData.createdAt" />
              </template>
            </VirtualList> -->
          </div>
          <div v-if="rows.length === 0" class="flex h-full items-center justify-center text-gray-400">说点什么吧</div>

          <div style="right: 20px ;bottom: 70px;" class=" fixed" v-if="showButton">
            <n-button size="large" circle color="#e5e7eb" @click="toBottom">
              <template #icon>
                <n-icon class="text-black">
                  <ArrowBarToDown />
                </n-icon>
              </template>
            </n-button>
          </div>

          <!-- flex-grow -->
          <div class=" edit-area flex justify-between space-x-2 my-2 px-2">
            <n-input type="textarea" :rows="1" autosize v-model:value="textToSend" :on-keydown="keyUp" ref="textInputRef">
              <template #prefix>
                <n-popover trigger="click">
                  <template #trigger>
                    <n-button text>
                      <template #icon>
                        <n-icon :component="Plus" size="20" />
                      </template>
                    </n-button>
                  </template>
                  <div class=" text-base">表情(仅当前设备)</div>
                  <div class="grid grid-cols-4 gap-4">
                    <div v-for="i in uploadImages"  @click="sendEmoji(i)">
                      <img :src="getSrc(i)" style="width: 4.8rem; height: 4.8rem; object-fit: contain;" />
                    </div>
                  </div>
                </n-popover>
              </template>

              <template #suffix>
                <n-space>
                  <n-popover trigger="hover">
                    <template #trigger>
                      <n-button text @click="doUpload">
                        <template #icon>
                          <n-icon :component="Upload" size="20" />
                        </template>
                      </n-button>
                    </template>
                    <span>上传图片</span>
                  </n-popover>

                </n-space>
              </template>
            </n-input>
            <div class="flex" style="align-items: end; padding-bottom: 1px;">
              <n-button class="" type="primary" @click="send" :disabled="chat.connectState !== 'connected'">发送</n-button>
            </div>
          </div>
        </div>
      </n-layout>

    </n-layout>

    <n-modal v-model:show="showModal" preset="dialog" title="添加频道" content="你确认?" positive-text="确认" negative-text="算了"
      @positive-click="newChannel">
      <n-input v-model:value="newChannelName"></n-input>
    </n-modal>

    <upload-support ref="uploadSupportRef" />
  </main>
</template>

<style lang="scss" scoped>
.chat>.virtual-list__client {
  @apply px-4 pt-4;

  &>div {
    margin-bottom: -1rem;
  }
}

.chat-item {
  @apply pb-8; // margin会抖动，pb不会
}
</style>

<style lang="scss">
.chat>.virtual-list__client {
  &>div {
    margin-bottom: -1rem;
  }
}
</style>
