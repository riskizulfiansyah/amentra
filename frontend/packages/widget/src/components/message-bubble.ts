import { LitElement, html } from 'lit'
import { customElement, property } from 'lit/decorators.js'
import { unsafeHTML } from 'lit/directives/unsafe-html.js'
import { marked } from 'marked'
import type { Message } from '../types.js'

@customElement('message-bubble')
export class MessageBubble extends LitElement {
  @property({ type: Object }) message!: Message

  createRenderRoot() {
    return this
  }

  render() {
    const isUser = this.message.role === 'user'

    return html`
      <div
        class="w-full rounded-2xl px-4 py-2.5 text-sm leading-relaxed ${isUser
          ? 'rounded-br-md text-white'
          : 'rounded-bl-md border border-gray-200 bg-white text-gray-800'}"
        style=${isUser
          ? `background-color: var(--amentra-primary, #3b82f6)`
          : ''}
      >
        ${isUser
          ? html`<p class="whitespace-pre-wrap">${this.message.content}</p>`
          : html`<div class="prose prose-sm max-w-none prose-headings:text-gray-800 prose-a:text-blue-600 prose-code:rounded prose-code:bg-gray-100 prose-code:px-1 prose-code:py-0.5 prose-pre:bg-gray-900 prose-pre:text-gray-100">
              ${unsafeHTML(marked.parse(this.message.content, { async: false }) as string)}
            </div>`}
      </div>
    `
  }
}
