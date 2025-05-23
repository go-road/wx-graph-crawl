import {createApp} from 'vue'
import ElementPlus from 'element-plus'
import App from './App.vue'
import './third-party.css'

import router from "./router"

const app = createApp(App)

app.use(ElementPlus)
app.use(router)
app.mount('#app')

app.config.errorHandler = (err) => {
    console.error('捕捉到了 wx-graph-crawl 错误：', err)
}