<script setup lang="tsx">
import router from '@/router';
import { useChatStore } from '@/stores/chat';
import { useUserStore } from '@/stores/user';
import { Plus } from '@vicons/tabler';
import { NIcon, useDialog, useMessage } from 'naive-ui';
import { computed, ref, type Component, h, defineAsyncComponent } from 'vue';
import Notif from './notif.vue'
import UserProfile from './user-profile.vue'
// import AdminSettings from './admin-settings.vue'
import { useI18n } from 'vue-i18n'
import { setLocale, setLocaleByNavigator } from '@/lang';

const AdminSettings = defineAsyncComponent(() => import('./admin-settings.vue'));

const { t } = useI18n()

const notifShow = ref(false)
const userProfileShow = ref(false)
const adminShow = ref(false)
const chat = useChatStore();
const user = useUserStore();

const options = computed(() => [
  {
    label: t('headerMenu.profile'),
    key: 'profile',
    // icon: renderIcon(UserIcon)
  },
  user.info.role === 'role-admin' ? {
    label: t('headerMenu.admin'),
    key: 'admin',
    // icon: renderIcon(UserIcon)
  } : null,
  {
    label: t('headerMenu.lang'),
    key: 'lang',
    children: [
      {
        label: t('headerMenu.langAuto'),
        key: 'lang:auto'
      },
      {
        label: '简体中文',
        key: 'lang:zh-cn'
      },
      {
        label: 'English',
        key: 'lang:en'
      },
      {
        label: '日本語',
        key: 'lang:ja'
      }
    ]
    // icon: renderIcon(UserIcon)
  },
  {
    label: t('headerMenu.notice'),
    key: 'notice',
    // icon: renderIcon(UserIcon)
  },
  {
    label: t('headerMenu.logout'),
    key: 'logout',
    // icon: renderIcon(LogoutIcon)
  }
].filter(i => i != null))


const handleSelect = async (key: string | number) => {
  switch (key) {
    case 'notice':
      userProfileShow.value = false;
      adminShow.value = false;
      notifShow.value = !notifShow.value;
      break;

    case 'profile':
      notifShow.value = false;
      adminShow.value = false;
      userProfileShow.value = !userProfileShow.value;
      break;

    case 'admin':
      notifShow.value = false;
      userProfileShow.value = false;
      adminShow.value = !adminShow.value;
      break;

    case 'logout':
      dialog.warning({
        title: t('dialogLogOut.title'),
        content: t('dialogLogOut.content'),
        positiveText: t('dialogLogOut.positiveText'),
        negativeText: t('dialogLogOut.negativeText'),
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
      if (typeof key == "string" && key.startsWith('lang:')) {
        if (key == 'lang:auto') {
          setLocaleByNavigator();
        } else {
          setLocale(key.replace('lang:', ''));
        }
      }
      break;
  }
}

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
      label: (i.type === 3 || (i as any).isPrivate) ? i.name : `${i.name} (${(i as any).membersCount})`,
      key: i.id,
      icon: undefined as any,
      props: undefined as any,
    }
  })
  lst.push({ label: t('channelListNew'), key: 'new', icon: renderIcon(Plus), props: { style: { 'font-weight': 'bold' } } })
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
    message.error(t('dialoChannelgNew.channelNameHint'));
    return;
  }
  await chat.channelCreate(newChannelName.value);
  await chat.channelList();
}
</script>

<template>
  <div class="flex justify-between items-center">
    <div>
      <span class="text-sm font-bold sm:text-xl">{{ $t('headText') }}</span>
      <!-- <n-button>登录</n-button>
      <n-button>切换房间</n-button> -->
      <span class="ml-4">
        <n-dropdown trigger="click" :options="chOptions" @select="channelSelect">
          <!-- <n-button>{{ chat.curChannel?.name || '加载中 ...' }}</n-button> -->
          <n-button text v-if="(chat.curChannel?.type === 3 || (chat.curChannel as any)?.isPrivate)">{{
            chat.curChannel?.name ? `${chat.curChannel?.name}` : '加载中 ...' }} ▼</n-button>
          <n-button text v-else>{{
            chat.curChannel?.name ? `${chat.curChannel?.name} (${(chat.curChannel as
              any).membersCount})`
            : '加载中 ...' }} ▼</n-button>
        </n-dropdown>
      </span>
    </div>

    <div class="space-x-8 flex items-center">
      <!-- ● -->
      <span v-if="chat.connectState === 'connecting'" class=" text-blue-500">{{ $t('connectState.connecting') }}</span>
      <span v-if="chat.connectState === 'connected'" class=" text-green-600">{{ $t('connectState.connected') }}</span>
      <span v-if="chat.connectState === 'disconnected'" class=" text-red-500">{{ $t('connectState.disconnected') }}</span>
      <span v-if="chat.connectState === 'reconnecting'" class=" text-orange-400">{{ $t('connectState.reconnecting',
        [chat.iReconnectAfterTime]) }}</span>

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

  <n-modal v-model:show="showModal" preset="dialog" :title="$t('dialoChannelgNew.title')"
    :positive-text="$t('dialoChannelgNew.positiveText')" :negative-text="$t('dialoChannelgNew.negativeText')"
    @positive-click="newChannel">
    <n-input v-model:value="newChannelName"></n-input>
  </n-modal>

  <div v-if="userProfileShow" style="background-color: var(--n-color); margin-left: -1.5rem;"
    class="absolute flex justify-center items-center w-full h-full pointer-events-none z-10">
    <user-profile @close="userProfileShow = false" />
  </div>
  <div v-if="adminShow" style="background-color: var(--n-color); margin-left: -1.5rem;"
    class="absolute flex justify-center items-center w-full h-full pointer-events-none z-10">
    <admin-settings @close="adminShow = false" />
  </div>
  <notif v-show="notifShow" />
</template>
