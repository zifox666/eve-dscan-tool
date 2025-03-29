// static/js/i18n.js
let currentLang = localStorage.getItem('language') || 'zh';

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

  // 更新页面标题
  document.title = translate('page_title') + ' - Dscan.icu';

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
