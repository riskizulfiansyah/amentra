import { LitElement, html } from 'lit'
import { customElement, property } from 'lit/decorators.js'

@customElement('chat-toggle')
export class ChatToggle extends LitElement {
  @property({ type: Boolean }) open = false

  createRenderRoot() {
    return this
  }

  render() {
    return html`
      <button
        part="toggle-button"
        @click=${this._handleClick}
        style="background-color: var(--ai-chat-primary, #3b82f6)"
        class="fixed bottom-5 right-5 z-50 flex h-14 w-14 cursor-pointer items-center justify-center rounded-full text-white shadow-lg transition-transform hover:scale-105 active:scale-95 focus:outline-none focus:ring-2 focus:ring-offset-2"
        aria-label=${this.open ? 'Close chat' : 'Open chat'}
      >
        ${this.open ? this._closeIcon() : this._chatIcon()}
      </button>
    `
  }

  private _handleClick(e: Event) {
    e.stopPropagation()
    this.dispatchEvent(new CustomEvent('toggle'))
  }

  private _chatIcon() {
    return html`
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>
    `
  }

  private _closeIcon() {
    return html`
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
      </svg>
    `
  }
}
