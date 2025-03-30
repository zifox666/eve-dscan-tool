// 获取默认语言，优先使用本地存储的语言，其次使用浏览器语言，最后默认为英文
function getDefaultLanguage() {
  // 首先检查localStorage
  const savedLang = localStorage.getItem('language');
  if (savedLang && translations[savedLang]) {
    return savedLang;
  }

  // 然后检查Cookie
  const cookieLang = getCookieLanguage();
  if (cookieLang && translations[cookieLang]) {
    return cookieLang;
  }

  // 最后检查浏览器语言
  const browserLang = navigator.language || navigator.userLanguage;
  // 如果浏览器语言以'en'开头(如en-US, en-GB)，返回'en'，否则默认'zh'
  if (browserLang && browserLang.startsWith('en')) {
    return 'en';
  }

  // 默认使用中文
  return 'en';
}

// 从cookie中获取语言设置
function getCookieLanguage() {
  const cookies = document.cookie.split(';');
  for (let i = 0; i < cookies.length; i++) {
    const cookie = cookies[i].trim();
    if (cookie.startsWith('lang=')) {
      return cookie.substring(5);
    }
  }
  return null;
}

// 初始化当前语言
let currentLang = getDefaultLanguage();

// 翻译函数
function translate(key) {
  if (translations[currentLang] && translations[currentLang][key]) {
    return translations[currentLang][key];
  }
  // 回退到中文
  if (translations['zh'] && translations['zh'][key]) {
    return translations['zh'][key];
  }
  // 如果没有找到翻译，返回键名
  return key;
}

// 切换语言
function switchLanguage(lang) {
  if (translations[lang]) {
    currentLang = lang;
    localStorage.setItem('language', lang);
    updatePageTranslations();
  }
}

// 更新页面上的所有翻译
function updatePageTranslations() {
  document.querySelectorAll('[data-i18n]').forEach(element => {
    const key = element.getAttribute('data-i18n');
    element.textContent = translate(key);
  });

  // 更新页面标题（如果存在翻译）
  const titleText = translate('page_title');
  if (titleText !== 'page_title') {
    document.title = titleText + ' - Dscan.icu';
  }

  // 更新占位符
  document.querySelectorAll('[data-i18n-placeholder]').forEach(element => {
    const key = element.getAttribute('data-i18n-placeholder');
    element.placeholder = translate(key);
  });

  // 触发自定义事件，让其他组件知道语言已更改
  document.dispatchEvent(new CustomEvent('languageChanged', {
    detail: { language: currentLang }
  }));
}

function getCurrentLanguage() {
  return currentLang;
}

// 检查指定语言是否为当前语言
function isCurrentLanguage(lang) {
  return currentLang === lang;
}

// 页面加载完成后初始化语言
document.addEventListener('DOMContentLoaded', function() {
  // 应用初始语言设置
  updatePageTranslations();
  console.log('当前语言设置为:', currentLang);
});