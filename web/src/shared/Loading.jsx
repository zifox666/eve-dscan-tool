export function Loading({ t }) {
  return <div className="panel text-sm">{t.loading}...</div>
}

export function ErrorBox({ t, error, navigate }) {
  return (
    <div className="panel">
      <div className="mb-3 rounded bg-red-100 p-3 text-sm text-red-700 dark:bg-red-950 dark:text-red-200">{error.message || t.error}</div>
      <button className="btn-secondary" onClick={() => navigate('/')}>{t.back}</button>
    </div>
  )
}
