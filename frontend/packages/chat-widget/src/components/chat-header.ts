import { LitElement, html } from 'lit'
import { customElement, property } from 'lit/decorators.js'

@customElement('chat-header')
export class ChatHeader extends LitElement {
  @property({ type: String }) title = 'Chat Assistant'

  createRenderRoot() {
    return this
  }

  render() {
    return html`
      <div
        part="header"
        style="background-color: var(--ai-chat-primary, #3b82f6)"
        class="flex items-center justify-between rounded-t-xl px-4 py-3 text-white"
      >
        <div class="flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
          </svg>
          <span class="text-sm font-semibold">${this.title}</span>
        </div>
        <button
          part="close-button"
          @click=${() => this.dispatchEvent(new CustomEvent('close'))}
          class="flex cursor-pointer items-center justify-center rounded-full p-1 text-white/80 hover:bg-white/20 hover:text-white focus:outline-none"
          aria-label="Close chat"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
    `
  }
}
