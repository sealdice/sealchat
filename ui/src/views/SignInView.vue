<script setup lang="ts">
import router from '@/router';
import { defineComponent, ref, reactive } from 'vue';
import type { FormRules, FormItemRule, FormItemInst, FormInst } from 'naive-ui';
import { useMessage } from 'naive-ui';
import { useUserStore } from '@/stores/user';

const message = useMessage()

const formRef = ref<FormInst | null>(null)

interface ModelType {
  account: string;
  password: string | null
  reenteredPassword: string | null
}

const model = ref<ModelType>({
  account: '',
  password: null,
  reenteredPassword: null
})

function validatePasswordStartWith(
  rule: FormItemRule,
  value: string
): boolean {
  return (
    !!model.value.password &&
    model.value.password.startsWith(value) &&
    model.value.password.length >= value.length
  )
}

function validatePasswordSame(rule: FormItemRule, value: string): boolean {
  return value === model.value.password
}

const handleValidateButtonClick = async (e: MouseEvent) => {
  e.preventDefault()
  formRef.value?.validate(async (errors) => {
    if (!errors) {
      try {
        const resp = await userStore.signIn(model.value.account, model.value.password || '');
        const ret = resp.data;
        message.success('验证成功，3秒后返回首页')
        if (ret.token) {
          setTimeout(() => {
            router.replace({ name: 'home' })
          }, 3000);
        }
      } catch (err) {
        message.error('登录失败: ' + ((err as any).data?.message || '密码错误'))
      }
    } else {
      console.log(errors)
      message.error('验证失败')
    }
  })
}


const rules: FormRules = {
  password: [
    {
      required: true,
      message: '请输入密码'
    }
  ],
  // reenteredPassword: [
  //   {
  //     required: true,
  //     message: '请再次输入密码',
  //     trigger: ['input', 'blur']
  //   },
  //   {
  //     validator: validatePasswordStartWith,
  //     message: '两次密码输入不一致',
  //     trigger: 'input'
  //   },
  //   {
  //     validator: validatePasswordSame,
  //     message: '两次密码输入不一致',
  //     trigger: ['blur', 'password-input']
  //   }
  // ]
}

const rPasswordFormItemRef = ref<FormItemInst | null>(null)

const handlePasswordInput = () => {
  if (model.value.reenteredPassword) {
    rPasswordFormItemRef.value?.validate({ trigger: 'password-input' })
  }
}

const userStore = useUserStore();
</script>

<template>
  <div class="flex h-full w-full justify-center items-center">
    <div class="w-[50%] flex items-center justify-center flex-col" style="min-width: 20rem;">
      <h2 class="font-bold text-xl mb-8">摸鱼中心</h2>

      <n-form ref="formRef" :model="model" :rules="rules" class="w-full px-8 max-w-md">
        <n-form-item path="account" label="帐号">
          <n-input v-model:value="model.account" @keydown.enter.prevent />
        </n-form-item>

        <n-form-item path="password" label="密码">
          <n-input v-model:value="model.password" type="password" @input="handlePasswordInput" @keydown.enter.prevent />
        </n-form-item>

        <!-- <n-form-item ref="rPasswordFormItemRef" first path="reenteredPassword" label="重复密码">
          <n-input v-model:value="model.reenteredPassword" :disabled="!model.password" type="password"
            @keydown.enter.prevent />
        </n-form-item> -->

        <n-row :gutter="[0, 24]">
          <n-col :span="24">
            <div class=" flex justify-between">
              <router-link :to="{ name: 'user-signup' }">
                <n-button type="text">注册</n-button>
              </router-link>

              <n-button :disabled="model.account === ''" round type="primary" @click="handleValidateButtonClick">
                登录
              </n-button>
            </div>
          </n-col>
        </n-row>
      </n-form>

    </div>
  </div>
</template>
  
<style>
.sign-bg {
  background-size: cover;
  background-position: center;
}
</style>
