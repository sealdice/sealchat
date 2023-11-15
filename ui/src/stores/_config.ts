import axiosFactory, { Axios } from "axios"
const axios = axiosFactory.create()

export const urlBase = '//' + window.location.hostname + ":" + 3212;

export const api = axiosFactory.create({
  // basePath: 'http://127.0.0.1:4005',
  baseURL: urlBase + '/',
//   withCredentials: true,
  timeout: 10000,
  maxRedirects: 3,
  transitional: {
    silentJSONParsing: false
  },
  responseType: 'json',
})
