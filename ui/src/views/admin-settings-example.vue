<script setup lang="tsx">
import { useUtilsStore } from '@/stores/utils';
import type { ServerConfig } from '@/types';
import { Message } from '@vicons/tabler';
import { cloneDeep } from 'lodash-es';
import { useMessage } from 'naive-ui';
import { computed, nextTick } from 'vue';
import { onMounted, ref, watch } from 'vue';

const tokens = ref([
  { name: '海豹', value: 'KHhD0rCfVnXVQEBybZIBm5FND10s0EQE', expireAt: 123 }
])

const emit = defineEmits(['close']);

const close = () => {
  emit('close');
}

const utils = useUtilsStore();
const message = useMessage()

const showModal = ref(false);
const newTokenName = ref('')
// const newChannel = async () => {
//   if (!newChannelName.value.trim()) {
//     message.error(t('dialoChannelgNew.channelNameHint'));
//     return;
//   }
//   await chat.channelCreate(newChannelName.value);
//   await chat.channelList();
// }

const addToken = async () => {

}

onMounted(async () => {
  const resp = await utils.configGet();
})
</script>

<template>
  <div class="overflow-y-auto pr-2" style="max-height: 61vh;  margin-top: 0;">
    <n-list>
      <template #header>
        当前token列表
      </template>

      <n-list-item v-for="i in tokens">
        <template #suffix>
          <div class="flex items-center space-x-2">
            <div style="width: 9rem;">
              <span>到期时间</span>
              <n-date-picker v-model:value="i.expireAt" type="date" />
            </div>
            <div>
              <span>操作</span>
              <n-button>删除</n-button>
            </div>
          </div>
        </template>
        <n-thing :title="i.name" :description="i.value"></n-thing>
      </n-list-item>

      <template #footer>
        <n-button @click="showModal = true">添加</n-button>
      </template>
    </n-list>
  </div>
  <div class="space-x-2 float-right">
    <n-button @click="close">关闭</n-button>
    <!-- <n-button type="primary" :disabled="!modified" @click="save">保存</n-button> -->
  </div>
</template>
