<template>
  <div class="min-h-screen bg-gray-100 p-8">
    <div class="max-w-4xl mx-auto space-y-8">
      <!-- 标题 -->
      <h1 class="text-3xl font-bold text-gray-800 text-center mb-8">微信公众号「图片/文字」采集工具</h1>

      <!-- 功能区1：URL采集 -->
      <div class="bg-white rounded-lg shadow-md p-6">
        <h2 class="text-xl font-semibold text-gray-700 mb-4">URL图片采集</h2>
        <div class="space-y-4">
          <!-- URL输入区域 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">URL地址列表（每行一个，最多50个）</label>
            <textarea
              v-model="urls"
              rows="4"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              :placeholder="urlInputPlaceholder"
            ></textarea>
          </div>

          <!-- 文件选择和保存路径区域 -->
          <div class="space-y-3">
            <div class="flex items-center space-x-4">
              <button
                @click="selectFile"
                class="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
              >
                选择文件
              </button>
              <div class="flex-1">
                <div v-if="selectedFilePath" class="text-sm text-gray-600 bg-gray-50 p-2 rounded border border-gray-200">
                  <span class="font-medium">已选择文件：</span>
                  <span class="text-blue-600 break-all">{{ selectedFilePath }}</span>
                </div>
                <div v-else class="text-sm text-gray-400 italic">
                  未选择文件
                </div>
              </div>
            </div>

            <div class="flex items-center space-x-4">
              <button
                @click="selectSavePath"
                class="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
              >
                选择保存路径
              </button>
              <div class="flex-1">
                <div v-if="savePath" class="text-sm text-gray-600 bg-gray-50 p-2 rounded border border-gray-200">
                  <span class="font-medium">保存路径：</span>
                  <span class="text-blue-600 break-all">{{ savePath }}</span>
                </div>
                <div v-else class="text-sm text-gray-400 italic">
                  未选择保存路径
                </div>
              </div>
            </div>
          </div>

          <!-- 超时时间和开始采集按钮 -->
          <div class="flex items-center space-x-4">
            <div class="flex items-center">
              <label class="text-sm text-gray-700 mr-2">超时时间（秒）：</label>
              <input
                type="number"
                v-model="timeout"
                min="1"
                max="50"
                @input="handleTimeoutInput"
                class="w-20 px-2 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <button
              @click="startCrawling"
              :disabled="isCrawling"
              class="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {{ isCrawling ? '采集中...' : '开始采集' }}
            </button>
          </div>

          <!-- 进度条 -->
          <div v-if="isCrawling" class="w-full bg-gray-200 rounded-full h-2.5">
            <div
              class="bg-blue-500 h-2.5 rounded-full transition-all duration-300"
              :style="{ width: `${progress}%` }"
            ></div>
          </div>
        </div>
      </div>

      <!-- 功能区2和3：图片裁剪和打乱 -->
      <div class="grid grid-cols-2 gap-6">
        <!-- 功能区2：图片裁剪 -->
        <div class="bg-white rounded-lg shadow-md p-6">
          <h2 class="text-xl font-semibold text-gray-700 mb-4">图片裁剪</h2>
          <div class="space-y-4">
            <div class="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-4">
              <div class="flex">
                <div class="flex-shrink-0">
                  <svg class="h-5 w-5 text-yellow-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div class="ml-3">
                  <p class="text-sm text-yellow-700 font-medium">
                    当前只裁剪图片底部区域
                  </p>
                </div>
              </div>
            </div>
            <div class="flex items-center space-x-4">
              <div class="flex items-center">
                <label class="text-sm text-gray-700 mr-2">裁剪高度（像素）：</label>
                <input
                  type="number"
                  v-model="cropHeight"
                  min="1"
                  max="500"
                  @input="handleCropHeightInput"
                  class="w-20 px-2 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <button
                @click="startCropping"
                :disabled="isCropping"
                class="px-6 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50"
              >
                {{ isCropping ? '裁剪中...' : '开始裁剪' }}
              </button>
            </div>
          </div>
        </div>

        <!-- 功能区3：图片打乱 -->
        <div class="bg-white rounded-lg shadow-md p-6">
          <h2 class="text-xl font-semibold text-gray-700 mb-4">图片打乱</h2>
          <div class="flex flex-col items-center justify-center h-full">
            <p class="text-sm text-gray-600 mb-4">随机打乱图片顺序</p>
            <button
              @click="startShuffling"
              :disabled="isShuffling"
              class="px-6 py-2 bg-purple-500 text-white rounded-md hover:bg-purple-600 focus:outline-none focus:ring-2 focus:ring-purple-500 disabled:opacity-50"
            >
              {{ isShuffling ? '打乱中...' : '开始打乱' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {GetPreferenceInfo} from "../wailsjs/go/handlers/User.js"
import {LogPrint} from "../wailsjs/runtime/runtime.js"

// 状态变量
const urls = ref('')
const selectedFilePath = ref('')
const savePath = ref('')
const timeout = ref(15)
const progress = ref(0)
const cropHeight = ref(20)

// 操作状态
const isCrawling = ref(false)
const isCropping = ref(false)
const isShuffling = ref(false)

const urlInputPlaceholder = '请输入微信“小绿书”URL地址，每行一个。例如： \n' +
    'https://mp.weixin.qq.com/s/oCpFfUCtIYd9oAGsuDi6BA\n' +
    'https://mp.weixin.qq.com/s/hQf0N8P4vaaCaxt8OFzwfw\n'

onMounted(() => {
  // 获取用户偏好设置
  GetPreferenceInfo().then((res) => {
    if (res) {
      // 设置默认值
      timeout.value = res.download_timeout || 15
      cropHeight.value = res.crop_img_bottom_pixel || 20
      savePath.value = res.save_img_path || ''
      console.log(res)
    }
  })
})

// 输入处理函数
const handleTimeoutInput = (event) => {
  const value = event.target.value
  if (value === '') {
    timeout.value = 15 // 默认值
    return
  }
  const num = parseInt(value)
  if (isNaN(num)) {
    timeout.value = 15 // 默认值
    return
  }
  if (num < 1) {
    timeout.value = 1
  } else if (num > 50) {
    timeout.value = 50
  } else {
    timeout.value = num
  }
}

const handleCropHeightInput = (event) => {
  const value = event.target.value
  if (value === '') {
    cropHeight.value = 20 // 默认值
    return
  }
  const num = parseInt(value)
  if (isNaN(num)) {
    cropHeight.value = 20 // 默认值
    return
  }
  if (num < 1) {
    cropHeight.value = 1
  } else if (num > 500) {
    cropHeight.value = 500
  } else {
    cropHeight.value = num
  }
}

// 方法
const selectFile = () => {
  // TODO: 实现文件选择逻辑
}

const selectSavePath = () => {
  // TODO: 实现保存路径选择逻辑
}

const startCrawling = () => {
  isCrawling.value = true
  // TODO: 实现图片采集逻辑
  // 模拟进度条更新
  const interval = setInterval(() => {
    if (progress.value >= 100) {
      clearInterval(interval)
      isCrawling.value = false
      progress.value = 0
    } else {
      progress.value += 10
    }
  }, timeout.value * 1000)

}

const startCropping = () => {
  isCropping.value = true
  // TODO: 实现图片裁剪逻辑
}

const startShuffling = () => {
  isShuffling.value = true
  // TODO: 实现图片打乱逻辑
}
</script>

<style>
/* 可以添加自定义样式 */
</style>
