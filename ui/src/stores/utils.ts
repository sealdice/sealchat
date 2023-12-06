import { defineStore } from "pinia"
import type { ServerConfig, UserInfo } from "@/types";
import { Howl, Howler } from 'howler';

import axiosFactory from "axios"
import { cloneDeep } from "lodash-es";

import type { AxiosResponse } from "axios";
import { api } from "./_config";
import { useChatStore } from "./chat";
import { useUserStore } from "./user";

interface SoundItem {
  sound: Howl;
  time: number;
  playing: boolean;
}

interface UtilsState {
  config: ServerConfig | null;
  botCommands: { [key: string]: any };
  sounds: Map<string, SoundItem>;
  soundsTimer: any;
}

export const useUtilsStore = defineStore({
  id: 'utils',

  state: (): UtilsState => ({
    config: null,
    botCommands: {} as any,
    sounds: new Map<string, SoundItem>(),
    soundsTimer: null,
  }),

  getters: {
    fileSizeLimit: (state) => {
      if (state.config) {
        return state.config.imageSizeLimit * 1024;
      }
      return 2 * 1024 * 1024
    }
  },

  actions: {
    async soundsTryInit() {
      if (this.soundsTimer) return;
      this.soundsTimer = setInterval(() => {
        for (let [k,v] of this.sounds.entries()) {
          v.time = v.sound.seek();
        }
      }, 1000);
    },

    async configGet() {
      const user = useUserStore();
      const resp = await api.get('api/v1/config', {
        headers: { 'Authorization': user.token }
      })
      this.config = resp.data as ServerConfig;
      return resp
    },

    async botTokenList() {
      const user = useUserStore();
      const resp = await api.get('api/v1/bot_token/list', {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async botTokenAdd(name: string) {
      const user = useUserStore();
      const resp = await api.post('api/v1/bot_token/add', { name }, {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async botTokenDelete(id: string) {
      const user = useUserStore();
      const resp = await api.delete(`api/v1/bot_token/${id}`, {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async configSet(data: ServerConfig) {
      const user = useUserStore();
      const resp = await api.put('api/v1/config', data, {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async adminUserList() {
      const user = useUserStore();
      const resp = await api.get('api/v1/admin/user/list', {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async userResetPassword(id: string) {
      const user = useUserStore();
      const resp = await api.put(`api/v1/admin/user/reset_password/${id}`, null, {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async userEnable(id: string) {
      const user = useUserStore();
      const resp = await api.put(`api/v1/admin/user/enable/${id}`, null, {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async userDisable(id: string) {
      const user = useUserStore();
      const resp = await api.put(`api/v1/admin/user/disable/${id}`, null, {
        headers: { 'Authorization': user.token }
      })
      return resp
    },

    async commandsRefresh() {
      const user = useUserStore();
      const resp = await api.get(`api/v1/commands`, {
        headers: { 'Authorization': user.token }
      })
      this.botCommands = resp.data as any;
      return resp
    },
  },
})
