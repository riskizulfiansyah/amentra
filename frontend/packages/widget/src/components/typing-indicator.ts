import { LitElement, html } from 'lit'
import { customElement } from 'lit/decorators.js'

@customElement('typing-indicator')
export class TypingIndicator extends LitElement {
  createRenderRoot() {
    return this
  }

  render() {
    return html`
      <div class="flex items-center gap-1 px-1 py-2">
        <span class="typing-dot"></span>
        <span class="typing-dot"></span>
        <span class="typing-dot"></span>
      </div>
    `
  }
}
