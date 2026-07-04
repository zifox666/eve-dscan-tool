import { useState } from 'react'
import { submitDScan } from '../api.js'

const reconShips = [
  { key: 'huginn', label: 'Huginn' },
  { key: 'lachesis', label: 'Lachesis' },
  { key: 'rook', label: 'Rook' },
  { key: 'curse', label: 'Curse' }
]

export default function SubmitPage({ t, uiLang, gameLang, navigate }) {
  const [data, setData] = useState('')
  const [filterDistance, setFilterDistance] = useState(false)
  const [counts, setCounts] = useState(Object.fromEntries(reconShips.map((ship) => [ship.key, 0])))
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function onSubmit(event) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      const manualShips = reconShips.map((ship) => ({ key: ship.key, count: counts[ship.key] })).filter((ship) => ship.count > 0)
      const result = await submitDScan({
        data,
        filter_distance: filterDistance,
        ui_lang: uiLang,
        game_lang: gameLang,
        manual_ships: manualShips
      })
      navigate(result.view_url)
    } catch (err) {
      setError(err.message || t.error)
    } finally {
      setSubmitting(false)
    }
  }

  function updateCount(key, delta) {
    setCounts((current) => ({ ...current, [key]: Math.max(0, current[key] + delta) }))
  }

  return (
    <section className="panel">
      <h1 className="text-2xl font-bold text-primary-600 dark:text-primary-400">{t.title}</h1>
      <p className="mt-2 text-sm text-gray-600 dark:text-gray-300">{t.subtitle}</p>
      <form className="mt-4 space-y-4" onSubmit={onSubmit}>
        <label className="block">
          <span className="mb-1 block text-sm font-medium">{t.dscanData}</span>
          <textarea className="input-field min-h-56" value={data} onChange={(e) => setData(e.target.value)} placeholder={t.paste} required />
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={filterDistance} onChange={(e) => setFilterDistance(e.target.checked)} />
          {t.filterDistance}
        </label>
        <div className="border-t border-gray-200 pt-3 dark:border-gray-800">
          <div className="mb-2 text-sm font-semibold">{t.manualRecon}</div>
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
            {reconShips.map((ship) => (
              <div key={ship.key} className="flex items-center justify-between rounded border border-gray-200 p-2 dark:border-gray-800">
                <span className="font-medium">{ship.label}</span>
                <div className="flex items-center gap-2">
                  <button type="button" className="btn-secondary h-8 px-2" onClick={() => updateCount(ship.key, -1)}>
                    -1
                  </button>
                  <span className="w-6 text-center">{counts[ship.key]}</span>
                  <button type="button" className="btn-secondary h-8 px-2" onClick={() => updateCount(ship.key, 1)}>
                    +1
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
        {error && <div className="rounded bg-red-100 p-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-200">{error}</div>}
        <button className="btn-primary" disabled={submitting}>
          {submitting ? t.loading : t.submit}
        </button>
      </form>
    </section>
  )
}
