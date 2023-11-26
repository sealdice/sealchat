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
import { NIcon, c, useDialog, useMessage, type MentionOption } from 'naive-ui';
import VueScrollTo from 'vue-scrollto'
import UploadSupport from './upload.vue'
import { liveQuery } from "dexie";
import { useObservable } from "@vueuse/rxjs";
import { db, getSrc, type Thumb } from '@/models';
import { throttle } from 'lodash-es';
import ChatHeader from './header.vue'
import AvatarVue from '@/components/avatar.vue';
import { Howl, Howler } from 'howler';
import SoundMessageCreated from '@/assets/message.mp3';

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

async function replaceUsernames(text: string) {
  const resp = await chat.guildMemberList('');
  const infoMap = (resp.data as any[]).reduce((obj, item) => {
    obj[item.nick] = item;
    return obj;
  }, {})

  // 匹配 @ 后跟着字母数字下划线的用户名
  const regex = /@(\S+)/g;

  // 使用 replace 方法来替换匹配到的用户名
  const replacedText = text.replace(regex, (match, username) => {
    if (username in infoMap) {
      const info = infoMap[username];
      return `<at id="${info.id}" name="${info.nick}" />`
    }
    return match;
  });

  return replacedText;
}

const textToSend = ref('');
const send = async () => {
  if (chat.connectState !== 'connected') {
    message.error('尚未连接，请稍等');
    return;
  }
  let t = textToSend.value;
  if (t.trim() === '') {
    message.error('不能发送空消息');
    return;
  }
  if (t.length > 10000) {
    message.error('消息过长，请分段发送');
    return;
  }
  t = await replaceUsernames(t)
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
  const elInput = textInputRef.value;
  if (elInput) {
    // 注: n-mention 不支持这个事件监听，所以这里手动监听
    elInput.$el.getElementsByTagName('textarea')[0].onkeydown = keyDown;
  }

  var sound = new Howl({
    src: [SoundMessageCreated],
    html5: true
  });

  chatEvent.off('message-created', '*');
  chatEvent.on('message-created', (e?: Event) => {
    console.log('???', e)
    if (e && e.message && e.channel?.id == chat.curChannel?.id) {
      if (e.message.user?.id !== user.info.id) {
        // 不是自己发的消息，播放声音
        sound.play();
      }
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
      chat.channelSwitchTo(chat.channelTree[0].id);
    }
  })

  chatEvent.on('channel-switch-to', (e) => {
    rows.value = []
    showButton.value = false;
    // 具体不知道原因，但是必须在这个位置reset才行
    // virtualListRef.value?.reset();
    loadMessages();
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

const pauseKeydown = ref(false);
const keyDown = function (e: KeyboardEvent) {
  if (pauseKeydown.value) return;
  if (e.key === 'Enter' && (!e.ctrlKey) && (!e.shiftKey)) {
    send();
    e.preventDefault();

    // if (textInputRef.value?.blur) {
    //   (textInputRef.value as any).blur()
    // }
  }
}

const atOptions = ref<MentionOption[]>([])
const atLoading = ref(true)
const atRenderLabel = (option: MentionOption) => {
  return <div class="flex items-center space-x-1">
    <AvatarVue size={24} border={false} src={(option as any).data?.avatar} />
    <span>{option.label}</span>
  </div>
}

const atHandleSearch = async (pattern: string, prefix: string) => {
  pauseKeydown.value = true;
  atLoading.value = true;

  const atElementCheck = () => {
    const els = document.getElementsByClassName("v-binder-follower-content");
    if (els.length) {
      return els[0].children.length > 0;
    }
    return false;
  }

  // 如果at框非正常消失，那么也一样要恢复回车键功能
  let x = setInterval(() => {
    if (!atElementCheck()) {
      pauseKeydown.value = false;
      clearInterval(x);
    }
  }, 100)

  const lst = (await chat.guildMemberList('')).data.map((i: any) => {
    return {
      value: i.nick,
      label: i.nick,
      data: i,
    }
  })
  atOptions.value = lst;
  atLoading.value = false;
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

const sendEmoji = throttle((i: Thumb) => {
  chat.messageCreate(`<img src="id:${i.id}" />`)
}, 1000)

</script>

<template>
  <main class=" h-screen">
    <n-layout-header style="height: 4rem; padding: 24px" bordered>
      <chat-header />
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
              <chat-item :avatar="itemData.member?.avatar || itemData.user?.avatar" :username="itemData.member?.nick"
                :content="itemData.content" :is-rtl="isMe(itemData)" :item="itemData" />
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
          <div class="edit-area flex justify-between space-x-2 my-2 px-2 relative">
            <div class="flex justify-between relative w-full">
              <!-- 输入框左侧按钮，因为n-mention不支持#prefix和#suffix，所以单独拿出来了 -->
              <div class="absolute" style="z-index: 1; left: 0.5rem; top: .55rem;">
                <n-popover trigger="click">
                  <template #trigger>
                    <n-button text>
                      <template #icon>
                        <n-icon :component="Plus" size="20" />
                      </template>
                    </n-button>
                  </template>
                  <div class="text-base">{{ $t('inputBox.emojiTitle') }}</div>
                  <div class="grid grid-cols-4 gap-4">
                    <div v-for="i in uploadImages" @click="sendEmoji(i)">
                      <img :src="getSrc(i)" style="width: 4.8rem; height: 4.8rem; object-fit: contain;" />
                    </div>
                  </div>
                </n-popover>
              </div>

              <div class="absolute" style="z-index: 1; right: 0.6rem; top: .55rem;">
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
              </div>

              <n-mention type="textarea" :rows="1" autosize v-model:value="textToSend" :on-keydown="keyDown"
                ref="textInputRef" class="chat-text" :placeholder="$t('inputBox.placeholder')" :options="atOptions" :loading="atLoading"
                @search="atHandleSearch" @select="pauseKeydown = false" :render-label="atRenderLabel">
              </n-mention>
            </div>
            <div class="flex" style="align-items: end; padding-bottom: 1px;">
              <n-button class="" type="primary" @click="send" :disabled="chat.connectState !== 'connected'">{{ $t('inputBox.send') }}</n-button>
            </div>
          </div>
        </div>
      </n-layout>
    </n-layout>

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

.chat-text>.n-input>.n-input-wrapper {
  padding-left: 2rem;
  padding-right: 2rem;
}
</style>
