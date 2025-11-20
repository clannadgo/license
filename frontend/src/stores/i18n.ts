import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useI18nStore = defineStore('i18n', () => {
  // 当前语言
  const currentLanguage = ref('zhCN')
  
  // 可用语言列表
  const availableLanguages = [
    { value: 'zhCN', label: '中文', flag: '🇨🇳' },
    { value: 'enUS', label: 'English', flag: '🇺🇸' },
    { value: 'jaJP', label: '日本語', flag: '🇯🇵' }
  ]
  
  // 切换语言
  const switchLanguage = (lang: string) => {
    if (availableLanguages.some(l => l.value === lang)) {
      currentLanguage.value = lang
      
      // 保存到localStorage
      localStorage.setItem('preferredLanguage', lang)
    }
  }
  
  // 初始化语言设置
  const initLanguage = () => {
    const savedLang = localStorage.getItem('preferredLanguage')
    if (savedLang && availableLanguages.some(l => l.value === savedLang)) {
      currentLanguage.value = savedLang
    } else {
      // 根据浏览器语言自动选择
      const browserLang = navigator.language.toLowerCase()
      if (browserLang.includes('zh')) {
        currentLanguage.value = 'zhCN'
      } else if (browserLang.includes('ja')) {
        currentLanguage.value = 'jaJP'
      } else {
        currentLanguage.value = 'enUS'
      }
    }
  }
  
  return {
    currentLanguage,
    availableLanguages,
    switchLanguage,
    initLanguage
  }
})