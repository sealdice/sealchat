<script lang="tsx" setup>
import { Plus } from '@vicons/tabler'
</script>

<template>
  <n-button style="height: 3rem; width: 3rem;" class="flex items-center justify-center" quaternary circle>
    <template #icon>
      <n-icon size="24">
        <Plus />
      </n-icon>
    </template>
  </n-button>
</template>
