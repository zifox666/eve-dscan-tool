import { useEffect, useState } from 'react'

export function useStoredLang(key, fallback) {
  const [value, setValue] = useState(() => localStorage.getItem(key) || fallback)
  useEffect(() => {
    localStorage.setItem(key, value)
    document.cookie = `${key}=${value}; max-age=${60 * 60 * 24 * 365}; path=/; SameSite=Lax`
    if (key === 'ui_lang') document.documentElement.lang = value === 'en' ? 'en' : 'zh-CN'
  }, [key, value])
  return [value, setValue]
}

export function useStoredBoolean(key, fallback) {
  const [value, setValue] = useState(() => localStorage.getItem(key) === null ? fallback : localStorage.getItem(key) === 'true')
  useEffect(() => localStorage.setItem(key, String(value)), [key, value])
  return [value, setValue]
}

export function useAsyncData(loader, deps) {
  const [state, setState] = useState({ loading: true, error: null, data: null })
  useEffect(() => {
    let active = true
    setState({ loading: true, error: null, data: null })
    loader()
      .then((data) => active && setState({ loading: false, error: null, data }))
      .catch((error) => active && setState({ loading: false, error, data: null }))
    return () => {
      active = false
    }
  }, deps) // eslint-disable-line react-hooks/exhaustive-deps
  return state
}
