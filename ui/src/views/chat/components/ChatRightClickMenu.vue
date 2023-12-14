<script setup lang="tsx">
import { useChatStore } from '@/stores/chat';
import type { MenuOptions } from '@imengyu/vue3-context-menu';
import { computed } from 'vue';
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

const clickReplyTo = async () => {
  chat.setReplayTo(chat.messageMenu.item)
}

const user = useUserStore()

const clickDelete = async () => {
  if (chat.curChannel?.id && chat.messageMenu.item?.id) {
    await chat.messageDelete(chat.curChannel?.id, chat.messageMenu.item?.id)
    message.success('撤回成功')
  }
}

const addToMyEmoji = async () => {
  const items = Element.parse(chat.messageMenu.item?.content || '');
  for (let item of items) {
    if (item.type == "img") {
      const id = item.attrs.src.replace('id:', '');
      try {
        await db.thumbs.add({
          id: id,
          recentUsed: Number(Date.now()),
          filename: 'image.png',
          mimeType: '',
          data: null, // 无数据，按id加载
        });
        message.success('收藏成功');
      } catch (e: any) {
        if (e.name === "ConstraintError") {
          message.error('该表情已经存在于收藏了');
        }
      }
    }
  }
}
</script>

<template>
  <context-menu v-model:show="chat.messageMenu.show" :options="chat.messageMenu.optionsComponent">
    <context-menu-item v-if="chat.messageMenu.hasImage" label="添加到表情收藏" @click="addToMyEmoji" />
    <!-- <context-menu-sperator /> -->
    <!-- <context-menu-item label="Item with a icon" icon="icon-reload-1" @click="alertContextMenuItemClicked('Item2')" /> -->
    <!-- <context-menu-item label="Test Item" @click="alertContextMenuItemClicked('Item2')" /> -->
    <context-menu-item label="回复" @click="clickReplyTo" />
    <context-menu-item label="撤回" @click="clickDelete"
      v-if="chat.messageMenu.item?.user?.id && (chat.messageMenu.item?.user?.id === user.info.id)" />
    <!-- <context-menu-group label="Menu with child">
      <context-menu-item label="Item1" @click="alertContextMenuItemClicked('Item2-1')" />
      <context-menu-item label="Item1" @click="alertContextMenuItemClicked('Item2-2')" />
      <context-menu-group label="Child with v-for 50">
        <context-menu-item v-for="index of 50" :key="index" :label="'Item3-' + index"
          @click="alertContextMenuItemClicked('Item3-' + index)" />
      </context-menu-group>
    </context-menu-group> -->
  </context-menu>
</template>
