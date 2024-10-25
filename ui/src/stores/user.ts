import { defineStore } from "pinia"
import type { UserEmojiModel, UserInfo } from "@/types";
import Cookies from 'js-cookie';
// import router from "@/router";

import axiosFactory from "axios"
import { cloneDeep } from "lodash-es";

import type { AxiosResponse } from "axios";
import { api } from "./_config";
import { useChatStore } from "./chat";
import { PermResult, type PermCheckKey, type SystemRolePermSheet } from "@/types-perm";

interface UserState {
  _accessToken: string
  info: UserInfo;
  lastCheckTime: number;
  emojiCount: number,

  permSysMap: SystemRolePermSheet;
}

export const useUserStore = defineStore({
  id: 'user',

  state: (): UserState => ({
    _accessToken: '',
    lastCheckTime: 0,
    emojiCount: 1,

    permSysMap: {} as any,
    // 这样比info?好的地方在于可以建立watch关联
    info: {
      id: "",
      createdAt: "",
      updatedAt: "",
      deletedAt: null,
      username: "",
      nick: '',
      avatar: '',
      brief: '',
      disabled: false,
    },
  }),

  getters: {
    token: (state) => {
      if (!state._accessToken) {
        state._accessToken = localStorage.getItem('accessToken') || '';
        Cookies.set('Authorization', state._accessToken);
        // state._accessToken = Cookies.get('accessToken') || '';
      }
      return state._accessToken;
    }
  },

  actions: {
    async changePassword(form: { password: string, passwordNew: string }) {
      const resp = await api.post('api/v1/user-password-change', {
        password: form.password, passwordNew: form.passwordNew
      }, {
        headers: { 'Authorization': this.token }
      })

      // 密码重置后，之前的所有token都会被重置
      const data = resp.data as { token: string, message: string };
      const accessToken = data.token;
      return resp;
    },

    async signIn(username: string, password: string) {
      // 在此处进行用户鉴权操作，获取 accessToken
      const resp = await api.post('api/v1/user-signin', {
        username, password
      })

      const data = resp.data as { token: string, message: string };
      const accessToken = data.token;

      // 将 accessToken 存入 localStorage 中
      // Cookies.set('accessToken', accessToken, { expires: 7 })
      localStorage.setItem('accessToken', accessToken);

      // 更新 state 中的 accessToken
      this._accessToken = accessToken;

      return resp;
    },

    async timelineList() {
      const resp = await api.get('api/v1/timeline-list', {
        headers: { 'Authorization': this.token }
      });
      return resp;
    },

    // 强制更新用户信息
    async infoUpdate() {
      const resp = await api.get('api/v1/user-info', {
        headers: { 'Authorization': this.token }
      })

      this.info = resp.data.user as UserInfo;

      let permSysMap: { [key: string]: number } = {};
      for (let i of resp.data.permSys) {
        permSysMap[i] = PermResult.ALLOWED;
      }
      this.permSysMap = permSysMap as any;
      return this.info;
    },

    async changeInfo(info: { nick: string, brief: string }) {
      const resp = await api.post('api/v1/user-info-update', info, {
        headers: { 'Authorization': this.token }
      })
      return resp;
    },

    async checkUserSession() {
      const now = Number(Date.now());
      if (now + this.lastCheckTime > 60 * 1000) {
        // 向服务器发请求
        try {
          const firstTime = !this.info;
          await this.infoUpdate();
          if (firstTime) {
            // 初次进入
            useChatStore().tryInit();
          }
          // console.log('check', this.info)
          this.lastCheckTime = Number(Date.now());
          return true;
        } catch (e: any) {
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
        const resp = await api.post('api/v1/user-signup', {
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
        return (err as any).response?.data?.message || '错误';
      }
    },

    logout() {
      // 将 accessToken 从 localStorage 中删除
      localStorage.removeItem('accessToken')
      this.info.id = ''
      // 更新 state 中的 accessToken
      this._accessToken = ''
    },

    async emojiAdd(attachmentId: string) {
      const user = useUserStore();
      const resp = await api.post('api/v1/user-emoji-add', { attachmentId }, {
        headers: { 'Authorization': user.token }
      });
      this.emojiCount += 1;
      return resp;
    },

    async emojiDelete(id: string) {
      const user = useUserStore();
      const resp = await api.post('api/v1/user-emoji-delete', { id }, {
        headers: { 'Authorization': user.token }
      });
      return resp;
    },

    async emojiList(): Promise<AxiosResponse<{ items: UserEmojiModel[] }, any>> {
      const user = useUserStore();
      const resp = await api.get('api/v1/user-emoji-list', {
        headers: { 'Authorization': user.token }
      });
      return resp;
    },

    // 满足任意一个即可，这个read是啥意思我也忘了
    checkPerm(...keys: Array<PermCheckKey>) {
      for (let key of keys) {
        if (this.permSysMap[key] === PermResult.ALLOWED) {
          return true;
        }
      }
    },

  },
})
