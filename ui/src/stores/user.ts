import { defineStore } from "pinia"
import type { UserInfo } from "@/types";
// import Cookies from 'js-cookie';
// import router from "@/router";

import axiosFactory from "axios"
import { cloneDeep } from "lodash-es";

import type { AxiosResponse } from "axios";
import { api } from "./_config";
import { useChatStore } from "./chat";

interface UserState {
  _accessToken: string
  info: UserInfo;
  lastCheckTime: number;
}

export const useUserStore = defineStore({
  id: 'user',

  state: (): UserState => ({
    _accessToken: '',
    lastCheckTime: 0,
    // 这样比info?好的地方在于可以建立watch关联
    info: {
      id: "",
      createdAt: "",
      updatedAt: "",
      deletedAt: null,
      username: "",
      nick: ''
    },
  }),

  getters: {
    token: (state) => {
      if (!state._accessToken) {
        state._accessToken = localStorage.getItem('accessToken') || '';
        // state._accessToken = Cookies.get('accessToken') || '';
      }
      return state._accessToken;
    }
  },

  actions: {
    async signIn(username: string, password: string) {
      // 在此处进行用户鉴权操作，获取 accessToken
      const resp = await api.post('api/v1/user/signin', {
        username, password
      })

      const data = resp.data as { token: string, message: string };
      const accessToken = data.token;

      // 将 accessToken 存入 localStorage 中
      // Cookies.set('accessToken', accessToken, { expires: 7 })
      localStorage.setItem('accessToken', accessToken)

      // 更新 state 中的 accessToken
      this._accessToken = accessToken

      return resp
    },

    async checkUserSession() {
      const now = Number(Date.now());
      if (now + this.lastCheckTime > 60 * 1000) {
        // 向服务器发请求
        try {
          const resp = await api.get('api/v1/user/info', {
            headers: { 'Authorization': this.token }
          })
          if (!this.info) {
            // 初次进入
            useChatStore().tryInit();
          }
          this.info = resp.data.user as UserInfo;
          // console.log('check', this.info)
          this.lastCheckTime = Number(Date.now());
          return true;
        } catch (e: any) {
          console.log(222, e);
          if (e.code !== "ERR_NETWORK") {
            // 未登录，清除数据
            this.info.id = '';
            localStorage.setItem('accessToken', '');
            this._accessToken = '';
          }
          return false;
        }
      } else {
        if (this.token) return true;
      }
    },

    async signUp(form: { username: string, password: string, nickname: string }) {
      try {
        // 在此处进行用户鉴权操作，获取 accessToken
        const resp = await api.post('api/v1/user/signup', {
          username: form.username,
          password: form.password,
          nickname: form.nickname,
        })

        const data = resp.data as { token: string, message: string };
        const accessToken = data.token;

        // 将 accessToken 存入 localStorage 中
        localStorage.setItem('accessToken', accessToken)
        // Cookies.set('accessToken', accessToken, { expires: 7 })

        // 更新 state 中的 accessToken
        this._accessToken = accessToken

        return ''
      } catch (err) {
        // console.error('Authentication failed:', err)
        return (err as any).data?.message || '错误';
      }
    },

    logout() {
      // 将 accessToken 从 localStorage 中删除
      localStorage.removeItem('accessToken')
      this.info.id = ''
      // 更新 state 中的 accessToken
      this._accessToken = ''
    },
  },
})
