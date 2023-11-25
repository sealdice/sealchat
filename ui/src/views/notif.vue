<script lang="tsx" setup>
import { useUserStore } from '@/stores/user';
import { onMounted, ref } from 'vue';

const user = useUserStore();

const x = ref<any>([]);

onMounted(async () => {
  const resp = await user.timelineList()
  x.value = resp.data.items;
})
</script>

<template>
  <div class="absolute justify-center items-center flex w-full h-full pointer-events-none z-10">
    <div class="border pointer-events-auto">
      <div v-for="i in x" class="p-4 bg-white">
        行为: {{ i.type }}
        发送者: {{ i.senderId }}
        地点: {{ i.locPostType }} {{ i.locPostId }}
        已读: {{ i.isRead }}
        时间: {{ i.createdAt }}
      </div>
    </div>
  </div>
</template>
