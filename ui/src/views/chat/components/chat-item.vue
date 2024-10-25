<script setup lang="tsx">
import dayjs from 'dayjs';
import Element from '@satorijs/element'
import { onMounted, ref, h, computed } from 'vue';
import { urlBase } from '@/stores/_config';
import DOMPurify from 'dompurify';
import { useUserStore } from '@/stores/user';
import { useChatStore } from '@/stores/chat';
import { useUtilsStore } from '@/stores/utils';
import { Howl, Howler } from 'howler';
import { useMessage } from 'naive-ui';
import Avatar from '@/components/avatar.vue'

const user = useUserStore();
const chat = useChatStore();
const utils = useUtilsStore();

function timeFormat(time?: string) {
  if (!time) return '未知';
  // console.log('???', time, typeof time)
  // return dayjs(time).format('MM-DD HH:mm:ss');
  return dayjs(time).fromNow();
}

function timeFormat2(time?: string) {
  if (!time) return '未知';
  // console.log('???', time, typeof time)
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss');
}

let hasImage = ref(false);

const parseContent = (props: any) => {
  const content = props.content;
  const items = Element.parse(content);
  let textItems = []

  for (const item of items) {
    switch (item.type) {
      case 'img':
        if (item.attrs.src && item.attrs.src.startsWith('id:')) {
          item.attrs.src = item.attrs.src.replace('id:', `${urlBase}/api/v1/attachments/`);
        }
        textItems.push(DOMPurify.sanitize(item.toString()));
        hasImage.value = true;
        break;
      case 'audio':
        let src = ''
        if (!item.attrs.src) break;

        src = item.attrs.src;
        if (item.attrs.src.startsWith('id:')) {
          src = item.attrs.src.replace('id:', `${urlBase}/api/v1/attachments/`);
        }

        let info = utils.sounds.get(src);

        if (!info) {
          const sound = new Howl({
            src: [src],
            html5: true
          });

          info = {
            sound,
            time: 0,
            playing: false
          }
          utils.sounds.set(src, info);
          utils.soundsTryInit()
        }

        const doPlay = () => {
          if (!info) return;
          if (info.playing) {
            info.sound.pause();
            info.playing = false;
          } else {
            info.sound.play();
            info.playing = true;
          }
        }

        textItems.push(<n-button rounded onClick={doPlay} type="primary">
          {info.playing ? `暂停 ${Math.floor(info.time)}/${Math.floor(info.sound.duration()) || '-'}` : '播放'}
        </n-button>)
        // textItems.push(DOMPurify.sanitize(item.toString()));
        // hasImage.value = true;
        break;
      case "at":
        if (item.attrs.id == user.info.id) {
          textItems.push(`<span class="text-blue-500 bg-gray-400 px-1" style="white-space: pre-wrap">@${item.attrs.name}</span>`);
        } else {
          textItems.push(`<span class="text-blue-500" style="white-space: pre-wrap">@${item.attrs.name}</span>`);
        }
      default:
        textItems.push(`<span style="white-space: pre-wrap">${item.toString()}</span>`);
        break;
    }
  }

  return <span>
    {textItems.map((item) => {
      if (typeof item === 'string') {
        return <span v-html={item}></span>
      } else {
        // vnode
        return item;
      }
    })}
  </span>
}

const props = defineProps({
  username: String,
  content: String,
  avatar: String,
  isRtl: Boolean,
  item: Object,
})

const timeText = ref(timeFormat(props.item?.createdAt));
const timeText2 = ref(timeFormat2(props.item?.createdAt));

const onContextMenu = (e: MouseEvent, item: any) => {
  e.preventDefault();
  //Set the mouse position

  chat.messageMenu.optionsComponent.x = e.x;
  chat.messageMenu.optionsComponent.y = e.y;
  //Show menu
  chat.messageMenu.show = true;
  chat.messageMenu.item = item;
  chat.messageMenu.hasImage = hasImage.value;
}

const message = useMessage()
const doAvatarClick = (e: MouseEvent) => {
  if (!props.item?.member?.nick) {
    message.warning('此用户无法查看')
    return;
  }
  chat.avatarMenu.show = true;

  chat.messageMenu.optionsComponent.x = e.x;
  chat.messageMenu.optionsComponent.y = e.y;
  chat.avatarMenu.item = props.item as any;
  emit('avatar-click')
}

const emit = defineEmits(['avatar-longpress', 'avatar-click']);

onMounted(() => {
  setInterval(() => {
    timeText.value = timeFormat(props.item?.createdAt);
    timeText2.value = timeFormat2(props.item?.createdAt);
  }, 10000);
})

const nick = computed(() => {
  if (props.item?.sender_member_name) {
    return props.item?.sender_member_name;
  }
  return props.item?.member?.nick || props.item?.user?.name || '未知';
});

</script>

<template>
  <div v-if="item?.is_revoked" class="py-4 text-center">一条消息已被撤回</div>
  <div v-else :id="item?.id" class="chat-item" :style="props.isRtl ? { direction: 'rtl' } : {}"
    :class="props.isRtl ? ['is-rtl'] : []">
    <Avatar :src="props.avatar" @longpress="emit('avatar-longpress')" @click="doAvatarClick" />
    <!-- <img class="rounded-md w-12 h-12 border-gray-500 border" :src="props.avatar" /> -->
    <!-- <n-avatar :src="imgAvatar" size="large" bordered>海豹</n-avatar> -->
    <div class="right">
      <span class="title">
        <!-- 右侧 -->
        <n-popover trigger="hover" placement="bottom" v-if="props.isRtl">
          <template #trigger>
            <span class="time">{{ timeText }}</span>
          </template>
          <span>{{ timeText2 }}</span>
        </n-popover>
        <span v-if="props.isRtl" class="name">{{ nick }}</span>

        <span v-if="!props.isRtl" class="name">{{ nick }}</span>
        <n-popover trigger="hover" placement="bottom" v-if="!props.isRtl">
          <template #trigger>
            <span class="time">{{ timeText }}</span>
          </template>
          <span>{{ timeText2 }}</span>
        </n-popover>

        <!-- <span v-if="props.isRtl" class="time">{{ timeText }}</span> -->
        <span v-if="props.item?.user?.is_bot || props.item?.user_id?.startsWith('BOT:')"
          class=" bg-blue-500 rounded-md px-2 text-white">bot</span>
      </span>
      <div class="content break-all relative" @contextmenu="onContextMenu($event, item)">
        <!-- <div v-html="parseContent(props)" @contextmenu="onContextMenu($event, item)"></div> -->
        <div>
          <div v-if="props.item?.quote?.id" class="border-l-4 pl-2 border-blue-500  mb-2">
            <span v-if="props.item?.quote?.is_revoked" class="text-gray-400">此消息已撤回</span>
            <span v-else class="text-gray-500">
              <component :is="parseContent(props.item?.quote)" />
            </span>
          </div>
          <component :is="parseContent(props)" />
        </div>
        <div v-if="props.item?.failed" class="failed absolute bg-red-600 rounded-md px-2 text-white">!</div>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
.chat-item {
  @apply flex;

  >.n-avatar {
    @apply rounded-md;
  }

  &.is-rtl {
    >.right {
      @apply mr-4;

      >.title {
        @apply justify-end;
      }

      >.content {
        &>.failed {
          left: -2rem;
          right: auto;
          top: 0;
        }

        &:before {
          display: none;
        }

        &::after {
          position: absolute;
          top: 0.5rem;
          height: 0.75rem;
          width: 0.75rem;
          background-color: inherit;
          content: "";
          right: -0.75rem;
          transform: scaleY(-1) scaleX(-1);
          mask-size: contain;
          mask-image: url("data:image/svg+xml,%3csvg width='3' height='3' xmlns='http://www.w3.org/2000/svg'%3e%3cpath fill='black' d='m 0 3 L 3 3 L 3 0 C 3 1 1 3 0 3'/%3e%3c/svg%3e");
        }
      }
    }
  }

  >.right {
    @apply ml-4;

    >.title {
      display: flex;
      gap: 0.5rem;
      direction: ltr;

      >.name {
        @apply font-semibold;
      }

      >.time {
        @apply text-gray-400;
      }
    }

    >.content {
      &>.failed {
        right: -2rem;
        top: 0;
      }

      &:before {
        position: absolute;
        top: 0.5rem;
        height: 0.75rem;
        width: 0.75rem;
        background-color: inherit;
        content: "";
        left: -0.75rem;
        transform: scaleY(-1);
        mask-size: contain;
        mask-image: url("data:image/svg+xml,%3csvg width='3' height='3' xmlns='http://www.w3.org/2000/svg'%3e%3cpath fill='black' d='m 0 3 L 3 3 L 3 0 C 3 1 1 3 0 3'/%3e%3c/svg%3e");
      }

      width: fit-content;
      direction: ltr;
      @apply text-base mt-1 px-4 py-2 relative;
      @apply rounded bg-gray-200 text-gray-900;
    }
  }

  .content img {
    max-width: min(36vw, 200px);
  }
}
</style>
