<script setup lang="tsx">
import dayjs from 'dayjs';
import imgAvatar from '@/assets/head2.png'
import ChatItem from '@/components/chat-item.vue';
import { computed, ref, watch, h } from 'vue'
import VGrid, { VGridVueTemplate } from "@revolist/vue-datagrid";

// const myVue = Vue.component("my-component", {
//   props: ["rowIndex", "model"],
//   computed: {
//     count() {
//       return this.model.count || 0;
//     },
//   },
//   methods: {
//     iAmClicked(e) {
//       Vue.set(this.model, "count", this.count + 1);
//     },
//   },
//   template:
//     '<button v-on:click="iAmClicked">You clicked me {{ count }} times.</button>',
// });

// const columns = ref([
//   {
//     name: "Vue",
//     size: 200,
//     cellTemplate: VGridVueTemplate(<div>xxxx</div>),
//   },
//   {
//     prop: "details",
//     autoSize: true,
//   },
// ])
import { VirtualList } from 'vue-tiny-virtual-list';
import type { TalkMessage } from '@/types';

const virtualListRef = ref<InstanceType<typeof VirtualList> | null>(null);

const rows = ref<TalkMessage[]>([
  {
    id: '1',
    time: 123,
    name: '海豹',
    content: '已经就绪',
    isMe: false
  },
  {
    id: '2',
    time: 123,
    name: '零冲',
    content: '你好',
    isMe: true
  },
  {
    id: '3',
    time: 123,
    name: '零冲',
    content: '.r d20',
    isMe: true
  },
  {
    id: '4',
    time: 123,
    name: '海豹',
    content: '<零冲>掷出了 D20=17',
    isMe: false
  },
  {
    id: '5',
    time: 123,
    name: '海豹',
    content: '<零冲>掷出了 D20=17',
    isMe: false
  },
  {
    id: '6',
    time: 123,
    name: '海豹',
    content: '<零冲>掷出了 D20=17',
    isMe: false
  },
  {
    id: '7',
    time: 123,
    name: '海豹',
    content: '<零冲>掷出了 D20=17',
    isMe: false
  },
  {
    id: '8',
    time: 123,
    name: '海豹',
    content: '<零冲>掷出了 D20=17',
    isMe: false
  },
])

for (let i = 0; i < 100; i++) {
  rows.value.push({
    id: `x${i}`,
    time: 123,
    name: '海豹',
    content: '测试' + Math.random(),
    isMe: false
  })
}

const textToSend = ref('');
const send = () => {
  const t = textToSend.value;
  console.log(222, textToSend.value);
  rows.value.push({
    id: Math.random().toString(),
    time: Date.now(),
    name: '零冲',
    content: textToSend.value,
    isMe: true
  });
  textToSend.value = '';
  virtualListRef.value?.scrollToBottom();
}
</script>

<template>
  <main class=" h-screen">
    <n-layout-header style="height: 4rem; padding: 24px" bordered>
      <span class="text-xl font-bold">海豹尬聊</span>
    </n-layout-header>

    <n-layout has-sider position="absolute" style="margin-top: 4rem;">
      <n-layout-sider :collapsed="false" content-style="padding: 24px;" :native-scrollbar="false" bordered>
        <n-h2>左侧</n-h2>
        <n-h2>左侧</n-h2>
        <n-h2>左侧</n-h2>
        <n-h2>左侧</n-h2>
        <n-h2>左侧</n-h2>
        <n-h2>左侧</n-h2>
        <n-h2>左侧</n-h2>
      </n-layout-sider>

      <n-layout>
        <div class="flex flex-col h-full justify-between">
          <div class="chat overflow-y-auto">
            <VirtualList itemKey="id" :list="rows" :minSize="20" ref="virtualListRef">
              <template #default="{ itemData }">
                <chat-item :avatar="imgAvatar" :username="itemData.name" :content="itemData.content"
                  :is-rtl="itemData.isMe" />
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
