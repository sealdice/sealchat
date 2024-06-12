<script setup lang="tsx">
import router from '@/router';
import { useChatStore } from '@/stores/chat';
import { useUserStore } from '@/stores/user';
import { Plus } from '@vicons/tabler';
import { NIcon, useDialog, useMessage } from 'naive-ui';
import { computed, ref, type Component, h, defineAsyncComponent, watch, onMounted } from 'vue';
import Notif from '../notif.vue'
import UserProfile from './user-profile.vue'
// import AdminSettings from './admin-settings.vue'
import { useI18n } from 'vue-i18n'
import { setLocale, setLocaleByNavigator } from '@/lang';
import type { Channel } from '@satorijs/protocol';
import IconNumber from '@/components/icons/IconNumber.vue'
import IconFluentMention24Filled from '@/components/icons/IconFluentMention24Filled.vue'

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
const newChannelName = ref('');
const newChannel = async () => {
	if (!newChannelName.value.trim()) {
		message.error(t('dialoChannelgNew.channelNameHint'));
		return;
	}
	await chat.channelCreate(newChannelName.value);
	await chat.channelList();
}

const doChannelSwitch = async (i: Channel) => {
	await chat.channelSwitchTo(i.id);
}

const showModal2 = ref(false);
const doSetting = async (i: Channel) => {
	alert('还没做');
}

import { useSpeechRecognition } from '@vueuse/core'

// const {
// 	isSupported,
// 	isListening,
// 	isFinal,
// 	result,
// 	start,
// 	stop,
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

</script>

<template>
	<div class="w-full pt-4">
		<template v-if="chat.curChannel?.name">
			<div class="space-y-2 flex flex-col px-2">
				<div v-for="i in chat.channelTree" class="rounded px-2 py-2 cursor-pointer flex sider-item justify-between"
					:class="i.id === chat.curChannel?.id ? ['active'] : []" @click="doChannelSwitch(i)">

					<div class="flex space-x-1 items-center">
						<template v-if="(i.type === 3 || (i as any).isPrivate)">
							<!-- 私聊 -->
							<n-icon :component="IconFluentMention24Filled"></n-icon>
							<span>{{ `${i.name}` }}</span>
						</template>

						<template v-else>
							<!-- 公开频道 -->
							<n-icon :component="IconNumber"></n-icon>
							<span>{{ `${i.name} (${(i as any).membersCount})` }}</span>
						</template>
					</div>

					<div class="right" @click="doSetting(i)">
						设置
					</div>
				</div>

				<div class="rounded px-2 py-2 cursor-pointer flex sider-item justify-between" @click="showModal = true">
					<div class="flex space-x-1 items-center font-bold">
						<n-icon :component="Plus"></n-icon>
						<span>{{ t('channelListNew') }}</span>
					</div>
				</div>
			</div>
		</template>
		<template v-else>
			加载中 ...
		</template>


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
	</div>

	<n-modal v-model:show="showModal" preset="dialog" :title="$t('dialoChannelgNew.title')"
		:positive-text="$t('dialoChannelgNew.positiveText')" :negative-text="$t('dialoChannelgNew.negativeText')"
		@positive-click="newChannel">
		<n-input v-model:value="newChannelName"></n-input>
	</n-modal>


	<n-modal v-model:show="showModal2" preset="dialog" :title="$t('dialoChannelgNew.title')"
		:positive-text="$t('dialoChannelgNew.positiveText')" :negative-text="$t('dialoChannelgNew.negativeText')"
		@positive-click="newChannel">
		<n-input v-model:value="newChannelName"></n-input>
	</n-modal>
</template>

<style lang="scss" scoped>
.sider-item {
	&:hover {
		@apply bg-blue-50;

		>.right {
			@apply block;
		}
	}

	&.active {
		@apply bg-blue-100;

	}

	>.right {
		@apply hidden;
	}
}
</style>
