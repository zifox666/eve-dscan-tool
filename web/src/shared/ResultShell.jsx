import { useState } from 'react'
import html2canvas from 'html2canvas'
import { Camera, Check, Clipboard, RotateCcw } from 'lucide-react'

export default function ResultShell({ t, createdAt, onBack, children }) {
  const [copied, setCopied] = useState(false)
  async function copyLink() {
    await navigator.clipboard.writeText(window.location.href)
    setCopied(true)
    setTimeout(() => setCopied(false), 1400)
  }
  async function screenshot() {
    const canvas = await html2canvas(document.body, { useCORS: true, scale: 1 })
    const link = document.createElement('a')
    link.download = `eve-dscan-${new Date().toISOString().slice(0, 10)}.png`
    link.href = canvas.toDataURL('image/png')
    link.click()
  }
  return (
    <section className="panel">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
        <div className="text-sm text-gray-500">{t.createdAt}: {new Date(createdAt).toLocaleString()}</div>
        <div className="flex gap-2">
          <button className="btn-secondary" onClick={copyLink}>{copied ? <Check className="h-4 w-4" /> : <Clipboard className="h-4 w-4" />}{copied ? t.copied : t.copyLink}</button>
          <button className="btn-secondary" onClick={screenshot}><Camera className="h-4 w-4" />{t.screenshot}</button>
          <button className="btn-secondary" onClick={onBack}><RotateCcw className="h-4 w-4" />{t.back}</button>
        </div>
      </div>
      {children}
    </section>
  )
}
