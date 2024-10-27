<script lang="tsx" setup>
import { ChannelType, type SChannel } from '@/types';
import { clone } from 'lodash-es';
import { useDialog, useMessage } from 'naive-ui';
import { onMounted, ref, watch, type PropType } from 'vue';
import TabMember from './TabMember.vue'

const message = useMessage();
const dialog = useDialog();

const modelShow = defineModel({
  default: false,
  set(value: any): any {
    if (value) {
      if (props.channel) {
        model.value = clone(props.channel);
      }
    }
    return value;
  }
});

// watch(modelShow[1], () => {
//   console.log(111111);
//   alert(1111)
// })

const props = defineProps({
  channel: {
    type: Object as PropType<SChannel>,
  }
})

const model = ref<SChannel>({
  id: '',
  type: 0, // 0 text
  typingIndicatorSetting: false,
})

onMounted(() => {
  if (props.channel) {
    model.value = clone(props.channel);
  }
});

const channelEdit = () => {
  ;
}

const memberColumns = [
  {
    title: '选择',
    key: 'select',
    render: (row: any) => {
      return <n-checkbox v-model:checked={row.selected} />
    }
  },
  {
    title: '用户名',
    key: 'username'
  },
  {
    title: '昵称',
    key: 'nickname'
  },
  {
    title: '角色',
    key: 'role',
    render: (row: any) => {
      return (
        <div>
          {(row.role ?? []).join(', ')}
        </div>
      )
    }
  },
  // {
  //   title: '操作',
  //   key: 'actions',
  //   render: (row: any) => {
  //     return (
  //       <n-button
  //         size="small"
  //         onClick={() => handleRemoveMember(row)}
  //       >
  //         移除
  //       </n-button>
  //     )
  //   }
  // }
]

const memberData = ref([
  { id: 1, username: '用户1', role: ['管理员'] },
  { id: 2, username: '用户2', role: ['成员'] },
  { id: 3, username: '用户3', role: ['成员'] }
])

const handleRemoveMember = (member: string) => {
  // 实现移除成员的逻辑
  console.log('移除成员:', member)
}

const checkedRowKeys = ref<string[]>([])

const handleCheck = (keys: string[]) => {
  checkedRowKeys.value = keys
}

const handleRemoveSelected = () => {
  if (checkedRowKeys.value.length === 0) {
    message.warning('请先选择要移除的成员');
    return;
  }

  dialog.warning({
    title: '确认移除',
    content: `您确定要移除选中的 ${checkedRowKeys.value.length} 名成员吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: () => {
      // 实现批量移除成员的逻辑
      memberData.value = memberData.value.filter(member => !checkedRowKeys.value.includes(member.id.toString()));
      checkedRowKeys.value = [];
      message.success('成员已成功移除');
    }
  });
}

const permissionOptions = [
  {
    label: '管理员',
    key: 'admin'
  },
  {
    label: '普通成员',
    key: 'member'
  },
  {
    label: '只读成员',
    key: 'readonly'
  }
]

const handlePermissionSelect = (key: string) => {
  if (checkedRowKeys.value.length === 0) {
    message.warning('请先选择要设置权限的成员')
    return
  }

  // 这里可以实现设置权限的逻辑
  console.log('设置权限给选中的成员:', checkedRowKeys.value)
}


const onAdd = () => {
  ;
}
</script>

<template>
  <n-modal v-model:show="modelShow" preset="dialog" :title="'频道设置'" :positive-text="$t('dialoChannelgNew.positiveText')"
    :negative-text="$t('dialoChannelgNew.negativeText')" @positive-click="channelEdit"
    style="min-height: 70vh;" class="modalX">

    <n-tabs type="line" animated>
      <n-tab-pane name="members" tab="成员管理">
        <TabMember :channel="channel" />
      </n-tab-pane>

      <n-tab-pane name="basic" tab="基础设置">
        <div>尚未制作完成</div>
        <n-form label-placement="top" label-width="auto" class="pt-4">
          <n-form-item label="频道名称">
            <n-input v-model:value="model.name" />
          </n-form-item>
          <!-- 可以在这里添加更多基础设置项 -->
          <n-form-item label="频道简介">
            <n-input v-model:value="model.desc" type="textarea" :autosize="{ minRows: 3, maxRows: 5 }" />
          </n-form-item>
          <n-form-item label="频道类型">
            <n-radio-group v-model:value="model.permType">
              <n-radio :value="ChannelType.Public">公开</n-radio>
              <n-radio :value="ChannelType.NonPublic">非公开</n-radio>
            </n-radio-group>
          </n-form-item>

          <n-form-item label="成员'正在输入'提示">
            <n-radio-group v-model:value="model.typingIndicatorSetting">
              <n-radio :value="true">启用</n-radio>
              <n-radio :value="false">禁用</n-radio>
            </n-radio-group>
          </n-form-item>
        </n-form>
      </n-tab-pane>

      <n-tab-pane name="members2" tab="权限配置">
        <div class="mb-4 flex space-x-2">
          <div>尚未制作完成</div>
        </div>
      </n-tab-pane>
    </n-tabs>

  </n-modal>
</template>

<style lang="scss">
.modalX {
  min-width: 83.333333%; // 5/6
}

@media (min-width: 768px) {
  .modalX {
    min-width: 66.666667%; // 2/3
  }
}
</style>
