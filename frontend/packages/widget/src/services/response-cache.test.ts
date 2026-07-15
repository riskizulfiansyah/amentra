import { describe, expect, test, beforeEach } from 'bun:test'
import { wordOverlap, findCachedResponse, cacheResponse } from './response-cache'

const store: Record<string, string> = {}

beforeEach(() => {
  Object.keys(store).forEach((k) => delete store[k])
  ;(globalThis as any).localStorage = {
    getItem: (k: string) => store[k] ?? null,
    setItem: (k: string, v: string) => { store[k] = v },
    removeItem: (k: string) => { delete store[k] },
    clear: () => { Object.keys(store).forEach((k) => delete store[k]) },
    get length() { return Object.keys(store).length },
    key: (i: number) => Object.keys(store)[i] ?? null,
  } as Storage
})

describe('wordOverlap', () => {
  test('exact match', () => {
    expect(wordOverlap('proyek', 'proyek')).toBe(1)
  })

  test('substring match', () => {
    expect(wordOverlap('proyek', 'pengalaman proyeknya')).toBe(1)
  })

  test('all query words match', () => {
    expect(wordOverlap('pengalaman proyek', 'pengalaman proyeknya seperti apa')).toBe(1)
  })

  test('no match', () => {
    expect(wordOverlap('makanan', 'pengalaman proyek')).toBe(0)
  })

  test('empty query', () => {
    expect(wordOverlap('', 'pengalaman proyek')).toBe(0)
  })

  test('short query matches longer word', () => {
    expect(wordOverlap('a', 'apakah')).toBe(1)
  })

  test('punctuation stripped', () => {
    expect(wordOverlap('proyek!', 'pengalaman proyeknya')).toBe(1)
  })

  test('case insensitive', () => {
    expect(wordOverlap('Proyek', 'PENGALAMAN PROYEKNYA')).toBe(1)
  })
})

describe('cacheResponse + findCachedResponse', () => {
  const appId = 'test-app'
  const text = 'pengalaman proyek'
  const reply = 'Berikut daftar proyek...'
  const summary = 'user bertanya tentang proyek'

  test('stores and retrieves exact match', () => {
    cacheResponse(appId, text, reply, summary)
    const result = findCachedResponse(appId, text)
    expect(result).not.toBeNull()
    expect(result!.reply).toBe(reply)
    expect(result!.summary).toBe(summary)
  })

  test('retrieves similar query', () => {
    cacheResponse(appId, 'pengalaman proyeknya seperti apa', reply, summary)
    const result = findCachedResponse(appId, 'proyek')
    expect(result).not.toBeNull()
    expect(result!.reply).toBe(reply)
  })

  test('returns null for unrelated query', () => {
    cacheResponse(appId, 'pengalaman proyek', reply, summary)
    const result = findCachedResponse(appId, 'makanan enak')
    expect(result).toBeNull()
  })

  test('returns null on empty cache', () => {
    const result = findCachedResponse(appId, 'halo')
    expect(result).toBeNull()
  })
})
