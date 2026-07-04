export async function submitDScan(payload) {
  return apiFetch('/api/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })
}

export async function fetchLocal(shortId) {
  return apiFetch(`/api/c/${encodeURIComponent(shortId)}`)
}

export async function fetchShip(shortId, gameLang) {
  return apiFetch(`/api/v/${encodeURIComponent(shortId)}?game_lang=${encodeURIComponent(gameLang)}`)
}

async function apiFetch(url, options = {}) {
  const response = await fetch(url, {
    ...options,
    headers: {
      Accept: 'application/json',
      ...(options.headers || {})
    }
  })
  const payload = await response.json().catch(() => null)
  if (!response.ok || !payload || payload.code >= 400) {
    throw new Error(payload?.msg || `HTTP ${response.status}`)
  }
  return payload.data
}

