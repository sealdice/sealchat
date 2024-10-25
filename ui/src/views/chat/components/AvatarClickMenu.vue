<script setup lang="tsx">
import { useChatStore } from '@/stores/chat';
import type { MenuOptions } from '@imengyu/vue3-context-menu';
import { computed, nextTick } from 'vue';
import Element from '@satorijs/element'
import { db } from '@/models/index';
import { urlBase } from '@/stores/_config';
import { useMessage } from 'naive-ui';
import { useUserStore } from '@/stores/user';

const chat = useChatStore()
const message = useMessage()

const alertContextMenuItemClicked = (name: string) => {
  alert('You clicked ' + name + ' !');
}

const clickTalkTo = async () => {
  const data = chat.avatarMenu.item;
  if (data && data.user) {
    if (data.user.id === user.info.id) return;
    const ch = await chat.channelPrivateCreate(data.user.id);
    if (ch?.channel?.id) {
      chat.sidebarTab = 'privateChats';
      await chat.ChannelPrivateList()
      nextTick(async () => {
        await chat.channelSwitchTo(ch.channel.id);
      })
    }
  }
}

const user = useUserStore()
</script>

<template>
  <context-menu v-model:show="chat.avatarMenu.show" :options="chat.messageMenu.optionsComponent">
    <context-menu-item label="私聊" @click="clickTalkTo" />
  </context-menu>
</template>
