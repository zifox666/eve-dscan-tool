export const translations = {
  zh: {
    title: 'EVE Online DScan 分析工具',
    subtitle: '粘贴本地频道成员列表或舰船 DScan，生成可分享的分析结果。',
    dscanData: 'DScan 数据',
    paste: '粘贴 DScan 数据',
    submit: '提交',
    filterDistance: '仅统计有距离信息的扫描结果',
    manualRecon: '手动添加侦查舰',
    uiLang: '界面语言',
    gameLang: '游戏名词',
    copyLink: '复制链接',
    copied: '已复制',
    screenshot: '截图',
    back: '返回首页',
    total: '总数量',
    ships: '舰船',
    capitals: '旗舰',
    structures: '建筑',
    other: '其他',
    characters: '角色',
    corporations: '公司',
    alliances: '联盟',
    noAlliance: '无联盟',
    createdAt: '创建时间',
    loading: '加载中',
    error: '请求失败'
  },
  en: {
    title: 'EVE Online DScan Tool',
    subtitle: 'Paste local members or ship DScan data and create a shareable analysis.',
    dscanData: 'DScan Data',
    paste: 'Paste DScan data',
    submit: 'Submit',
    filterDistance: 'Only count entries with distance',
    manualRecon: 'Add recon ships',
    uiLang: 'UI Language',
    gameLang: 'Game Terms',
    copyLink: 'Copy Link',
    copied: 'Copied',
    screenshot: 'Screenshot',
    back: 'Back',
    total: 'Total',
    ships: 'Ships',
    capitals: 'Capitals',
    structures: 'Structures',
    other: 'Other',
    characters: 'Characters',
    corporations: 'Corporations',
    alliances: 'Alliances',
    noAlliance: 'No Alliance',
    createdAt: 'Created At',
    loading: 'Loading',
    error: 'Request failed'
  }
}

export function systemLanguage() {
  const lang = navigator.language || 'zh'
  return lang.toLowerCase().startsWith('en') ? 'en' : 'zh'
}

