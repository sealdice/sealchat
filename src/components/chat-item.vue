<script setup lang="ts">
import dayjs from 'dayjs';
import imgAvatar from '@/assets/head2.png'

function timeFormat(time: string) {
  return dayjs(time).format('MM-DD HH:mm:ss');
}

const props = defineProps({
  username: String,
  content: String,
  avatar: String,
  isRtl: Boolean,
})
</script>

<template>
  <div class="chat-item" :style="props.isRtl ? { direction: 'rtl' } : {}" :class="props.isRtl ? ['is-rtl'] : []">
    <img class="rounded-md w-12 h-12 border-gray-500 border" :src="props.avatar" />
    <!-- <n-avatar :src="imgAvatar" size="large" bordered>海豹</n-avatar> -->
    <div class="right">
      <span class="title">
        <span v-if="!props.isRtl" class="name">{{ props.username }}</span>
        <span class="time">{{ timeFormat(new Date().toString()) }}</span>
      </span>
      <div class="content">
        {{ props.content }}
      </div>
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
