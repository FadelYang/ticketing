import { computed, reactive, ref } from 'vue'
import { defineStore } from 'pinia'
import { ApiError } from '@/services/api'
import { ticketService } from '@/services/ticketService'
import type { AuthUser } from '@/types/auth'
import type { CreateTicketPayload, Ticket, TicketPriority, TicketStatus } from '@/types/ticket'

export const useTicketStore = defineStore('tickets', () => {
  const tickets = ref<Ticket[]>([])
  const allUsers = ref<AuthUser[]>([])
  const total = ref(0)
  const page = ref(1)
  const limit = ref(10)
  const totalPages = ref(1)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const filters = reactive<{
    search: string
    status: TicketStatus | 'all'
    priority: TicketPriority | 'all'
    assigneeId: string
  }>({
    search: '',
    status: 'all',
    priority: 'all',
    assigneeId: 'all',
  })

  const userMap = computed(() => {
    const map = new Map<string, AuthUser>()
    for (const u of allUsers.value) {
      map.set(u.id, u)
    }
    return map
  })

  const assignees = computed(() =>
    allUsers.value.filter((u) => u.role === 'agent' || u.role === 'admin'),
  )

  const statusCounts = reactive<Record<TicketStatus | 'all', number>>({
    all: 0,
    open: 0,
    in_progress: 0,
    resolved: 0,
    closed: 0,
  })

  async function fetchStatusCounts() {
    try {
      const counts = await ticketService.getStatusCounts({
        search: filters.search,
        priority: filters.priority,
        assignee_id: filters.assigneeId === 'all' ? undefined : filters.assigneeId,
      })
      statusCounts.all = counts.all
      statusCounts.open = counts.open
      statusCounts.in_progress = counts.in_progress
      statusCounts.resolved = counts.resolved
      statusCounts.closed = counts.closed
    } catch {
      // Non-blocking: keep previous counts if count refresh fails
    }
  }

  async function fetchUsers() {
    try {
      allUsers.value = await ticketService.listUsers()
    } catch {
      // Non-blocking: assignee names fall back to IDs if users endpoint fails
    }
  }

  async function fetchTickets(targetPage = page.value) {
    loading.value = true
    error.value = null
    try {
      const res = await ticketService.list({
        search: filters.search,
        status: filters.status,
        priority: filters.priority,
        assignee_id: filters.assigneeId === 'all' ? undefined : filters.assigneeId,
        page: targetPage,
        limit: limit.value,
      })
      tickets.value = res.items ?? []
      total.value = res.total
      page.value = res.page
      limit.value = res.limit
      totalPages.value = Math.max(1, res.total_pages)
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Failed to load tickets'
    } finally {
      loading.value = false
    }
  }

  async function createTicket(payload: CreateTicketPayload): Promise<Ticket | null> {
    loading.value = true
    error.value = null
    try {
      const created = await ticketService.create(payload)
      return created
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Failed to create ticket'
      return null
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  function resetFilters() {
    filters.search = ''
    filters.status = 'all'
    filters.priority = 'all'
    filters.assigneeId = 'all'
    page.value = 1
    return Promise.all([fetchTickets(1), fetchStatusCounts()])
  }

  return {
    tickets,
    allUsers,
    assignees,
    userMap,
    total,
    page,
    limit,
    totalPages,
    loading,
    error,
    filters,
    statusCounts,
    fetchUsers,
    fetchTickets,
    fetchStatusCounts,
    createTicket,
    clearError,
    resetFilters,
  }
})
