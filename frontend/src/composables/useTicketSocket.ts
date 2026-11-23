import { onMounted, onUnmounted, ref, watch } from 'vue'
import { getStoredToken } from '@/services/api'
import { useAuthStore } from '@/stores/authStore'
import { useTicketStore } from '@/stores/ticketStore'
import type { Ticket } from '@/types/ticket'

export interface TicketNotification {
  id: string
  ticket: Ticket
  receivedAt: string
}

export interface TicketSocketEvent {
  type: 'ticket.created' | 'ticket.updated'
  data: Ticket
}

const notifications = ref<TicketNotification[]>([])
const lastTicketEvent = ref<TicketSocketEvent | null>(null)
const connected = ref(false)

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempts = 0
let activeConsumers = 0
let intentionallyClosed = false

function buildWebSocketUrl(token: string): string {
  const apiBase = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/$/, '')
  const encodedToken = encodeURIComponent(token)

  if (apiBase.startsWith('http://') || apiBase.startsWith('https://')) {
    const wsBase = apiBase.replace(/^http/, 'ws')
    return `${wsBase}/ws/tickets?token=${encodedToken}`
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}${apiBase}/ws/tickets?token=${encodedToken}`
}

function dismissNotification(id: string) {
  notifications.value = notifications.value.filter((item) => item.id !== id)
}

function pushNotification(ticket: Ticket) {
  const id = `${ticket.id}-${Date.now()}`
  notifications.value = [
    {
      id,
      ticket,
      receivedAt: new Date().toISOString(),
    },
    ...notifications.value.slice(0, 4),
  ]

  setTimeout(() => {
    dismissNotification(id)
  }, 6000)
}

function clearReconnectTimer() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

function disconnect() {
  intentionallyClosed = true
  clearReconnectTimer()
  if (socket) {
    socket.close()
    socket = null
  }
  connected.value = false
}

function connect() {
  const token = getStoredToken()
  if (!token) {
    disconnect()
    return
  }

  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }

  intentionallyClosed = false
  const wsUrl = buildWebSocketUrl(token)
  const ws = new WebSocket(wsUrl)
  socket = ws

  ws.onopen = () => {
    connected.value = true
    reconnectAttempts = 0
  }

  ws.onmessage = (event) => {
    try {
      const parsed = JSON.parse(event.data) as TicketSocketEvent
      if (!parsed || !parsed.type || !parsed.data) return

      const authStore = useAuthStore()
      const currentUser = authStore.user
      if (!currentUser) return

      const isAdminOrAgent = currentUser.role === 'admin' || currentUser.role === 'agent'
      const isOwnRequesterTicket =
        currentUser.role === 'requester' && parsed.data.requester_id === currentUser.id

      if (!isAdminOrAgent && !isOwnRequesterTicket) {
        return
      }

      lastTicketEvent.value = parsed
      const ticketStore = useTicketStore()
      void ticketStore.fetchTickets()
      void ticketStore.fetchStatusCounts()
      void ticketStore.fetchLateTickets()

      if (parsed.type === 'ticket.created') {
        pushNotification(parsed.data)
      }
    } catch {
      // Ignore malformed frames
    }
  }

  ws.onclose = () => {
    if (socket === ws) {
      socket = null
    }
    connected.value = false

    if (!intentionallyClosed && activeConsumers > 0 && getStoredToken()) {
      clearReconnectTimer()
      const delay = Math.min(1000 * 2 ** reconnectAttempts, 15000)
      reconnectAttempts += 1
      reconnectTimer = setTimeout(() => {
        connect()
      }, delay)
    }
  }

  ws.onerror = () => {
    ws.close()
  }
}

export function useTicketSocket() {
  const authStore = useAuthStore()

  onMounted(() => {
    activeConsumers += 1
    if (authStore.isAuthenticated) {
      connect()
    }
  })

  watch(
    () => authStore.isAuthenticated,
    (isAuth) => {
      if (isAuth && activeConsumers > 0) {
        connect()
      } else if (!isAuth) {
        disconnect()
      }
    },
  )

  onUnmounted(() => {
    activeConsumers = Math.max(0, activeConsumers - 1)
    if (activeConsumers === 0) {
      disconnect()
    }
  })

  return {
    notifications,
    lastTicketEvent,
    connected,
    dismissNotification,
  }
}
