<script setup lang="tsx">
import { useChatStore } from '@/stores/chat';
import { useUtilsStore } from '@/stores/utils';
import type { ServerConfig, UserInfo } from '@/types';
import { Message, Refresh } from '@vicons/tabler';
import { cloneDeep } from 'lodash-es';
import { useDialog, useMessage } from 'naive-ui';
import { computed, nextTick } from 'vue';
import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const tokens = ref([
  { name: '海豹', value: 'KHhD0rCfVnXVQEBybZIBm5FND10s0EQE', expireAt: 123 }
])

const emit = defineEmits(['close']);

const close = () => {
  emit('close');
}

const chat = useChatStore();
const utils = useUtilsStore();
const message = useMessage()

const { t } = useI18n()

const data = ref([]);
// const data = ref({
//   total: 0,
//   items: [] as UserInfo[]
// })
onMounted(async () => {
  refresh()
})

const refresh = async () => {
  const resp = await utils.adminUserList();
  data.value = resp.data.items;
}

const dialog = useDialog()

const tryUserResetPassword = (i: UserInfo) => {
  dialog.warning({
    title: t('dialogLogOut.title'),
    content: '重置此用户密码为123456吗？',
    positiveText: t('dialogLogOut.positiveText'),
    negativeText: t('dialogLogOut.negativeText'),
    onPositiveClick: async () => {
      try {
        await utils.userResetPassword(i.id);
        message.success('重置成功');
      } catch (error) {
        message.error('重置失败: ' + (error as any).response?.data?.message || '未知错误');
      }
    },
    onNegativeClick: () => {
    }
  })
}

const tryUserDisable = (i: UserInfo) => {
  dialog.warning({
    title: t('dialogLogOut.title'),
    content: '确定要禁用此帐号吗？',
    positiveText: t('dialogLogOut.positiveText'),
    negativeText: t('dialogLogOut.negativeText'),
    onPositiveClick: async () => {
      try {
        await utils.userDisable(i.id);
        message.success('停用成功');
        refresh();
      } catch (error) {
        message.error('停用失败: ' + (error as any).response?.data?.message || '未知错误');
      }
    },
    onNegativeClick: () => {
    }
  })
}

const tryUserEnable = (i: UserInfo) => {
  // 偷懒复制了
  dialog.warning({
    title: t('dialogLogOut.title'),
    content: '确定要启用此帐号吗？',
    positiveText: t('dialogLogOut.positiveText'),
    negativeText: t('dialogLogOut.negativeText'),
    onPositiveClick: async () => {
      try {
        await utils.userEnable(i.id);
        message.success('启用成功');
        refresh();
      } catch (error) {
        message.error('启用失败: ' + (error as any).response?.data?.message || '未知错误');
      }
    },
    onNegativeClick: () => {
    }
  })
}

const handleRoleChange = async (userId: string, roleLst: string[], oldRoleLst: string[]) => {
  // 计算需要移除和添加的成员
  const toRemove = oldRoleLst.filter(id => !roleLst.includes(id));
  const toAdd = roleLst.filter(id => !oldRoleLst.includes(id));
  console.log('toAdd', toAdd);
  console.log('toRemove', toRemove);

  try {
    if (toAdd.length) await utils.userRoleLinkByUserId(userId, toAdd);
    if (toRemove.length) await utils.userRoleUnlinkByUserId(userId, toRemove);
    refresh();
    message.success('角色已成功添加');
  } catch (error) {
    console.error('添加角色失败:', error);
    message.error('添加角色失败，请重试');
  }
};

const columns = ref([
  // {
  //   title: 'id',
  //   key: 'id'
  // },
  {
    title: 'username',
    key: 'username'
  },
  {
    title: 'nickname',
    key: 'nick'
  },
  {
    title: 'role',
    key: 'role',
    render: (row: UserInfo) => {
      return (
        // { label: '游客', value: 'sys-visitor' }
        <n-select
          v-model:value={row.roleIds}
          multiple
          options={[
            { label: '管理员', value: 'sys-admin' },
            { label: '普通用户', value: 'sys-user' },
          ]}
          size="small"
          on-update:value={(value: any) => handleRoleChange(row.id, value, row.roleIds ?? [])}
        />
      )
    },
  },
  {
    title: 'actions',
    render: (row: UserInfo) => {
      const isDisabled = row.disabled;
      return <div class="flex space-x-2">
        <n-button type="warning" size="small" onClick={() => tryUserResetPassword(row)}>重置密码</n-button>
        {!isDisabled ? <n-button type="error" size="small" onClick={() => tryUserDisable(row)}>停用</n-button> :
          <n-button type="success" size="small" onClick={() => tryUserEnable(row)}>启用</n-button>}
      </div>
    }
  }
]);
const pagination = ref(false);
</script>

<template>
  <div class="overflow-y-auto pr-2" style="max-height: 61vh;  margin-top: 0;">
    <n-data-table :columns="columns" :data="data" :pagination="pagination" :bordered="false" />
  </div>
  <div class="space-x-2 float-right mt-4">
    <n-button @click="close">关闭</n-button>
    <!-- <n-button type="primary" :disabled="!modified" @click="save">保存</n-button> -->
  </div>
</template>
