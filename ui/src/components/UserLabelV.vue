<script setup lang="tsx">
import Avatar from './avatar.vue'

const props = defineProps<{
  src?: string,
  userId?: string,
  name?: string
}>()
</script>

<template>
  <span class="flex flex-col items-center ">
    <Avatar :size="48" :src="src"></Avatar>
    <span class=" text-more" style="max-width: 5rem" >{{ name || '未知' }}</span>
    <slot name="suffix"></slot>
  </span>
</template>
