import axiosFactory, { Axios } from "axios"
const axios = axiosFactory.create()

// export const urlBase = '//' + window.location.hostname + ":" + 3212;
// export const urlBase = '//' + window.location.host + '/';

export const urlBase = import.meta.env.MODE === 'development'
  ? '//' + window.location.hostname + ":" + 3212
  : '//' + window.location.host;

console.log('mode', import.meta.env.MODE)

export const fileSizeLimit = 2 * 1024 * 1024

export const api = axiosFactory.create({
  baseURL: urlBase + '/',
  //   withCredentials: true,
  timeout: 10000,
  maxRedirects: 3,
  transitional: {
    silentJSONParsing: false
  },
  responseType: 'json',
})
