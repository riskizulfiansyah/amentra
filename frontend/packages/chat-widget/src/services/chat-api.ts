import type { ChatRequest, ChatResponse } from '../types.js'

export interface StreamCallbacks {
  onToken: (token: string) => void
  onDone: (summary: string) => void
  onError: (error: string) => void
}

export async function sendMessage(
  baseUrl: string,
  req: ChatRequest
): Promise<ChatResponse> {
  const res = await fetch(`${baseUrl}/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  return res.json()
}

export async function streamMessage(
  baseUrl: string,
  req: ChatRequest,
  callbacks: StreamCallbacks,
  signal?: AbortSignal
): Promise<void> {
  const res = await fetch(`${baseUrl}/chat-stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
    signal,
  })

  if (!res.ok) {
    const body = await res.json().catch(() => null)
    callbacks.onError(body?.error ?? 'request failed')
    return
  }

  const reader = res.body?.getReader()
  if (!reader) {
    callbacks.onError('no response body')
    return
  }

  const decoder = new TextDecoder()
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6)

        let event: Record<string, unknown>
        try {
          event = JSON.parse(data)
        } catch {
          continue
        }

        switch (event.type) {
          case 'token':
            callbacks.onToken(String(event.content ?? ''))
            break
          case 'done':
            callbacks.onDone(String(event.summary ?? ''))
            return
          case 'error':
            callbacks.onError(String(event.message ?? 'unknown error'))
            return
        }
      }
    }
  } catch (err) {
    if ((err as Error)?.name !== 'AbortError') {
      callbacks.onError(String(err))
    }
  }
}
