<script setup lang="ts">
import router from '@/router';
import { onMounted, reactive, ref } from 'vue';
import { formDataToJson } from '@/utils/tools';
import type { AxiosError } from 'axios';
import { flow } from 'lodash-es';
import { useUserStore } from '@/stores/user';
import { useMessage } from 'naive-ui';
import { useUtilsStore } from '@/stores/utils';
import type { ServerConfig } from '@/types';

const userStore = useUserStore();

const form = reactive({
  username: '',
  password: '',
  password2: '',
  nickname: '',
})

const message = useMessage()

const signUp = async () => {
  const ret = await userStore.signUp(form);
  if (ret) {
    message.error(ret)
  } else {
    message.success('注册成功，即将返回首页')
    router.replace({ name: 'home' })
  }
}

const randomUsername = () => {
  const characters = 'abcdefghjkmnpqrstuvwxyz';
  const characters2 = 'abcdefghjkmnpqrstuvwxyz23456789';
  let result = '';
  for (let i = 0; i < 1; i++) {
    result += characters.charAt(Math.floor(Math.random() * characters.length));
  }
  for (let i = 0; i < 4; i++) {
    result += characters2.charAt(Math.floor(Math.random() * characters2.length));
  }
  form.username = result;
}

const utils = useUtilsStore();
const config = ref<ServerConfig | null>(null)

onMounted(async () => {
  const resp = await utils.configGet();
  config.value = resp.data;
})
</script>

<template>
  <div class="flex items-center justify-center h-full w-full">
    <div class="w-full max-w-sm mx-auto overflow-hidden bg-white rounded-lg shadow-md dark:bg-gray-800"
      v-if="config?.registerOpen">
      <div class="px-6 py-4">
        <div class="flex justify-center mx-auto">
          <!-- <img class="w-auto h-7 sm:h-8" src="https://merakiui.com/images/logo.svg" alt=""> -->
        </div>

        <h3 class="mt-3 text-xl font-medium text-center text-gray-600 dark:text-gray-200">注册</h3>

        <div style="font-size: 12; overflow-y: auto; max-height: 10rem;">
          <!-- {{ authStore.session }} -->
        </div>

        <form class="min-w-20rem">

          <div class="w-full mt-4">
            <div class="relative">
              <input v-model="form.username"
                class="block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-500 bg-white border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 focus:border-blue-400 dark:focus:border-blue-300 focus:ring-opacity-40 focus:outline-none focus:ring focus:ring-blue-300"
                type="username" placeholder="用户名，用于登录和识别，可被其他人看到" aria-label="用户名" />
              <button @click.prevent="randomUsername"
                class="absolute right-0 h-full top-0 px-1 mr-1 text-sm font-medium text-blue-500 capitalize" tabindex="-1">随机
              </button>
            </div>
          </div>

          <div class="w-full mt-4">
            <input v-model="form.nickname"
              class="block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-500 bg-white border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 focus:border-blue-400 dark:focus:border-blue-300 focus:ring-opacity-40 focus:outline-none focus:ring focus:ring-blue-300"
              type="text" placeholder="昵称" aria-label="昵称" />
          </div>

          <div class="w-full mt-4">
            <input v-model="form.password"
              class="block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-500 bg-white border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 focus:border-blue-400 dark:focus:border-blue-300 focus:ring-opacity-40 focus:outline-none focus:ring focus:ring-blue-300"
              type="password" placeholder="密码" aria-label="密码" />
          </div>

          <!-- <div class="w-full mt-4">
            <input v-model="form.password2"
              class="block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-500 bg-white border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 focus:border-blue-400 dark:focus:border-blue-300 focus:ring-opacity-40 focus:outline-none focus:ring focus:ring-blue-300"
              type="password" placeholder="重复密码" aria-label="重复密码" />
          </div> -->

          <div class="flex items-center justify-between mt-4">
            <div></div>
            <!-- <a href="#" class="text-sm text-gray-600 dark:text-gray-200 hover:text-gray-500">忘记密码</a> -->

            <button @click.prevent="signUp"
              class="px-6 py-2 text-sm font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-blue-500 rounded-lg hover:bg-blue-400 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-50">
              注册
            </button>
          </div>
        </form>
      </div>

      <div class="flex items-center justify-center py-4 text-center bg-gray-50 dark:bg-gray-700">
        <span class="text-sm text-gray-600 dark:text-gray-200">已有账号 ？</span>
        <router-link :to="{ name: 'user-signin' }"
          class="mx-2 text-sm font-bold text-blue-500 dark:text-blue-400 hover:underline">登录</router-link>
      </div>
    </div>
    <div class="w-full max-w-sm mx-auto overflow-hidden bg-white rounded-lg shadow-md dark:bg-gray-800" v-else>
      <div class="p-6">你来晚了，门已经悄然关闭。</div>
    </div>
  </div>
</template>

<style></style>
