import { createComponent } from '@lit/react'
import React from 'react'
import { AiChat as AiChatElement } from '@ai-chat/widget'

export const AiChat = createComponent({
  tagName: 'ai-chat',
  elementClass: AiChatElement,
  react: React,
  events: {
    onMessageSent: 'message-sent',
    onMessageReceived: 'message-received',
    onError: 'error',
    onOpenChange: 'open-change',
  },
})
