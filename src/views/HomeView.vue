<script setup lang="ts">
import dayjs from 'dayjs';
import imgAvatar from '@/assets/head2.png'

function timeFormat(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss');
}
</script>

<template>
  <main>
    <div style="padding: 10%">
      <div class="chat">
        <div class="chat-item" v-for="i in [1, 2, 3]">
          <img class="rounded-md w-12 h-12 border-gray-500 border" :src="imgAvatar" />
          <!-- <n-avatar :src="imgAvatar" size="large" bordered>海豹</n-avatar> -->
          <div class="right">
            <span class="title">
              <span class="name">伊莎玛拉</span>
              <span class="time">{{ timeFormat(new Date().toString()) }}</span>
            </span>
            <div class="content">
              {{ '<纯爱战士>掷出了 D20=17' }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<style lang="scss">
.chat {
  @apply space-y-8
}

.chat-item {
  @apply flex;

  >.n-avatar {
    @apply rounded-md;
  }

  >.right {
    margin-left: 0.5rem;
    @apply ml-4;

    >.title {
      display: flex;
      gap: 0.5rem;

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
  
        [dir="rtl"] & {
          mask-image: url("data:image/svg+xml,%3csvg width='3' height='3' xmlns='http://www.w3.org/2000/svg'%3e%3cpath fill='black' d='m 0 3 L 1 3 L 3 3 C 2 3 0 1 0 0'/%3e%3c/svg%3e");
        }
      }

      @apply text-base mt-1 px-4 py-2 relative;
      @apply rounded bg-gray-200 text-gray-900;
    }

  }
}
</style>
