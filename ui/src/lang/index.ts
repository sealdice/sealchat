import { createI18n } from 'vue-i18n'
import LocaleZhcn from './locales/zh-cn.json'
import LocaleEn from './locales/en.json'
import LocaleJa from './locales/ja.json'
import { ref } from 'vue'

export const curLocale = ref('zh-cn')

export const i18n = createI18n({
  locale: 'zh-cn',
  legacy: false,
  messages: {
    'zh-cn': LocaleZhcn,
    'en': LocaleEn,
    'ja': LocaleJa,
  }
})

export function setLocale(locale: string) {
  if (locale.startsWith('zh-')) locale = 'zh-cn';
  else if (locale.startsWith('en-')) locale = 'en';
  else if (locale.startsWith('ja-')) locale = 'ja';
  else if (!['zh-cn', 'en', 'ja'].includes(locale)) return false;
  // i18n.global.setLocaleMessage(locale, LocaleZhcn)
  i18n.global.locale.value = locale as any;
  document.documentElement.lang = locale;
  localStorage.setItem('locale', locale);
  return true;
}

export function setLocaleByNavigator() {
  for (let lang of navigator.languages) {
    // 找到适配语言就中止
    if (setLocale(lang.toLowerCase())) break;
  }
}

export function setLocaleByNavigatorWithStorage() {
  if (setLocale(localStorage.getItem('locale') || '')) return;
  for (let lang of navigator.languages) {
    // 找到适配语言就中止
    if (setLocale(lang.toLowerCase())) break;
  }
}

(globalThis as any).setLocale = setLocale;
// console.log('setLocale', setLocale)
