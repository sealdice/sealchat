import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)

import './assets/main.css'
import 'dayjs/locale/zh-cn'
import { useUserStore } from './stores/user'

router.beforeEach(async (to, from, next) => {
  if (to.name === 'user-signin' || to.name === 'user-signup') {
    return next();
  }

  const user = useUserStore();
  const r = await user.checkUserSession();
  if (r) {
    return next();
  }

  next({ name: 'user-signin' })
  // window.location.href = '//' + window.location.hostname + ":4455/login";
  return;
})

// 这几句详见 https://www.naiveui.com/zh-CN/os-theme/docs/style-conflict
const meta = document.createElement('meta')
meta.name = 'naive-ui-style'
document.head.appendChild(meta)

app.mount('#app')
