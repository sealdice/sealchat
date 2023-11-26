<script setup lang="ts">
import dayjs from 'dayjs';
import imgAvatar from '@/assets/head2.png'
import Element from '@satorijs/element'
import { onMounted, ref } from 'vue';
import { urlBase } from '@/stores/_config';
import DOMPurify from 'dompurify';
import { useUserStore } from '@/stores/user';
import { useChatStore } from '@/stores/chat';
import type { Message } from '@satorijs/protocol';

const user = useUserStore();
const chat = useChatStore();

function timeFormat(time?: string) {
  if (!time) return '未知';
  // console.log('???', time, typeof time)
  // return dayjs(time).format('MM-DD HH:mm:ss');
  return dayjs(time).fromNow();
}

let hasImage = ref(false);

const parseContent = (props: any) => {
  const content = props.content;
  const items = Element.parse(content);
  let textItems = []

  for (const item of items) {
    switch (item.type) {
      case 'img':
        if (item.attrs.src && item.attrs.src.startsWith('id:')) {
          item.attrs.src = item.attrs.src.replace('id:', `${urlBase}/api/v1/attachments/`);
        }
        textItems.push(DOMPurify.sanitize(item.toString()));
        hasImage.value = true;
        break;
      case "at":
        if (item.attrs.id == user.info.id) {
          textItems.push(`<span class="text-blue-500 bg-gray-400 px-1" style="white-space: pre-wrap">@${item.attrs.name}</span>`);
        } else {
          textItems.push(`<span class="text-blue-500" style="white-space: pre-wrap">@${item.attrs.name}</span>`);
        }
      default:
        textItems.push(`<span style="white-space: pre-wrap">${item.toString()}</span>`);
        break;
    }
  }

  return textItems.join('');
}

const props = defineProps({
  username: String,
  content: String,
  avatar: String,
  isRtl: Boolean,
  key: String,
  item: Object,
})

const timeText = ref(timeFormat(props.item?.createdAt));

const onContextMenu = (e: MouseEvent, item: any) => {
  e.preventDefault();
  //Set the mouse position

  chat.messageMenu.optionsComponent.x = e.x;
  chat.messageMenu.optionsComponent.y = e.y;
  //Show menu
  chat.messageMenu.show = true;
  chat.messageMenu.item = item;
  chat.messageMenu.hasImage = hasImage.value;
}

onMounted(() => {
  setInterval(() => {
    timeText.value = timeFormat(props.item?.createdAt);
  }, 10000);
})
</script>

<template>
  <div :id="item?.id" class="chat-item" :style="props.isRtl ? { direction: 'rtl' } : {}"
    :class="props.isRtl ? ['is-rtl'] : []" :key="key">
    <avatar :src="props.avatar" />
    <!-- <img class="rounded-md w-12 h-12 border-gray-500 border" :src="props.avatar" /> -->
    <!-- <n-avatar :src="imgAvatar" size="large" bordered>海豹</n-avatar> -->
    <div class="right">
      <span class="title">
        <span v-if="!props.isRtl" class="name">{{ props.username }}</span>
        <span class="time">{{ timeText }}</span>
        <span v-if="props.item?.user?.is_bot || props.item?.user_id?.startsWith('BOT:')" class=" bg-blue-500 rounded-md px-2 text-white">bot</span>
      </span>
      <div class="content break-all" v-html="parseContent(props)" @contextmenu="onContextMenu($event, item)"></div>
    </div>
  </div>
</template>

<style lang="scss">
.chat-item {
  @apply flex;

  >.n-avatar {
    @apply rounded-md;
  }

  &.is-rtl {
    >.right {
      @apply mr-4;

      >.title {
        @apply justify-end;
      }

      >.content {
        &:before {
          display: none;
        }

        &::after {
          position: absolute;
          top: 0.5rem;
          height: 0.75rem;
          width: 0.75rem;
          background-color: inherit;
          content: "";
          right: -0.75rem;
          transform: scaleY(-1) scaleX(-1);
          mask-size: contain;
          mask-image: url("data:image/svg+xml,%3csvg width='3' height='3' xmlns='http://www.w3.org/2000/svg'%3e%3cpath fill='black' d='m 0 3 L 3 3 L 3 0 C 3 1 1 3 0 3'/%3e%3c/svg%3e");
        }
      }
    }
  }

  >.right {
    @apply ml-4;

    >.title {
      display: flex;
      gap: 0.5rem;
      direction: ltr;

      >.name {
        @apply font-semibold;
      }

      >.time {
        @apply text-gray-400;
      }
    }

    >.content {
      &:before {
        position: absolute;
        top: 0.5rem;
        height: 0.75rem;
        width: 0.75rem;
        background-color: inherit;
        content: "";
        left: -0.75rem;
        transform: scaleY(-1);
        mask-size: contain;
        mask-image: url("data:image/svg+xml,%3csvg width='3' height='3' xmlns='http://www.w3.org/2000/svg'%3e%3cpath fill='black' d='m 0 3 L 3 3 L 3 0 C 3 1 1 3 0 3'/%3e%3c/svg%3e");
      }

      width: fit-content;
      direction: ltr;
      @apply text-base mt-1 px-4 py-2 relative;
      @apply rounded bg-gray-200 text-gray-900;
    }
  }
}
</style>
