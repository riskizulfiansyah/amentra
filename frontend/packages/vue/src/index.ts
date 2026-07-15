import { defineComponent, h } from 'vue'
import '@amentra/widget'

export const AmentraWidget = defineComponent({
  name: 'AmentraWidget',
  props: {
    appId: { type: String, required: true },
    apiBaseUrl: { type: String, default: 'http://localhost:8080' },
    title: { type: String, default: 'Chat Assistant' },
    themeColor: { type: String, default: '#3b82f6' },
    placeholder: { type: String, default: 'Type a message...' },
    maxChars: { type: Number, default: 1000 },
    open: { type: Boolean, default: false },
  },
  emits: ['message-sent', 'message-received', 'error', 'open-change'],
  setup(props, { expose, emit }) {
    let element: HTMLElement | null = null

    const onRef = (el: unknown) => {
      element = el as HTMLElement
    }

    expose({
      open: () => element?.setAttribute('open', ''),
      close: () => element?.removeAttribute('open'),
    })

    return () =>
      h('amentra-widget', {
        ref: onRef,
        'app-id': props.appId,
        'api-base-url': props.apiBaseUrl,
        title: props.title,
        'theme-color': props.themeColor,
        placeholder: props.placeholder,
        'max-chars': props.maxChars,
        ...(props.open ? { open: '' } : {}),
      })
  },
})
