<script lang="tsx" setup>
import type { UserInfo, UserRoleModel } from '@/types';
import { X, Check } from '@vicons/tabler'
import { ref, defineProps } from 'vue';

const props = defineProps({
  memberList: {
    type: Array<UserInfo>,
  },
  startSelectedList: {
    type: Array<UserInfo>,
  }
});


const selectedList = ref<string[]>([]);

for (let i of props.startSelectedList || []) {
  if (i.id) {
    selectedList.value.push(i.id);
  }
}

const toggleSelection = (userId?: string) => {
  if (!userId) return;
  const index = selectedList.value.indexOf(userId);
  if (index === -1) {
    selectedList.value.push(userId);
  } else {
    selectedList.value.splice(index, 1);
  }
};

const isSelected = (userId?: string) => {
  if (!userId) return;
  return selectedList.value.includes(userId);
};

const emit = defineEmits(['confirm']);

const handleConfirm = () => {
  emit('confirm', selectedList.value, props.startSelectedList?.map(i => i.id) ?? []);
};
</script>

<template>
  <div class="flex flex-wrap justify-center relative">
    <div v-if="props.memberList" class="relative group pr-1 select-none" v-for="j in props.memberList"
      @click="toggleSelection(j.id)">
      <UserLabelV :name="j.nick ?? j.username" :src="j.avatar" />

      <div class="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center " v-if="isSelected(j.id)">
        <n-icon size="24" color="#ffffff">
          <Check />
        </n-icon>
      </div>
    </div>
  </div>


  <div class="flex justify-end mt-2">
    <n-button class="mt-4 w-full" type="primary" @click="handleConfirm">
      确定
    </n-button>
  </div>

</template>
