import { Languages, Moon, Sun } from 'lucide-react'

export default function Header({ t, dark, setDark, uiLang, setUiLang, gameLang, setGameLang, navigate }) {
  return (
    <header className="border-b border-gray-200 bg-white dark:border-gray-800 dark:bg-gray-900">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-2">
        <button className="text-lg font-bold text-primary-600 dark:text-primary-400" onClick={() => navigate('/')}>
          Dscan.icu
        </button>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300">
            <Languages className="h-4 w-4" />
            {t.uiLang}
            <select className="input-field h-8 w-20 py-1" value={uiLang} onChange={(e) => setUiLang(e.target.value)}>
              <option value="zh">中文</option>
              <option value="en">EN</option>
            </select>
          </label>
          <label className="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300">
            {t.gameLang}
            <select className="input-field h-8 w-20 py-1" value={gameLang} onChange={(e) => setGameLang(e.target.value)}>
              <option value="zh">中文</option>
              <option value="en">EN</option>
            </select>
          </label>
          <button className="icon-button" onClick={() => setDark(!dark)} title={dark ? 'Light' : 'Dark'}>
            {dark ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
          </button>
          <button className="icon-button" onClick={() => navigate('https://github.com/zifox666/eve-dscan-tool')} title={t.about}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                <path fill-rule="evenodd" clip-rule="evenodd"
                    d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.166 6.839 9.489.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.603-3.369-1.342-3.369-1.342-.454-1.155-1.11-1.462-1.11-1.462-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.026 2.747-1.026.546 1.378.202 2.397.1 2.65.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.161 22 16.416 22 12c0-5.523-4.477-10-10-10z" />
            </svg>
          </button>
        </div>
      </div>
    </header>
  )
}
