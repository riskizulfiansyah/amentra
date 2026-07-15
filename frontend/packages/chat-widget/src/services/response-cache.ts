const PREFIX = 'ai-chat-cache:'
const THRESHOLD = 0.5
const TTL_MS = 30 * 60 * 1000

interface CacheEntry {
  query: string
  reply: string
  summary: string
  timestamp: number
}

export function wordOverlap(query: string, cached: string): number {
  const norm = (s: string) => s.toLowerCase().replace(/[^\w\s]/g, '')
  const qWords = norm(query).split(/\s+/).filter(Boolean)
  if (qWords.length === 0) return 0
  const cWords = norm(cached).split(/\s+/).filter(Boolean)
  const matches = qWords.filter((qw) =>
    cWords.some((cw) => cw.includes(qw) || qw.includes(cw)),
  )
  return matches.length / qWords.length
}

function prune(): void {
  const now = Date.now()
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i)
    if (!key?.startsWith(PREFIX)) continue
    try {
      const entry: CacheEntry = JSON.parse(localStorage.getItem(key)!)
      if (now - entry.timestamp > TTL_MS) localStorage.removeItem(key)
    } catch {
      localStorage.removeItem(key)
    }
  }
}

export function findCachedResponse(
  appId: string,
  text: string,
): { reply: string; summary: string } | null {
  prune()
  let bestSim = 0
  let bestKey = ''
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i)
    if (!key?.startsWith(PREFIX + appId + ':')) continue
    try {
      const entry: CacheEntry = JSON.parse(localStorage.getItem(key)!)
      const sim = wordOverlap(text, entry.query || '')
      if (sim > bestSim) {
        bestSim = sim
        bestKey = key
      }
    } catch {
      localStorage.removeItem(key)
    }
  }
  if (bestSim >= THRESHOLD) {
    try {
      const entry: CacheEntry = JSON.parse(localStorage.getItem(bestKey)!)
      if (Date.now() - entry.timestamp <= TTL_MS) {
        return { reply: entry.reply, summary: entry.summary }
      }
      localStorage.removeItem(bestKey)
    } catch { /* bestKey went stale */ }
  }
  return null
}

export function cacheResponse(
  appId: string,
  text: string,
  reply: string,
  summary: string,
): void {
  prune()
  const key = PREFIX + appId + ':' + Date.now()
  localStorage.setItem(
    key,
    JSON.stringify({
      query: text,
      reply,
      summary,
      timestamp: Date.now(),
    } as CacheEntry),
  )
}
