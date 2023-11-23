<script lang="tsx" setup>
import { urlBase, fileSizeLimit } from '@/stores/_config';
import { computed, onMounted, ref } from 'vue';
import FileUpload from 'vue-upload-component'
import { useUserStore } from '@/stores/user';
import { filesize } from "filesize";
import { useMessage } from 'naive-ui';
import { useChatStore } from '@/stores/chat';

const user = useUserStore();
const chat = useChatStore();
const message = useMessage()

const files = ref<any[]>([])
const uploadRef = ref<any>(null)
const putAction = urlBase + '/api/v1/upload'
const headers = computed(() => {
  return {
    Authorization: `${user.token}`
  }
})

const inputFile = (newFile: any, oldFile: any) => {
  if (newFile && oldFile && !newFile.active && oldFile.active) {
    // 获得相应数据
    console.log('response', newFile.response.files)
    dialogVisible.value = false;
    files.value = [];

    if (newFile.xhr) {
      if (newFile.xhr.status === 200) {
        // 上传成功
        chat.messageCreate(`<img src="id:${newFile.response.files}" />`)
        console.log('success')
      } else {
        // 上传失败
        message.error('上传失败')
      }
    }
  }
}

const inputFilter = function (newFile: any, oldFile: any, prevent: any) {
  if (newFile && !oldFile) {
    // 过滤不是图片后缀的文件
    if (!/\.(jpeg|jpe|jpg|gif|png|webp)$/i.test(newFile.name)) {
      return prevent()
    } else {
      // 创建 blob 字段 用于图片预览
      newFile.blob = ''
      let URL = window.URL || window.webkitURL
      if (URL && URL.createObjectURL) {
        newFile.blob = URL.createObjectURL(newFile.file)
      }

      newFile.thumb = ''
      if (newFile.blob && newFile.type.substr(0, 6) === 'image/') {
        newFile.thumb = newFile.blob
      }
      dialogVisible.value = true;
    }
  }
}

document.addEventListener('paste', function (event) {
  // 获取粘贴事件中的剪贴板数据
  const items = (event.clipboardData || (event as any).originalEvent.clipboardData).items;

  // 遍历剪贴板中的每一项
  for (let i = 0; i < items.length; i++) {
    // 如果是文件类型
    if (items[i].kind === 'file') {
      const file = items[i].getAsFile();

      // 模拟文件上传
      uploadRef.value.add(file)
      break;
    }
  }
});

const dialogVisible = ref(false)
</script>

<template>
  <div class="  absolute">
    <file-upload ref="uploadRef" v-model="files" :post-action="putAction" @input-file="inputFile"
      @input-filter="inputFilter" :headers="headers" :size="fileSizeLimit">
    </file-upload>
  </div>

  <n-modal v-model:show="dialogVisible" preset="dialog" title="上传文件" :auto-focus="false">
    <div class="p-4">
      <ul class="mb-4">
        <li v-for="file in files" :key="file.name" class="text-gray-600">
          <img class="td-image-thumb" v-if="file.thumb" :src="file.thumb" />
          <div class="text-center">
            <span>{{ file.name }}</span>
            <span> - </span>
            <span :class="file.size > fileSizeLimit ? 'text-red-500' : ''">{{ filesize(file.size) }}</span>
          </div>
          <!-- {{ file.name }} - Error: {{ file.error }}, Success: {{ file.success }} -->
        </li>
      </ul>
    </div>

    <template #action>
      <n-button autofocus v-show="!uploadRef || !uploadRef.active" @click.prevent="uploadRef.active = true" type="primary"
        :disabled="Boolean(files.length) && files[0].size > fileSizeLimit">开始上传</n-button>
      <n-button v-show="uploadRef && uploadRef.active" @click.prevent="uploadRef.active = false"
        type="button">停止上传</n-button>
      <n-button @click="dialogVisible = false">取消</n-button>
    </template>
  </n-modal>
</template>
