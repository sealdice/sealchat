<script setup lang="tsx">
import dayjs from 'dayjs';
import imgAvatar from '@/assets/head2.png'
import ChatItem from '@/components/chat-item.vue';
import { computed, ref, watch, h, onMounted, onBeforeMount } from 'vue'
import { VirtualList } from 'vue-tiny-virtual-list';
import type { TalkMessage } from '@/types';
import { chatEvent, useChatStore } from '@/stores/chat';
import type { Event, Message } from '@satorijs/protocol'
import { useUserStore } from '@/stores/user';

const virtualListRef = ref<InstanceType<typeof VirtualList> | null>(null);

const rows = ref<Message[]>([
  {
    id: '1',
    timestamp: 123,
    member: {
      nick: '海豹',
      avatar: 'https://avatars.githubusercontent.com/u/12621342?v=4'
    },
    content: '已经就绪',
  }
]);

const textToSend = ref('');
const send = () => {
  const t = textToSend.value;
  // console.log('XXX', t);
  chat.messageCreate(t);

  textToSend.value = '';
  virtualListRef.value?.scrollToBottom();
}

const chat = useChatStore();
const user = useUserStore();

const isMe = (item: Message) => {
  return user.info.id === item.user?.id;
}

const channelList = ref<any[]>([]);

onBeforeMount(async () => {
  await chat.reinit();
  const tree = await chat.channelList();
  channelList.value = tree;

  chatEvent.on('message-created', (e: Event) => {
    if (e.message) {
      rows.value.push(e.message);
      virtualListRef.value?.scrollToBottom();
    }
  });

  
  const messages = await chat.messageList(tree[0].id);
  console.log(222, messages.data);
  rows.value.push(...messages.data);
})

const newChannel = async () => {
  chat.channelCreate('测试频道');
}
</script>

<template>
  <main class=" h-screen">
    <n-layout-header style="height: 4rem; padding: 24px" bordered>
      <span class="text-xl font-bold">海豹尬聊</span>
      <!-- <n-button>登录</n-button>
      <n-button>切换房间</n-button> -->
      <n-button class="ml-4" @click="newChannel" style="display: none;">新建频道</n-button>
    </n-layout-header>

    <n-layout has-sider position="absolute" style="margin-top: 4rem;">
      <n-layout-sider :collapsed="false" content-style="padding: 24px;" :native-scrollbar="false" bordered>
        <n-h2 v-for="i in channelList">{{ i.name }}</n-h2>
      </n-layout-sider>

      <n-layout>
        <div class="flex flex-col h-full justify-between">
          <div class="chat overflow-y-auto">
            <VirtualList itemKey="id" :list="rows" :minSize="20" ref="virtualListRef">
              <template #default="{ itemData }">
                <chat-item :avatar="imgAvatar" :username="itemData.member?.nick" :content="itemData.content"
                  :is-rtl="isMe(itemData)" />
              </template>
            </VirtualList>
          </div>

          <!-- flex-grow -->
          <div class=" edit-area flex justify-between space-x-2 my-2 px-2">
            <n-input type="textarea" :rows="1" autosize v-model:value="textToSend"></n-input>
            <div class="flex" style="align-items: end; padding-bottom: 1px;">
              <n-button class="" type="primary" @click="send">发送</n-button>
            </div>
          </div>
        </div>
      </n-layout>

    </n-layout>

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
