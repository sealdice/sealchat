<script setup lang="tsx">
import router from '@/router';
import { useChatStore } from '@/stores/chat';
import { useUserStore } from '@/stores/user';
import { Plus } from '@vicons/tabler';
import { NIcon, useDialog, useMessage } from 'naive-ui';
import { computed, ref, type Component, h } from 'vue';
import Notif from './notif.vue'
import UserProfile from './user-profile.vue'

const notifShow = ref(false)
const userProfileShow = ref(false)
const chat = useChatStore();
const user = useUserStore();

const options = [
  {
    label: '个人信息',
    key: 'profile',
    // icon: renderIcon(UserIcon)
  },
  {
    label: '消息提醒',
    key: 'notice',
    // icon: renderIcon(UserIcon)
  },
  {
    label: '退出账号',
    key: 'logout',
    // icon: renderIcon(LogoutIcon)
  }
]

const renderIcon = (icon: Component) => {
  return () => {
    return h(NIcon, null, {
      default: () => h(icon)
    })
  }
}

const chOptions = computed(() => {
  const lst = chat.channelTree.map(i => {
    return {
      label: `${i.name} (${(i as any).membersCount})`,
      key: i.id,
      icon: undefined as any,
      props: undefined as any,
    }
  })
  lst.push({ label: '新建', key: 'new', icon: renderIcon(Plus), props: { style: { 'font-weight': 'bold' } } })
  return lst;
})

const channelSelect = async (key: string) => {
  if (key === 'new') {
    showModal.value = true;
    // chat.channelCreate('测试频道');
    // message.info('暂不支持新建频道');
  } else {
    await chat.channelSwitchTo(key);
  }
}

const message = useMessage()
const usernameOverlap = ref(false);
const dialog = useDialog()

const showModal = ref(false);
const newChannelName = ref('');
const newChannel = async () => {
  if (!newChannelName.value.trim()) {
    message.error('频道名不能为空');
    return;
  }
  await chat.channelCreate(newChannelName.value);
  await chat.channelList();
}

const handleSelect = async (key: string | number) => {
  console.log('!!!', key)
  switch (key) {
    case 'notice':
      notifShow.value = !notifShow.value;
      break;

    case 'profile':
      userProfileShow.value = !userProfileShow.value;
      break;

    case 'logout':
      dialog.warning({
        title: '警告',
        content: '退出登录？',
        positiveText: '是的',
        negativeText: '取消',
        onPositiveClick: () => {
          user.logout();
          chat.subject?.unsubscribe();
          window.location.reload();
          // router.push({ name: 'user-signin' });
        },
        onNegativeClick: () => {
        }
      })
      break;

    default:
      break;
  }
}
</script>

<template>
  <div class="flex justify-between items-center">
    <div>
      <span class="text-sm font-bold sm:text-xl">海豹尬聊</span>
      <!-- <n-button>登录</n-button>
      <n-button>切换房间</n-button> -->
      <span class="ml-4">
        <n-dropdown trigger="click" :options="chOptions" @select="channelSelect">
          <!-- <n-button>{{ chat.curChannel?.name || '加载中 ...' }}</n-button> -->
          <n-button text>{{ chat.curChannel?.name ? `${chat.curChannel?.name} (${(chat.curChannel as
            any).membersCount})`
            : '加载中 ...' }} ▼</n-button>
        </n-dropdown>
      </span>

    </div>
    <div class="space-x-8 flex items-center">
      <!-- ● -->
      <span v-if="chat.connectState === 'connecting'" class=" text-blue-500">连接中 ...</span>
      <span v-if="chat.connectState === 'connected'" class=" text-green-600">已连接</span>
      <span v-if="chat.connectState === 'disconnected'" class=" text-red-500">已断开</span>
      <span v-if="chat.connectState === 'reconnecting'" class=" text-orange-400">{{ chat.iReconnectAfterTime }}s
        后自动重连</span>

      <n-dropdown :overlap="usernameOverlap" placement="bottom-end" trigger="click" :options="options"
        @select="handleSelect">
        <span class="flex justify-center cursor-pointer">
          <span>{{ user.info.nick }}</span>
          <svg style="width: 1rem" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
            viewBox="0 0 24 24">
            <path d="M7 10l5 5l5-5H7z" fill="currentColor"></path>
          </svg>
        </span>
      </n-dropdown>
      <!-- <span>
        <n-button @click="notifShow = !notifShow">N</n-button>
      </span> -->
    </div>
  </div>

  <n-modal v-model:show="showModal" preset="dialog" title="添加频道" content="你确认?" positive-text="确认" negative-text="算了"
    @positive-click="newChannel">
    <n-input v-model:value="newChannelName"></n-input>
  </n-modal>

  <div v-if="userProfileShow" style="background-color: var(--n-color); margin-left: -1.5rem;"
    class="absolute flex justify-center items-center w-full h-full pointer-events-none z-10">
    <user-profile @close="userProfileShow = false" />
  </div>
  <notif v-show="notifShow" />
</template>
