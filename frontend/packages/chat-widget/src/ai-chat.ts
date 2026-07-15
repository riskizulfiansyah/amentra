import { LitElement, html, unsafeCSS } from 'lit'
import { customElement, property, state, query } from 'lit/decorators.js'
import { streamMessage } from './services/chat-api.js'
import type { Message } from './types.js'
import styles from './styles.js'
import './components/chat-toggle.js'
import './components/chat-header.js'
import './components/message-bubble.js'
import './components/chat-input.js'
import { ChatInput } from './components/chat-input.js'
import { findCachedResponse, cacheResponse } from './services/response-cache.js'
import './components/typing-indicator.js'

const STORAGE_PREFIX = 'ai-chat:'

@customElement('ai-chat')
export class AiChat extends LitElement {
  static styles = unsafeCSS(styles)

  @property({ attribute: 'app-id', type: String }) appId = ''
  @property({ type: String }) apiBaseUrl = 'http://localhost:8080'
  @property({ type: String }) title = 'Chat Assistant'
  @property({ type: String }) themeColor = '#3b82f6'
  @property({ type: String }) placeholder = 'Type a message...'
  @property({ type: Number }) maxChars = 1000
  @property({ type: Boolean, reflect: true }) open = false

  @state() private _messages: Message[] = []
  @state() private _summary = ''
  @state() private _isStreaming = false
  @state() private _error: string | null = null
  @state() private _hasOpened = false

  private _abortController: AbortController | null = null
  private _pendingAssistantId: string | null = null

  @query('.chat-messages-scroll')
  private _scrollContainer!: HTMLDivElement

  @query('chat-input')
  private _chatInput!: ChatInput

  connectedCallback() {
    super.connectedCallback()
    if (!this.appId) {
      const attr = this.getAttribute('app-id')
      if (attr) this.appId = attr
    }
    this._loadState()
  }

  disconnectedCallback() {
    super.disconnectedCallback()
    this._abortController?.abort()
  }

  private get _storageKey() {
    return `${STORAGE_PREFIX}${this.appId}`
  }

  private _saveState() {
    if (!this.appId) return
    try {
      localStorage.setItem(
        this._storageKey,
        JSON.stringify({ messages: this._messages, summary: this._summary }),
      )
    } catch { /* noop */ }
  }

  private _loadState() {
    if (!this.appId) return
    try {
      const raw = localStorage.getItem(this._storageKey)
      if (raw) {
        const data = JSON.parse(raw)
        this._messages = data.messages ?? []
        this._summary = data.summary ?? ''
      }
    } catch { /* noop */ }
  }

  private _handleToggle() {
    this.open = !this.open
    this._hasOpened = true
    this.dispatchEvent(new CustomEvent('open-change', { detail: { open: this.open }, bubbles: true, composed: true }))
    if (this.open) {
      requestAnimationFrame(() => {
        this._scrollToBottom()
        this._chatInput?.focus()
      })
    }
  }

  private _handleClose() {
    this.open = false
    this.dispatchEvent(new CustomEvent('open-change', { detail: { open: false }, bubbles: true, composed: true }))
  }

  private async _handleSend(e: CustomEvent) {
    const text: string = e.detail.text
    if (!text || !this.appId) return

    const cached = findCachedResponse(this.appId, text)
    if (cached) {
      this._messages = [
        ...this._messages,
        { id: crypto.randomUUID(), role: 'user', content: text, timestamp: Date.now() },
        { id: crypto.randomUUID(), role: 'assistant', content: cached.reply, timestamp: Date.now() },
      ]
      this._summary = cached.summary
      this._saveState()
      this.dispatchEvent(new CustomEvent('message-received', { bubbles: true, composed: true }))
      requestAnimationFrame(() => {
        this._scrollToBottom()
        this._chatInput?.focus()
      })
      return
    }

    this._error = null
    this._abortController = new AbortController()

    const userMsg: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content: text,
      timestamp: Date.now(),
    }

    this._messages = [...this._messages, userMsg]
    this._pendingAssistantId = null
    this._isStreaming = true
    this._saveState()
    requestAnimationFrame(() => this._scrollToBottom())

    this.dispatchEvent(new CustomEvent('message-sent', { detail: { text }, bubbles: true, composed: true }))

    const recent = this._messages.slice(-9).map((m) => ({
      role: m.role,
      content: m.content,
    }))

    streamMessage(
      this.apiBaseUrl,
      {
        app_id: this.appId,
        summary: this._summary,
        recent_messages: recent,
        message: text,
      },
      {
        onToken: (token) => this._onToken(token),
        onDone: (summary) => {
          const lastId = this._pendingAssistantId
          this._onStreamDone(summary)
          if (lastId) {
            const replyMsg = this._messages.find((m) => m.id === lastId)
            if (replyMsg) {
              cacheResponse(this.appId, text, replyMsg.content, summary)
            }
          }
        },
        onError: (err) => this._onStreamError(err),
      },
      this._abortController.signal,
    )
  }

  private _onToken(token: string) {
    if (!this._pendingAssistantId) {
      const msg: Message = {
        id: crypto.randomUUID(),
        role: 'assistant',
        content: token,
        timestamp: Date.now(),
      }
      this._pendingAssistantId = msg.id
      this._messages = [...this._messages, msg]
    } else {
      this._messages = this._messages.map((m) =>
        m.id === this._pendingAssistantId
          ? { ...m, content: m.content + token }
          : m,
      )
    }
    this._scrollToBottom()
  }

  private _onStreamDone(summary: string) {
    this._isStreaming = false
    this._summary = summary
    this._pendingAssistantId = null
    this._abortController = null
    this._saveState()
    this.dispatchEvent(new CustomEvent('message-received', { bubbles: true, composed: true }))
    requestAnimationFrame(() => {
      this._scrollToBottom()
      this._chatInput?.focus()
    })
  }

  private _onStreamError(err: string) {
    this._isStreaming = false
    this._error = err
    this._pendingAssistantId = null
    this._abortController = null
    this.dispatchEvent(new CustomEvent('error', { detail: { error: err }, bubbles: true, composed: true }))
    this._scrollToBottom()
  }

  private _scrollToBottom() {
    if (this._scrollContainer) {
      this._scrollContainer.scrollTop = this._scrollContainer.scrollHeight
    }
  }

  render() {
    if (!this.appId) {
      return html`
        <div class="chat-widget-wrapper">
          <div class="fixed bottom-5 right-5 z-50 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600 shadow-lg ring-1 ring-black/5">
            Missing <code class="font-semibold">app-id</code> attribute
          </div>
        </div>
      `
    }

    return html`
      <div class="chat-widget-wrapper" style="--ai-chat-primary: ${this.themeColor}">
        <chat-toggle
          .open=${this.open}
          @toggle=${this._handleToggle}
        ></chat-toggle>

        ${this.open && this._hasOpened ? this._renderDialog() : ''}
      </div>
    `
  }

  private _renderDialog() {
    return html`
      <div
        part="dialog"
        class="fixed bottom-24 right-5 z-50 flex w-[380px] max-h-[calc(100vh-120px)] flex-col rounded-xl bg-white shadow-2xl ring-1 ring-black/5 animate-slide-up overflow-hidden max-sm:bottom-0 max-sm:right-0 max-sm:w-full max-sm:h-[100dvh] max-sm:rounded-none"
      >
        <chat-header
          .title=${this.title}
          @close=${this._handleClose}
        ></chat-header>

        <div class="chat-messages-scroll flex-1 min-h-0 overflow-y-auto px-4 py-4 chat-scrollbar">
          ${this._messages.length === 0 ? this._renderEmpty() : ''}

          ${this._messages.map(
            (msg) => html`
              <div class="mb-3 flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}">
                <div class="max-w-[85%]">
                  <message-bubble .message=${msg}></message-bubble>
                </div>
              </div>
            `
          )}

          ${this._isStreaming && !this._pendingAssistantId
            ? html`
                <div class="mb-3 flex justify-start">
                  <div class="rounded-2xl rounded-bl-md border border-gray-200 bg-white px-4 py-3">
                    <typing-indicator></typing-indicator>
                  </div>
                </div>
              `
            : ''}
        </div>

        ${this._error ? this._renderError() : ''}

        <chat-input
          .disabled=${this._isStreaming}
          .maxChars=${this.maxChars}
          .placeholder=${this.placeholder}
          @send=${this._handleSend}
        ></chat-input>
      </div>
    `
  }

  private _renderEmpty() {
    return html`
      <div class="flex flex-col items-center justify-center py-10 text-center animate-fade-in">
        <div
          class="mb-4 flex h-16 w-16 items-center justify-center rounded-full"
          style="background-color: color-mix(in srgb, var(--ai-chat-primary, #3b82f6) 12%, transparent)"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="var(--ai-chat-primary, #3b82f6)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
          </svg>
        </div>
        <p class="mb-1 text-sm font-medium text-gray-700">Hello! How can I help you?</p>
        <p class="text-xs text-gray-400">Ask me anything about this topic.</p>
      </div>
    `
  }

  private _renderError() {
    return html`
      <div
        part="error-banner"
        class="mx-3 mb-2 flex items-center gap-2 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600 animate-fade-in"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="flex-shrink-0">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        <span class="flex-1">${this._error}</span>
        <button
          @click=${() => (this._error = null)}
          class="cursor-pointer font-medium text-red-700 underline hover:text-red-800"
        >
          Dismiss
        </button>
      </div>
    `
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'ai-chat': AiChat
  }
}
