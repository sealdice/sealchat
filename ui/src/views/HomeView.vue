<script setup lang="tsx">
import Chat from './chat/chat.vue'
import ChatHeader from './components/header.vue'
</script>

<template>
  <main class="h-screen">
    <n-layout-header style="height: 4rem; padding: 24px" bordered>
      <chat-header />
    </n-layout-header>

    <n-layout has-sider position="absolute" style="margin-top: 4rem;">
      <!-- <n-layout-sider :collapsed="false" content-style="padding: 24px;" :native-scrollbar="false" bordered>
        <n-h2 v-for="i in channelList">{{ i.name }}</n-h2>
      </n-layout-sider> -->

      <n-layout>
        <Chat />
      </n-layout>
    </n-layout>
  </main>
</template>

<style lang="scss"></style>
