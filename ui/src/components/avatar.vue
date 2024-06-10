<script setup lang="tsx">
import imgAvatar from '@/assets/head3.png'
import { urlBase } from '@/stores/_config';
import { useUserStore } from '@/stores/user';
import { computed, onMounted, ref } from 'vue';
import { onLongPress } from '@vueuse/core'

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
  if (!props.src) {
    opacity.value = 1;
  }
})

const emit = defineEmits(['longpress']);

const htmlRefHook = ref<HTMLElement | null>(null)
const onLongPressCallbackHook = (e: PointerEvent) => {
  emit('longpress', e)
}

onLongPress(
  htmlRefHook,
  onLongPressCallbackHook,
  { modifiers: { prevent: true } }
)
</script>

<template>
  <div class="rounded-md w-12 h-12 border-gray-300 relative overflow-clip" ref="htmlRefHook"
    :style="{ width: `${size}px`, height: `${size}px`, 'min-width': `${size}px`, 'min-height': `${size}px` }" :class="border ? ['border'] : []">
    <img class="w-full h-full" :src="src" v-if="src" :onload="onload" />
    <img class="absolute w-full h-full" :class="{ 'pointer-events-none': opacity === 0 }" :src="imgAvatar" style="top:0" :style="{ opacity: opacity }" />
  </div>
</template>
