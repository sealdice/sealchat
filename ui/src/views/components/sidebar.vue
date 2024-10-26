<script setup lang="tsx">
import router from '@/router';
import { useChatStore } from '@/stores/chat';
import { useUserStore } from '@/stores/user';
import { Plus } from '@vicons/tabler';
import { Menu, SettingsSharp } from '@vicons/ionicons5';
import { NIcon, useDialog, useMessage } from 'naive-ui';
import { computed, ref, type Component, h, defineAsyncComponent, watch, onMounted } from 'vue';
import Notif from '../notif.vue'
import UserProfile from './user-profile.vue'
import { useI18n } from 'vue-i18n'
import { setLocale, setLocaleByNavigator } from '@/lang';
import type { Channel } from '@satorijs/protocol';
import IconNumber from '@/components/icons/IconNumber.vue'
import IconFluentMention24Filled from '@/components/icons/IconFluentMention24Filled.vue'
import ChannelSettings from './ChannelSettings/ChannelSettings.vue'
import ChannelCreate from './ChannelCreate.vue'
import UserLabel from '@/components/UserLabel.vue'
import { Setting } from '@icon-park/vue-next';
import SidebarPrivate from './sidebar-private.vue';

const { t } = useI18n()

const notifShow = ref(false)
const userProfileShow = ref(false)
const adminShow = ref(false)
const chat = useChatStore();
const user = useUserStore();

const renderIcon = (icon: Component) => {
  return () => {
    return h(NIcon, null, {
      default: () => h(icon)
    })
  }
}

const message = useMessage()
const usernameOverlap = ref(false);
const dialog = useDialog()

const showModal = ref(false);

const doChannelSwitch = async (i: Channel) => {
  const success = await chat.channelSwitchTo(i.id);
  if (!success) {
    message.error('切换频道失败，你可能没有权限');
  }
}

const showModal2 = ref(false);
const channelToSettings = ref<SChannel | undefined>(undefined);
const doSetting = async (i: Channel) => {
  channelToSettings.value = i;
  showModal2.value = true;
}

import { useSpeechRecognition } from '@vueuse/core'

// const {
//   isSupported,
//   isListening,
//   isFinal,
//   result,
//   start,
//   stop,
// } = useSpeechRecognition()

const speech = useSpeechRecognition({
  lang: 'zh-CN',
  interimResults: true,
  continuous: true,
})

const { isListening, isSupported, stop, result } = speech

if (speech.isSupported.value) {
  // @ts-expect-error missing types
  const SpeechGrammarList = window.SpeechGrammarList || window.webkitSpeechGrammarList
  const speechRecognitionList = new SpeechGrammarList()
  // speechRecognitionList.addFromString(grammar, 1)
  speech.recognition!.grammars = speechRecognitionList

  watch(speech.result, () => {
  })
}

const startA = () => {
  speech.result.value = ''
  speech.start()
}

import { useSpeechSynthesis } from '@vueuse/core'
import type { SChannel } from '@/types';

const voice = ref<SpeechSynthesisVoice>(undefined as unknown as SpeechSynthesisVoice)
const voices = ref<SpeechSynthesisVoice[]>([])

const synth = useSpeechSynthesis(speech.result, {
  voice,
  pitch: 1,
  rate: 1,
  volume: 1,
})

onMounted(() => {
  if (speech.isSupported.value) {
    // load at last
    setTimeout(() => {
      const synth = window.speechSynthesis
      voices.value = synth.getVoices()
      voice.value = voices.value[0]
    })
  }
})

const speak = () => {
  if (synth.status.value === 'pause') {
    console.log('resume')
    window.speechSynthesis.resume()
  }
  else {
    synth.speak()
  }
}

const parentId = ref('');

const handleSelect = (key: string, data: any) => {
  switch (key) {
    case 'enter':
      doChannelSwitch(data.item);
      break;
    case 'addSubChannel':
      // 实现添加子频道的逻辑
      parentId.value = data.item.id;
      showModal.value = true;
      break;
    case 'manage':
      // 实现管理频道的逻辑
      doSetting(data.item);
      break;
    case 'leave':
      // 实现退出频道的逻辑
      alert('未实现');
      break;
    case 'dissolve':
      // 实现解散频道的逻辑
      alert('未实现');
      break;
    default:
      break;
  }
}

const suffix = (item: SChannel) => {
  if (item.permType === 'non-public') {
    return '[*]'
  }
  return ''
}

</script>

<template>
  <div class="w-full h-full sc-sidebar sc-sidebar-fill">
    <n-tabs type="segment" v-model:value="chat.sidebarTab" tab-class="sc-sidebar-fill" pane-class="sc-sidebar-fill">
      <n-tab-pane name="channels" tab="频道">
        <!-- 频道列表内容将在这里显示 -->
        <div class="space-y-1 flex flex-col px-2">
          <template v-if="chat.curChannel?.name">
            <!-- <template v-if="false"> -->
            <template v-for="i in chat.channelTree">
              <div class="sider-item" :class="i.id === chat.curChannel?.id ? ['active'] : []"
                @click="doChannelSwitch(i)">

                <div class="flex space-x-1 items-center">
                  <template v-if="(i.type === 3 || (i as any).isPrivate)">
                    <!-- 私聊 -->
                    <n-icon :component="IconFluentMention24Filled"></n-icon>
                    <span>{{ `${i.name}` }}</span>
                  </template>

                  <template v-else>
                    <!-- 公开频道 -->
                    <n-icon :component="IconNumber"></n-icon>
                    <span>{{ `${i.name}${suffix(i)} (${(i as any).membersCount})` }}</span>
                  </template>
                </div>

                <div class="right">
                  <div class="flex justify-center space-x-1">
                    <n-dropdown trigger="click" :options="[
                      { label: '进入', key: 'enter', item: i },
                      { label: '添加子频道', key: 'addSubChannel', show: !Boolean(i.parentId), item: i },
                      { label: '频道管理', key: 'manage', item: i },
                      { label: '退出', key: 'leave', item: i, show: i.permType === 'non-public' },
                      { label: '解散', key: 'dissolve', item: i, }
                    ]" @select="handleSelect">
                      <n-button text @click.stop quaternary circle size="tiny">
                        <template #icon>
                          <n-icon>
                            <Menu />
                          </n-icon>
                        </template>
                      </n-button>
                    </n-dropdown>
                    <n-button quaternary circle size="tiny" @click.stop="handleSelect('manage', { item: i })">
                      <template #icon>
                        <SettingsSharp />
                      </template>
                    </n-button>
                  </div>
                </div>

              </div>

              <!-- 当前频道的用户列表 -->
              <div class="pl-5 mt-2 space-y-2" v-if="i.id == chat.curChannel.id && chat.curChannelUsers.length">
                <UserLabel :name="j.nick" :src="j.avatar" v-for="j in chat.curChannelUsers" />
              </div>

              <div v-if="(i.children?.length ?? 0) > 0">
                <template v-for="child in i.children">
                  <div class="sider-item" :class="child.id === chat.curChannel?.id ? ['active'] : []"
                    @click="doChannelSwitch(child)">
                    <div class="flex space-x-1 items-center pl-4">
                      <template v-if="(child.type === 3 || (child as any).isPrivate)">
                        <n-icon :component="IconFluentMention24Filled"></n-icon>
                        <span>{{ `${child.name}` }}</span>
                      </template>
                      <template v-else>
                        <n-icon :component="IconNumber"></n-icon>
                        <span>{{ `${child.name}${suffix(child)} (${(child as any).membersCount})` }}</span>
                      </template>
                    </div>
                    <div class="right">

                      <div class="flex justify-center space-x-1">
                        <n-dropdown trigger="click" :options="[
                          { label: '进入', key: 'enter', item: child },
                          { label: '频道管理', key: 'manage', item: child },
                          { label: '退出', key: 'leave', item: i, show: i.permType === 'non-public' },
                          { label: '解散', key: 'dissolve', item: i, }
                        ]" @select="handleSelect">
                          <n-button text @click.stop quaternary circle size="tiny">
                            <template #icon>
                              <n-icon>
                                <Menu />
                              </n-icon>
                            </template>
                          </n-button>
                        </n-dropdown>

                        <n-button quaternary circle size="tiny" @click.stop="handleSelect('manage', { item: i })">
                          <template #icon>
                            <SettingsSharp />
                          </template>
                        </n-button>
                      </div>

                    </div>
                  </div>


                  <!-- 当前频道的用户列表 -->
                  <div class="pl-8 mt-2 space-y-2" v-if="child.id == chat.curChannel.id && chat.curChannelUsers.length">
                    <UserLabel :name="j.nick" :src="j.avatar" v-for="j in chat.curChannelUsers" />
                  </div>
                </template>
              </div>

            </template>

          </template>
          <template v-else>
            <div class="px-6">加载中 ...</div>
          </template>

          <div class="sider-item" @click="parentId = ''; showModal = true">
            <div class="flex space-x-1 items-center font-bold">
              <n-icon :component="Plus"></n-icon>
              <span>{{ t('channelListNew') }}</span>
            </div>
          </div>
        </div>
      </n-tab-pane>
      <n-tab-pane name="privateChats" tab="私聊">
        <!-- 私聊列表内容将在这里显示 -->
        <SidebarPrivate />
      </n-tab-pane>
    </n-tabs>
  </div>


  <!-- <div v-if="!isSupported">
      Your browser does not support SpeechRecognition API,
      <a href="https://caniuse.com/mdn-api_speechrecognition" target="_blank">more details</a>
    </div>
    <div v-else class="mt-8">
      <n-button v-if="!isListening" @click="startA">
        按下说话
      </n-button>
      <n-button v-if="isListening" class="orange" @click="stop">
        停止
      </n-button>
      <div v-if="isListening" class="">
        {{ speech.result }}
      </div>

      <div class="mt-8">
        <select v-model="voice" px-8 border-0 bg-transparent h-9 rounded appearance-none>
          <option bg="$vp-c-bg" disabled>
            Select Language
          </option>
          <option v-for="(voice, i) in voices" :key="i" bg="$vp-c-bg" :value="voice">
            {{ `${voice.name} (${voice.lang})` }}
          </option>
        </select>

        <n-button @click="speak">复读</n-button>
      </div>
    </div> -->
  <ChannelCreate v-model:show="showModal" :parentId="parentId" />
  <ChannelSettings :channel="channelToSettings" v-model:show="showModal2" />
</template>

<style lang="scss">
.sider-item {
  @apply rounded px-2 py-2 cursor-pointer flex justify-between;

  &:hover {
    @apply bg-blue-100;

    >.right {
      @apply block;
    }
  }

  &.active {
    @apply bg-blue-200;

  }

  >.right {
    @apply hidden;
  }
}
</style>
