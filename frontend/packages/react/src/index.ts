import { createComponent } from '@lit/react'
import React from 'react'
import { AmentraWidget as AmentraWidgetElement } from '@amentra/widget'

export const AmentraWidget = createComponent({
  tagName: 'amentra-widget',
  elementClass: AmentraWidgetElement,
  react: React,
  events: {
    onMessageSent: 'message-sent',
    onMessageReceived: 'message-received',
    onError: 'error',
    onOpenChange: 'open-change',
  },
})
