import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import en from 'element-plus/dist/locale/en.mjs'
import ja from 'element-plus/dist/locale/ja.mjs'

import App from './App.vue'
import router from './router'
import i18n from './locales'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(i18n)

// 获取当前语言设置
const getPreferredLanguage = () => {
  const savedLang = localStorage.getItem('preferredLanguage')
  if (savedLang && ['zhCN', 'enUS', 'jaJP'].includes(savedLang)) {
    return savedLang
  }
  
  // 根据浏览器语言自动选择
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.includes('zh')) {
    return 'zhCN'
  } else if (browserLang.includes('ja')) {
    return 'jaJP'
  } else {
    return 'enUS'
  }
}

// 根据当前语言设置Element Plus的语言
const currentLanguage = getPreferredLanguage()
const elementLocaleMap = {
  zhCN: zhCn,
  enUS: en,
  jaJP: ja
}

app.use(ElementPlus, {
  locale: elementLocaleMap[currentLanguage] || zhCn,
})

// 将Element Plus配置暴露到全局，以便动态切换语言
app.config.globalProperties.$ELEMENT = {
  locale: elementLocaleMap[currentLanguage] || zhCn
}

app.mount('#app')
