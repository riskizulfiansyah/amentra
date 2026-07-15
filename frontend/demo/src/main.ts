import '@ai-chat/widget'

const chat = document.querySelector('ai-chat')
if (chat) {
  const baseUrl = import.meta.env.VITE_API_BASE_URL
  const appId = import.meta.env.VITE_APP_ID
  const title = import.meta.env.VITE_APP_TITLE
  if (baseUrl) chat.setAttribute('api-base-url', baseUrl)
  if (appId) chat.setAttribute('app-id', appId)
  if (title) chat.setAttribute('title', title)
}
