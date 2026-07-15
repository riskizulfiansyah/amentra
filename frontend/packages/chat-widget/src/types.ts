export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: number
}

export interface ChatRequest {
  app_id: string
  summary: string
  recent_messages: { role: string; content: string }[]
  message: string
}

export interface ChatResponse {
  reply: string
  summary: string
}

export interface ChatState {
  isOpen: boolean
  messages: Message[]
  summary: string
  isStreaming: boolean
  error: string | null
}
