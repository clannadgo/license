<template>
  <div class="language-switch">
    <el-dropdown 
      trigger="click" 
      placement="bottom-end"
      @command="handleLanguageChange"
    >
      <span class="language-trigger">
        <span class="flag">{{ currentFlag }}</span>
        <span class="language-name">{{ currentLanguageName }}</span>
        <el-icon><arrow-down /></el-icon>
      </span>
      
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item 
            v-for="lang in availableLanguages" 
            :key="lang.value"
            :command="lang.value"
            :class="{ active: currentLanguage === lang.value }"
          >
            <span class="flag">{{ lang.flag }}</span>
            <span class="language-name">{{ lang.label }}</span>
            <el-icon v-if="currentLanguage === lang.value" class="check-icon">
              <check />
            </el-icon>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useI18nStore } from '../stores/i18n'
import { ArrowDown, Check } from '@element-plus/icons-vue'

const { locale } = useI18n()
const i18nStore = useI18nStore()

// 计算属性
const currentLanguage = computed(() => i18nStore.currentLanguage)
const availableLanguages = computed(() => i18nStore.availableLanguages)

const currentFlag = computed(() => {
  const lang = availableLanguages.value.find(l => l.value === currentLanguage.value)
  return lang ? lang.flag : '🌐'
})

const currentLanguageName = computed(() => {
  const lang = availableLanguages.value.find(l => l.value === currentLanguage.value)
  return lang ? lang.label : 'Language'
})

// 切换语言
const handleLanguageChange = (lang: string) => {
  i18nStore.switchLanguage(lang)
  locale.value = lang
  
  // 动态更新Element Plus的语言包
  const elementLocaleMap = {
    zhCN: () => import('element-plus/dist/locale/zh-cn.mjs'),
    enUS: () => import('element-plus/dist/locale/en.mjs'),
    jaJP: () => import('element-plus/dist/locale/ja.mjs')
  }
  
  if (elementLocaleMap[lang as keyof typeof elementLocaleMap]) {
    elementLocaleMap[lang as keyof typeof elementLocaleMap]().then(module => {
      // 更新Element Plus的locale配置
      const elementPlus = (window as any).$ELEMENT
      if (elementPlus) {
        elementPlus.locale = module.default
      }
    })
  }
}
</script>

<style scoped>
.language-switch {
  display: flex;
  align-items: center;
}

.language-trigger {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
  color: #606266;
}

.language-trigger:hover {
  background-color: #f5f7fa;
}

.flag {
  margin-right: 6px;
  font-size: 16px;
}

.language-name {
  margin-right: 6px;
  font-size: 14px;
}

.el-dropdown-menu {
  min-width: 120px;
}

.el-dropdown-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
}

.el-dropdown-item.active {
  background-color: #ecf5ff;
  color: #409eff;
}

.check-icon {
  margin-left: auto;
  color: #409eff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .language-name {
    display: none;
  }
  
  .language-trigger {
    padding: 6px 8px;
  }
}
</style>