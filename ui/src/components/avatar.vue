<script setup lang="tsx">
import imgAvatar from '@/assets/head3.png'
import { urlBase } from '@/stores/_config';
import { useUserStore } from '@/stores/user';
import { computed, onMounted, ref } from 'vue';

const props = defineProps({
  src: String,
  size: {
    type: Number,
    default: 48,
  },
  border: {
    type: Boolean,
    default: true,
  },
})

const opacity = ref(1)
const onload = function () {
  opacity.value = 0;
}

onMounted(() => {
})

const src = computed(() => {
  if (props.src?.startsWith('id:')) {
    return props.src.replace('id:', `${urlBase}/api/v1/attachments/`);
  }
  // console.log('src', props.src)
})
</script>

<template>
  <div class="rounded-md w-12 h-12 border-gray-300 relative overflow-clip"
    :style="{ width: `${size}px`, height: `${size}px`, 'min-width': `${size}px`, 'min-height': `${size}px` }" :class="border ? ['border'] : []">
    <img class="w-full h-full" :src="src" :onload="onload" />
    <img class="absolute w-full h-full" :src="imgAvatar" style="top:0" :style="{ opacity: opacity }" />
  </div>
</template>
