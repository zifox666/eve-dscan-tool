import { useEffect, useMemo, useState } from 'react'
import { systemLanguage, translations } from './i18n.js'
import { useStoredLang, useStoredBoolean } from './shared/hooks.js'
import Header from './shared/Header.jsx'
import SubmitPage from './submit/SubmitPage.jsx'
import LocalScanResult from './localscan/LocalScanResult.jsx'
import ShipScanResult from './shipscan/ShipScanResult.jsx'

export default function App() {
  const [uiLang, setUiLang] = useStoredLang('ui_lang', systemLanguage())
  const [gameLang, setGameLang] = useStoredLang('game_lang', systemLanguage())
  const [dark, setDark] = useStoredBoolean('dark_mode', false)
  const [path, setPath] = useState(window.location.pathname)
  const t = translations[uiLang] || translations.zh

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
  }, [dark])

  useEffect(() => {
    const onPop = () => setPath(window.location.pathname)
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  function navigate(nextPath) {
    window.history.pushState({}, '', nextPath)
    setPath(nextPath)
  }

  const page = useMemo(() => {
    const match = path.match(/^\/([cv])\/([^/]+)$/)
    if (!match) return { type: 'home' }
    return { type: match[1] === 'c' ? 'local' : 'ship', shortId: match[2] }
  }, [path])

  return (
    <div className="min-h-screen">
      <Header
        t={t}
        dark={dark}
        setDark={setDark}
        uiLang={uiLang}
        setUiLang={setUiLang}
        gameLang={gameLang}
        setGameLang={setGameLang}
        navigate={navigate}
      />
      <main className="mx-auto max-w-6xl px-4 py-4">
        {page.type === 'home' && <SubmitPage t={t} uiLang={uiLang} gameLang={gameLang} navigate={navigate} />}
        {page.type === 'local' && <LocalScanResult t={t} shortId={page.shortId} navigate={navigate} />}
        {page.type === 'ship' && <ShipScanResult t={t} shortId={page.shortId} gameLang={gameLang} navigate={navigate} />}
      </main>
    </div>
  )
}

