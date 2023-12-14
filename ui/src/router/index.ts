import { createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import UserSigninVue from '@/views/user/sign-in-view.vue'
import UserSignupVue from '@/views/user/sign-up-view.vue'
import UserPasswordResetView from '@/views/user/password-reset-view.vue'


const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/user/signin',
      name: 'user-signin',
      component: UserSigninVue
    },
    {
      path: '/user/signup',
      name: 'user-signup',
      component: UserSignupVue
    },
    {
      path: '/user/password-reset',
      name: 'user-password-reset',
      component: UserPasswordResetView
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/AboutView.vue')
    }
  ]
})

export default router
