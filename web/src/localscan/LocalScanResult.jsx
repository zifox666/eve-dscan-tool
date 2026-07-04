import { useCallback, useMemo, useRef, useState } from 'react'
import { useAsyncData } from '../shared/hooks.js'
import { sortByCount } from '../shared/utils.js'
import { Loading, ErrorBox } from '../shared/Loading.jsx'
import ResultShell from '../shared/ResultShell.jsx'
import { Stats } from '../shared/Stats.jsx'
import { fetchLocal } from '../api.js'

const FILTER_TYPES = ['alliance', 'corporation']
const TYPE_ORDER = { alliance: 0, corporation: 1, character: 2 }

export default function LocalScanResult({ t, shortId, navigate }) {
  const { data, error, loading } = useAsyncData(() => fetchLocal(shortId), [shortId])

  const rawChars = data?.processed_data?.characters || {}
  const relMap = useMemo(() => buildRelationships(rawChars), [rawChars])

  const [filterChain, setFilterChain] = useState([])
  const [hoveredKey, setHoveredKey] = useState(null)
  const timerRef = useRef(null)

  const chainForDisplay = filterChain.length > 0 ? filterChain : (hoveredKey ? [hoveredKey] : [])

  const onHover = useCallback((key) => {
    if (filterChain.length > 0) return
    if (timerRef.current) clearTimeout(timerRef.current)
    setHoveredKey(key)
  }, [filterChain])

  const onUnhover = useCallback(() => {
    if (filterChain.length > 0) return
    timerRef.current = setTimeout(() => setHoveredKey(null), 60)
  }, [filterChain])

  const onClick = useCallback((key) => {
    const [filterType, filterId] = key.split(':')
    const typeIdx = TYPE_ORDER[filterType]

    setFilterChain((prev) => {
      const existingIdx = prev.findIndex((f) => f.startsWith(filterType + ':'))
      if (existingIdx !== -1 && prev[existingIdx] === key) {
        return prev.slice(0, existingIdx)
      }

      const newChain = prev.slice(0, typeIdx)
      newChain.push(key)
      return newChain
    })
    setHoveredKey(null)
  }, [])

  if (loading) return <Loading t={t} />
  if (error) return <ErrorBox t={t} error={error} navigate={navigate} />

  const processed = data.processed_data
  const characters = Object.values(rawChars)
  const corporations = sortByCount(Object.values(processed.corporations || {}), 'character_count')
  const alliances = sortByCount(Object.values(processed.alliances || {}), 'character_count')
  const withoutAlliance = characters.filter((character) => !character.alliance_id).length

  return (
    <ResultShell t={t} createdAt={data.created_at} onBack={() => navigate('/')}>
      <Stats items={[
        [t.characters, processed.stats?.character_count || 0],
        [t.corporations, processed.stats?.corporation_count || 0],
        [t.alliances, processed.stats?.alliance_count || 0],
        [t.noAlliance, withoutAlliance]
      ]} />
      <div className="grid gap-3 md:grid-cols-3">
        <EntityColumn
          title={t.alliances}
          items={alliances}
          countKey="character_count"
          type="alliance"
          relMap={relMap}
          filterChain={chainForDisplay}
          onHover={onHover}
          onUnhover={onUnhover}
          onClick={onClick}
        />
        <EntityColumn
          title={t.corporations}
          items={corporations}
          countKey="character_count"
          type="corporation"
          relMap={relMap}
          filterChain={chainForDisplay}
          onHover={onHover}
          onUnhover={onUnhover}
          onClick={onClick}
        />
        <EntityColumn
          title={t.characters}
          items={characters}
          type="character"
          relMap={relMap}
          filterChain={chainForDisplay}
          onHover={onHover}
          onUnhover={onUnhover}
        />
      </div>
    </ResultShell>
  )
}

function buildRelationships(rawChars) {
  // allianceId -> { corpIds: Set<string>, charIds: Set<string> }
  // corpId    -> { allianceId: string|null, charIds: Set<string> }
  // charId    -> { allianceId: string|null, corpId: string }
  const allianceMap = {}
  const corpMap = {}
  const charMap = {}

  for (const charId of Object.keys(rawChars)) {
    const c = rawChars[charId]
    const corpId = String(c.corporation_id)
    const allianceId = c.alliance_id ? String(c.alliance_id) : null

    charMap[charId] = { allianceId, corpId }

    if (!corpMap[corpId]) corpMap[corpId] = { allianceId: null, charIds: new Set() }
    if (allianceId) corpMap[corpId].allianceId = allianceId
    corpMap[corpId].charIds.add(charId)

    if (allianceId) {
      if (!allianceMap[allianceId]) allianceMap[allianceId] = { corpIds: new Set(), charIds: new Set() }
      allianceMap[allianceId].corpIds.add(corpId)
      allianceMap[allianceId].charIds.add(charId)
    }
  }

  return { allianceMap, corpMap, charMap }
}

function computeValidChars(relMap, filterChain) {
  let valid = new Set(Object.keys(relMap.charMap))

  for (const entry of filterChain) {
    const [ft, fid] = entry.split(':')

    if (ft === 'alliance') {
      const aRel = relMap.allianceMap[fid]
      if (!aRel) return new Set()
      valid = new Set([...valid].filter((c) => aRel.charIds.has(c)))
    }

    if (ft === 'corporation') {
      const cRel = relMap.corpMap[fid]
      if (!cRel) return new Set()
      valid = new Set([...valid].filter((c) => cRel.charIds.has(c)))
    }

    if (ft === 'character') {
      valid = new Set([...valid].filter((c) => c === fid))
    }
  }

  return valid
}

function getItemState(item, type, relMap, filterChain) {
  if (!filterChain || filterChain.length === 0) {
    return { visible: true, active: false }
  }

  const itemId = String(item.id)
  const isActive = filterChain.some((f) => f === `${type}:${itemId}`)
  const hasSelfFilter = filterChain.some((f) => f.startsWith(type + ':'))

  if (hasSelfFilter) {
    return { visible: true, active: isActive }
  }

  const validChars = computeValidChars(relMap, filterChain)

  if (type === 'corporation') {
    const visible = [...validChars].some(
      (charId) => relMap.charMap[charId]?.corpId === itemId,
    )
    return { visible, active: false }
  }

  if (type === 'character') {
    return { visible: validChars.has(itemId), active: false }
  }

  return { visible: true, active: false }
}

function EntityColumn({ title, items, countKey, type, relMap, filterChain, onHover, onUnhover, onClick }) {
  return (
    <div className="overflow-hidden rounded border border-gray-200 dark:border-gray-800">
      <div className="bg-primary-600 px-3 py-2 text-center text-sm font-bold text-white">{title}</div>
      <div className="max-h-[70vh] overflow-y-auto">
        {items.map((item) => {
          const { visible, active } = getItemState(item, type, relMap, filterChain)
          const key = `${type}:${item.id}`
          return (
            <div
              key={item.id}
              className={`entity-item${active ? ' entity-item-active' : ''}`}
              style={{ display: visible ? '' : 'none' }}
              onMouseEnter={() => onHover(key)}
              onMouseLeave={onUnhover}
              onClick={onClick ? () => onClick(key) : undefined}
            >
              <img
                className="mr-2 h-8 w-8 flex-shrink-0 rounded-full"
                src={`https://images.evetech.net/${type}s/${item.id}/${type === 'character' ? 'portrait' : 'logo'}?size=32`}
                alt=""
              />
              <span className="flex-1 truncate">{item.name}</span>
              {countKey && (
                <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs dark:bg-gray-800">
                  {item[countKey] || 0}
                </span>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
