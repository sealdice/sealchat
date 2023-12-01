<script setup lang="tsx">
import { useUtilsStore } from '@/stores/utils';
import type { ServerConfig } from '@/types';
import { Message, Refresh } from '@vicons/tabler';
import { cloneDeep } from 'lodash-es';
import { useDialog, useMessage } from 'naive-ui';
import { computed, nextTick } from 'vue';
import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n()

const emit = defineEmits(['close']);

const cancel = () => {
  emit('close');
}

const showModal = ref(false);
const newTokenName = ref('bot')
// const newChannel = async () => {
//   if (!newChannelName.value.trim()) {
//     message.error(t('dialoChannelgNew.channelNameHint'));
//     return;
//   }
//   await chat.channelCreate(newChannelName.value);
//   await chat.channelList();
// }

const addToken = async () => {
  try {
    await utils.botTokenAdd(newTokenName.value);
    message.success('添加成功');
    refresh();
  } catch (error) {
    message.error('添加失败: ' + (error as any).response?.data?.message || '未知错误');
  }

  newTokenName.value = 'bot'
  // tokens.value.push({
  //   name: newTokenName.value,
  //   value: 'KHhD0rCfVnXVQEBybZIBm5FND10s0EQE',
  //   expireAt: 123
  // })
}

// const tokens = ref([
//   { name: '海豹', value: 'KHhD0rCfVnXVQEBybZIBm5FND10s0EQE', expireAt: 123 }
// ])
const tokens = ref({
  total: 0,
  items: [] as any[]
})

const utils = useUtilsStore();
const message = useMessage()
const dialog = useDialog()

const refresh = async () => {
  const resp = await utils.botTokenList();
  tokens.value = resp.data;
  console.log(222, resp.data)
}

const deleteItem = async (i: any) => {
  dialog.warning({
    title: t('dialogLogOut.title'),
    content: '确定要删除吗？',
    positiveText: t('dialogLogOut.positiveText'),
    negativeText: t('dialogLogOut.negativeText'),
    onPositiveClick: async () => {
      try {
        await utils.botTokenDelete(i.id);
        message.success('删除成功');
        refresh();
      } catch (error) {
        message.error('删除失败: ' + (error as any).response?.data?.message || '未知错误');
      }
    },
    onNegativeClick: () => {
    }
  })
}

onMounted(async () => {
  refresh()
})
</script>

<template>
  <div class="overflow-y-auto pr-2" style="max-height: 61vh;  margin-top: 0;">
    <n-list>
      <template #header>
        当前token列表
      </template>

      <n-list-item v-for="i in tokens.items">
        <template #suffix>
          <div class="flex items-center space-x-2">
            <div style="width: 9rem;">
              <span>到期时间</span>
              <n-date-picker v-model:value="i.expiresAt" type="date" />
              <!-- <div v-else>无期限</div> -->
            </div>
            <div>
              <span>操作</span>
              <n-button @click="deleteItem(i)">删除</n-button>
            </div>
          </div>
        </template>
        <n-thing :title="i.name" :description="i.token"></n-thing>
      </n-list-item>

      <template #footer>
        <n-button @click="showModal = true">添加</n-button>
      </template>
    </n-list>
  </div>
  <div class="space-x-2 float-right">
    <n-button @click="cancel">关闭</n-button>
    <!-- <n-button type="primary" :disabled="!modified" @click="save">保存</n-button> -->
  </div>
  <n-modal v-model:show="showModal" preset="dialog" :title="'起个名字'" :positive-text="$t('dialoChannelgNew.positiveText')"
    :negative-text="$t('dialoChannelgNew.negativeText')" @positive-click="addToken">
    <n-input v-model:value="newTokenName"></n-input>
  </n-modal>
</template>
