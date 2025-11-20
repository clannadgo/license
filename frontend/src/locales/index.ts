import { createI18n } from 'vue-i18n'

// 中文翻译
const zhCN = {
  common: {
    licenseManagement: 'License管理系统',
    addLicense: '新增License',
    testFingerprint: '测试生成机器码',
    refresh: '刷新数据',
    search: '搜索',
    searchPlaceholder: '请输入客户名称进行搜索',
    customer: '客户名称',
    fingerprint: '机器码',
    description: '描述',
    activatedAt: '激活时间',
    expiresAt: '过期时间',
    status: '状态',
    active: '已激活',
    expired: '已过期',
    operations: '操作',
    download: '下载',
    edit: '编辑',
    delete: '删除',
    cancel: '取消',
    confirm: '确定',
    update: '更新',
    previous: '上一页',
    next: '下一页',
    total: '总计',
    items: '条',
    pageSize: '条/页',
    goto: '前往',
    page: '页',
    enterCustomer: '请输入客户名称',
    enterFingerprint: '请输入机器码',
    enterDescription: '请输入描述信息',
    selectExpiresAt: '选择过期时间',
    editLicense: '编辑License',
    validityPeriod: '有效期',
    days: '天数',
    hours: '小时',
    minutes: '分钟',
    seconds: '秒',
    fingerprintFormatTip: '正确格式应为XXXXX-XXXXX-XXXXX-XXXXX-XXXXX（5组5位字母或数字）',
    validityPeriodTip: '至少需要设置一个时间单位',
    aboutPage: '关于页面'
  },
  license: {
    list: 'License列表',
    addTitle: '新增License',
    editTitle: '编辑License',
    customerRequired: '请输入客户名称',
    fingerprintRequired: '请输入机器码',
    fingerprintFormat: '正确格式应为XXXXX-XXXXX-XXXXX-XXXXX-XXXXX（5组5位字母或数字）',
    validity: '有效期',
    days: '天数',
    hours: '小时',
    minutes: '分钟',
    seconds: '秒',
    validityHint: '至少需要设置一个时间单位',
    descriptionPlaceholder: '请输入描述内容',
    fingerprintTitle: '机器码',
    currentFingerprint: '当前机器码：',
    copy: '复制',
    copySuccess: '复制成功',
    copyFailed: '复制失败'
  },
  language: {
    zh: '中文',
    en: 'English',
    ja: '日本語',
    switch: '切换语言'
  }
}

// 英文翻译
const enUS = {
  common: {
    licenseManagement: 'License Management System',
    addLicense: 'Add License',
    testFingerprint: 'Test Generate Fingerprint',
    refresh: 'Refresh Data',
    search: 'Search',
    searchPlaceholder: 'Please enter customer name to search',
    customer: 'Customer Name',
    fingerprint: 'Fingerprint',
    description: 'Description',
    activatedAt: 'Activated At',
    expiresAt: 'Expires At',
    status: 'Status',
    active: 'Active',
    expired: 'Expired',
    operations: 'Operations',
    download: 'Download',
    edit: 'Edit',
    delete: 'Delete',
    cancel: 'Cancel',
    confirm: 'Confirm',
    update: 'Update',
    previous: 'Previous',
    next: 'Next',
    total: 'Total',
    items: 'items',
    pageSize: 'items/page',
    goto: 'Go to',
    page: 'Page',
    enterCustomer: 'Please enter customer name',
    enterFingerprint: 'Please enter fingerprint',
    enterDescription: 'Please enter description',
    selectExpiresAt: 'Select expiration time',
    editLicense: 'Edit License',
    validityPeriod: 'Validity Period',
    days: 'Days',
    hours: 'Hours',
    minutes: 'Minutes',
    seconds: 'Seconds',
    fingerprintFormatTip: 'Correct format should be XXXXX-XXXXX-XXXXX-XXXXX-XXXXX (5 groups of 5 letters or numbers)',
    validityPeriodTip: 'At least one time unit needs to be set',
    aboutPage: 'About Page'
  },
  license: {
    list: 'License List',
    addTitle: 'Add License',
    editTitle: 'Edit License',
    customerRequired: 'Please enter customer name',
    fingerprintRequired: 'Please enter fingerprint',
    fingerprintFormat: 'Correct format should be XXXXX-XXXXX-XXXXX-XXXXX-XXXXX (5 groups of 5 letters or numbers)',
    validity: 'Validity Period',
    days: 'Days',
    hours: 'Hours',
    minutes: 'Minutes',
    seconds: 'Seconds',
    validityHint: 'At least one time unit needs to be set',
    descriptionPlaceholder: 'Please enter description',
    fingerprintTitle: 'Fingerprint',
    currentFingerprint: 'Current Fingerprint:',
    copy: 'Copy',
    copySuccess: 'Copy successful',
    copyFailed: 'Copy failed'
  },
  language: {
    zh: '中文',
    en: 'English',
    ja: '日本語',
    switch: 'Switch Language'
  }
}

// 日文翻译
const jaJP = {
  common: {
    licenseManagement: 'ライセンス管理システム',
    addLicense: 'ライセンス追加',
    testFingerprint: 'マシンコード生成テスト',
    refresh: 'データ更新',
    search: '検索',
    searchPlaceholder: '顧客名を入力して検索',
    customer: '顧客名',
    fingerprint: 'マシンコード',
    description: '説明',
    activatedAt: '有効化時間',
    expiresAt: '有効期限',
    status: '状態',
    active: '有効',
    expired: '期限切れ',
    operations: '操作',
    download: 'ダウンロード',
    edit: '編集',
    delete: '削除',
    cancel: 'キャンセル',
    confirm: '確定',
    update: '更新',
    previous: '前へ',
    next: '次へ',
    total: '合計',
    items: '件',
    pageSize: '件/ページ',
    goto: '移動',
    page: 'ページ',
    enterCustomer: '顧客名を入力してください',
    enterFingerprint: 'マシンコードを入力してください',
    enterDescription: '説明を入力してください',
    selectExpiresAt: '有効期限を選択',
    editLicense: 'ライセンス編集',
    validityPeriod: '有効期間',
    days: '日数',
    hours: '時間',
    minutes: '分',
    seconds: '秒',
    fingerprintFormatTip: '正しい形式はXXXXX-XXXXX-XXXXX-XXXXX-XXXXX（5組の5文字の英数字）です',
    validityPeriodTip: '少なくとも1つの時間単位を設定する必要があります',
    aboutPage: 'このページについて'
  },
  license: {
    list: 'ライセンス一覧',
    addTitle: 'ライセンス追加',
    editTitle: 'ライセンス編集',
    customerRequired: '顧客名を入力してください',
    fingerprintRequired: 'マシンコードを入力してください',
    fingerprintFormat: '正しい形式はXXXXX-XXXXX-XXXXX-XXXXX-XXXXX（5組の5文字の英数字）です',
    validity: '有効期間',
    days: '日数',
    hours: '時間',
    minutes: '分',
    seconds: '秒',
    validityHint: '少なくとも1つの時間単位を設定する必要があります',
    descriptionPlaceholder: '説明を入力してください',
    fingerprintTitle: 'マシンコード',
    currentFingerprint: '現在のマシンコード：',
    copy: 'コピー',
    copySuccess: 'コピー成功',
    copyFailed: 'コピー失敗'
  },
  language: {
    zh: '中文',
    en: 'English',
    ja: '日本語',
    switch: '言語切替'
  }
}

// 创建i18n实例
const i18n = createI18n({
  legacy: false, // 使用Composition API
  locale: 'zhCN', // 默认语言
  fallbackLocale: 'zhCN', // 回退语言
  messages: {
    zhCN,
    enUS,
    jaJP
  }
})

export default i18n