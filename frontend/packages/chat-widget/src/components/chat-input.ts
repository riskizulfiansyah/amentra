import { LitElement, html } from 'lit'
import { customElement, property, state, query } from 'lit/decorators.js'

@customElement('chat-input')
export class ChatInput extends LitElement {
  @property({ type: Boolean }) disabled = false
  @property({ type: Number }) maxChars = 1000
  @property({ type: String }) placeholder = 'Type a message...'

  @state() private _text = ''

  @query('textarea') private _textarea!: HTMLTextAreaElement

  createRenderRoot() {
    return this
  }

  render() {
    const remaining = this.maxChars - this._text.length
    const overLimit = remaining < 0

    return html`
      <div class="border-t border-gray-200 bg-white p-3">
        <div class="flex items-end gap-2">
          <div class="relative flex-1">
            <textarea
              part="input"
              .value=${this._text}
              @input=${this._onInput}
              @keydown=${this._onKeydown}
              ?disabled=${this.disabled}
              placeholder=${this.placeholder}
              rows="1"
              class="max-h-32 w-full resize-none rounded-xl border border-gray-300 bg-gray-50 px-3 py-2 text-sm text-gray-800 placeholder-gray-400 outline-none transition-colors focus:border-blue-500 focus:bg-white focus:ring-1 focus:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
            ></textarea>
          </div>
          <button
            part="send-button"
            @click=${this._send}
            ?disabled=${this.disabled || !this._text.trim() || overLimit}
            class="flex h-9 w-9 flex-shrink-0 cursor-pointer items-center justify-center rounded-full text-white transition-colors disabled:cursor-not-allowed disabled:opacity-40"
            style="background-color: var(--ai-chat-primary, #3b82f6)"
            aria-label="Send message"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>
            </svg>
          </button>
        </div>
        ${overLimit ? html`<p class="mt-1 text-xs text-red-500">${remaining} characters remaining</p>` : ''}
      </div>
    `
  }

  private _onInput(e: Event) {
    const textarea = e.target as HTMLTextAreaElement
    this._text = textarea.value
    this._autoResize(textarea)
  }

  private _onKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      this._send()
    }
  }

  private _autoResize(textarea: HTMLTextAreaElement) {
    textarea.style.height = 'auto'
    textarea.style.height = `${Math.min(textarea.scrollHeight, 128)}px`
  }

  private _send() {
    const text = this._text.trim()
    if (!text || this.disabled) return
    this.dispatchEvent(new CustomEvent('send', { detail: { text } }))
    this._text = ''
    if (this._textarea) {
      this._textarea.style.height = 'auto'
    }
  }

  focus() {
    if (this._textarea) this._textarea.focus()
  }
}
