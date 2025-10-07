<template>
  <div class="min-h-screen bg-gray-100 p-8">
    <div class="max-w-4xl mx-auto space-y-8">
      <!-- 标题 -->
      <div class="text-center space-y-2">
        <h1 class="text-3xl font-bold text-gray-800">微信公众号「图片/文字」采集工具</h1>
      </div>

      <!-- 功能区1：URL采集 -->
      <div class="bg-white rounded-lg shadow-md p-6">
        <h2 class="text-xl font-semibold text-gray-700 mb-4">URL 图片采集</h2>
        <div class="space-y-4">
          <!-- URL输入区域 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">URL 地址列表（一行一个 URL，最多 {{configureInit.maxDownloadURLCount}} 个）</label>
            <textarea
                v-model="urls"
                rows="4"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm placeholder:text-sm resize-none"
                :placeholder="urlInputPlaceholder"
            ></textarea>
          </div>

          <!-- 文件选择和保存路径区域 -->
          <div class="space-y-3">
            <div class="flex items-start space-x-4">
              <div class="flex-shrink-0">
                <button
                    @click="selectFile"
                    class="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  选择文件
                </button>
              </div>
              <div class="flex-1">
                <div class="text-sm text-gray-600 mb-2">
                  <p class="font-medium">文件导入说明：</p>
                  <ol class="list-decimal list-inside space-y-1 mt-1">
                    <li>也可通过选择文件来自动输入 URL 地址</li>
                    <li>仅支持 .txt 文件，且一行一个 URL 地址</li>
                  </ol>
                </div>
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
                选择图片保存路径
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
              <label class="text-sm text-gray-700 mr-2">图片下载超时时间（秒）：</label>
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
          <div v-if="isCrawling" class="w-full h-2.5">
            <el-progress
                :percentage="progress"
                :stroke-width="12"
                status="success"
                striped
                striped-flow
                :duration="10"
            >
            </el-progress>
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
import { ref, watch, onMounted, onUpdated } from 'vue'
import { ElNotification, ElMessage } from 'element-plus'
import {GetPreferenceInfo, SetPreferenceInfo} from "wailsjs/go/handlers/UserHandler.js"
import {SelectFile, SelectDirectory} from "wailsjs/go/handlers/FileHandler.js"
import {Crawling, Cropping, Shuffling} from "wailsjs/go/handlers/ImageHandler.js"

const configureInit = {
  maxDownloadURLCount: 500, // 最大下载URL数量
  maxNumImageSplitDirectory: 5, // 当一个目录中的图片超过多少张时，开始拆分目录
  crop: {
    defaultValue: 65, // 默认裁剪高度 （像素）
    minValue: 1,
    maxValue: 500,
  },
  downloadTimeout: {
    defaultValue: 5, // 默认下载超时时间（秒）
    minValue: 1,
    maxValue: 500,
  },
}

// 状态变量
const urls = ref('') // URL列表
const selectedFilePath = ref('') // 已选择的文件路径
const savePath = ref('') // 图片保存路径
const timeout = ref(configureInit.downloadTimeout.defaultValue) // 下载超时时间
const progress = ref(0)
const cropHeight = ref(configureInit.crop.defaultValue) // 裁剪高度

// 操作状态
const isCrawling = ref(false) // 是否正在采集
const isCropping = ref(false) // 是否正在裁剪
const isShuffling = ref(false) // 是否正在打乱

const urlInputPlaceholder = '请输入微信"小绿书" URL 地址，一行一个，' +
    '例如：\n' +
    'https://mp.weixin.qq.com/s/oCpFfUCtIYd9oAGsuDi6BA\n' +
    'https://mp.weixin.qq.com/s/hQf0N8P4vaaCaxt8OFzwfw\n'

onMounted(() => {
  // 获取用户偏好设置
  setPreferenceInfo()
})

watch([savePath, timeout, cropHeight], () => {
  // 保存用户偏好设置
  goSavePreferenceInfo()
})

const setPreferenceInfo = async () => {
  try {
    const res = await GetPreferenceInfo()
    if (res) {
      timeout.value = res.download_timeout || configureInit.downloadTimeout.defaultValue
      cropHeight.value = res.crop_img_bottom_pixel || configureInit.crop.defaultValue
      savePath.value = res.save_img_path || ''
    }
  } catch (e) {
    console.error("获取用户偏好设置失败", e)
    ElMessage.error({
      message: '获取用户偏好设置失败，请重试。错误原因：' + e,
      showClose: true,
      grouping: true,
    })
  }
}

const goSavePreferenceInfo = async () => {
  try {
    await SetPreferenceInfo({
      save_img_path: savePath.value,
      download_timeout: timeout.value,
      crop_img_bottom_pixel: cropHeight.value,
    })
  } catch (e) {
    console.error("保存用户偏好设置失败", e)
  }
}

// 输入处理函数
const handleTimeoutInput = (event) => {
  const value = event.target.value
  if (value === '') {
    timeout.value = configureInit.downloadTimeout.defaultValue // 默认值
    return
  }
  const num = parseInt(value)
  if (isNaN(num)) {
    timeout.value = configureInit.downloadTimeout.defaultValue // 默认值
    return
  }
  if (num < configureInit.downloadTimeout.minValue) {
    timeout.value = configureInit.downloadTimeout.minValue
  } else if (num > configureInit.downloadTimeout.maxValue) {
    timeout.value = configureInit.downloadTimeout.maxValue
  } else {
    timeout.value = num
  }
}

const handleCropHeightInput = (event) => {
  const value = event.target.value
  if (value === '') {
    cropHeight.value = configureInit.crop.defaultValue // 默认值
    return
  }
  const num = parseInt(value)
  if (isNaN(num)) {
    cropHeight.value = configureInit.crop.defaultValue // 默认值
    return
  }
  if (num < configureInit.crop.minValue) {
    cropHeight.value = configureInit.crop.minValue
  } else if (num > configureInit.crop.maxValue) {
    cropHeight.value = configureInit.crop.maxValue
  } else {
    cropHeight.value = num
  }
}

const selectFile = async () => {
  try {
    const { file_path: filePath, valid_urls: validUrls } = await SelectFile()
    if (filePath) {
      selectedFilePath.value = filePath
      if (validUrls && validUrls.length > 0) {
        urls.value = validUrls.join('\n')
      } else {
        urls.value = ''
        ElNotification.warning({
          title: '文件内容为空',
          message: '导入的文件中没有有效的小绿书URL地址',
        })
      }
    }
  } catch (e) {
    console.error("选择文件失败", e)
    ElMessage.error({
      message: '选择文件失败，请重试。错误原因：' + e,
      showClose: true,
      grouping: true,
    })
  }
}

const selectSavePath = async () => {
  try {
    const dirPath = await SelectDirectory();
    if (dirPath) {
      savePath.value = dirPath
    }
  } catch (e) {
    console.error("选择保存路径失败", e)
    ElMessage.error({
      message: '选择保存路径失败，请重试。错误原因：' + e,
      showClose: true,
      grouping: true,
    })
  }
}

const startCrawling = async () => {
  isCrawling.value = true

  // 验证URL列表
  const urlList = urls.value.trim().split('\n').filter(url => url.trim())
  if (urlList.length === 0) {
    ElNotification.warning({
      title: 'URL 列表为空',
      message: '请先输入需要采集的 URL 地址',
    })
    isCrawling.value = false
    return
  }

  // 验证保存路径
  if (!savePath.value) {
    ElNotification.warning({
      title: '保存路径未设置',
      message: '请先选择图片保存路径',
    })
    isCrawling.value = false
    return
  }

  // 验证URL数量
  if (urlList.length > configureInit.maxDownloadURLCount) {
    ElNotification.warning({
      title: 'URL数量超限',
      message: `一次最多只能采集${configureInit.maxDownloadURLCount}个URL地址`,
    })
    isCrawling.value = false
    return
  }

  try {
    progress.value = 30
    const crawlingResult = await Crawling({
      img_save_path: savePath.value,
      img_urls: urlList,
      timeout_seconds: timeout.value,
    })
    progress.value = 100
    console.log("采集完成", crawlingResult)
    let noticeMsg = '累计耗时：<span class="text-blue-600 font-medium">' + crawlingResult.cast_time_str + '</span>\n' +
        '成功采集了 <span class="text-green-600 font-medium">' + crawlingResult.crawl_url_count + '</span> 个 URL 地址，\n' +
        '总共下载了 <span class="text-purple-600 font-medium bg-purple-50 px-1 rounded">' + crawlingResult.crawl_img_count + '</span> 张图片，\n' +
        '全部文案内容保存于 <span class="text-gray-600 font-medium">' + crawlingResult.text_content_save_path + '</span> 文件中，\n' +
        '单个文件的文案内容保存于 <span class="text-gray-600 font-medium">' + crawlingResult.text_content_save_dir + '</span> 目录下。'
    if (crawlingResult.err_content !== '') {
      noticeMsg += '\n\n<span class="text-red-600 font-medium">出现了以下错误：</span>\n\n' +
          '<span class="text-red-500">' + crawlingResult.err_content + '</span>'
    }
    ElNotification.success({
      title: '恭喜🎉采集完成！',
      message: noticeMsg,
      duration: 10000,
      showClose: true,
      dangerouslyUseHTMLString: true,
    })
  } catch (e) {
    console.error("采集失败", e)
    ElMessage.error({
      message: '采集失败，请重试。错误原因：' + e,
      showClose: true,
      grouping: true,
    })
  } finally {
    isCrawling.value = false
    progress.value = 0
  }

}

const startCropping = async () => {
  try {
    isCropping.value = true

    if (!savePath.value) {
      ElNotification.warning({
        title: '图片路径未设置',
        message: '请先点击【选择图片保存路径】按钮，选择需要裁剪的图片目录',
      })
      return
    }

    const {
      crop_img_path: cropImgPath,
      crop_img_count: cropImgCount,
      err_content: errContent,
      cast_time_str: castTimeStr,
    } = await Cropping({
      img_save_path: savePath.value,
      bottom_pixel: cropHeight.value,
    })
    console.log("裁剪完成", cropImgPath, cropImgCount, errContent, castTimeStr)
    let noticeMsg = '累计耗时：<span class="text-blue-600 font-medium">' + castTimeStr + '</span>\n' +
        '裁剪了 <span class="text-green-600 font-medium">' + cropImgCount + '</span> 张图片，\n' +
        '裁剪图片目录： <span class="text-purple-600 font-medium bg-purple-50 px-1 rounded">' + cropImgPath + '</span>'
    if (errContent !== '') {
      noticeMsg += '\n\n<span class="text-red-600 font-medium">裁剪过程中，出现了以下错误：</span>\n\n' +
          '<span class="text-red-500">' + errContent + '</span>'
    }
    ElNotification.success({
      title: '恭喜🎉裁剪完成！',
      message: noticeMsg,
      duration: 10000,
      showClose: true,
      dangerouslyUseHTMLString: true,
    })
  } catch (e) {
    console.error("裁剪失败", e)
    ElMessage.error({
      message: '裁剪失败，请重试。错误原因：' + e,
      showClose: true,
      grouping: true,
    })
  } finally {
    isCropping.value = false
  }
}

const startShuffling = async () => {
  try {
    isShuffling.value = true
    if (!savePath.value) {
      ElNotification.warning({
        title: '图片路径未设置',
        message: '请先点击【选择图片保存路径】按钮，选择需要打乱的图片目录',
      })
      return
    }

    const shufflingResult = await Shuffling({
      img_save_path: savePath.value,
      max_num_image: configureInit.maxNumImageSplitDirectory
    })
    console.log("打乱完成", shufflingResult)
    let noticeMsg = '累计耗时：<span class="text-blue-600 font-medium">' + shufflingResult.cast_time_str + '</span>\n' +
        '打乱图片目录： <span class="text-purple-600 font-medium bg-purple-50 px-1 rounded">' + shufflingResult.shuffle_img_path + '</span>'
    ElNotification.success({
      title: '恭喜🎉打乱完成！',
      message: noticeMsg,
      duration: 10000,
      showClose: true,
      dangerouslyUseHTMLString: true,
    })
  } catch (e) {
    console.error("打乱失败", e)
    ElMessage.error({
      message: '打乱失败，请重试。错误原因：' + e,
      showClose: true,
      grouping: true,
    })
  } finally {
    isShuffling.value = false
  }
}

</script>

<style>
/* 可以添加自定义样式 */
</style>
