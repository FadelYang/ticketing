import { apiRequest } from '@/services/api'
import type { AuthUser } from '@/types/auth'
import type {
  CommentFilter,
  CommentListResult,
  CreateCommentPayload,
  CreateTicketPayload,
  PatchTicketPayload,
  Ticket,
  TicketComment,
  TicketFilter,
  TicketListResult,
  TicketStatusCounts,
} from '@/types/ticket'

export const ticketService = {
  list(filter: TicketFilter = {}): Promise<TicketListResult> {
    const params = new URLSearchParams()
    if (filter.search?.trim()) params.set('search', filter.search.trim())
    if (filter.status && filter.status !== 'all') params.set('status', filter.status)
    if (filter.priority && filter.priority !== 'all') params.set('priority', filter.priority)
    if (filter.assignee_id && filter.assignee_id !== 'all') {
      params.set('assignee_id', filter.assignee_id)
    }
    if (filter.page) params.set('page', String(filter.page))
    if (filter.limit) params.set('limit', String(filter.limit))
    if (filter.sort) params.set('sort', filter.sort)

    const qs = params.toString()
    return apiRequest<TicketListResult>(`/tickets/${qs ? `?${qs}` : ''}`)
  },

  getStatusCounts(
    filter: Omit<TicketFilter, 'status' | 'page' | 'limit' | 'sort'> = {},
  ): Promise<TicketStatusCounts> {
    const params = new URLSearchParams()
    if (filter.search?.trim()) params.set('search', filter.search.trim())
    if (filter.priority && filter.priority !== 'all') params.set('priority', filter.priority)
    if (filter.assignee_id && filter.assignee_id !== 'all') {
      params.set('assignee_id', filter.assignee_id)
    }

    const qs = params.toString()
    return apiRequest<TicketStatusCounts>(`/tickets/counts${qs ? `?${qs}` : ''}`)
  },

  getById(id: string): Promise<Ticket> {
    return apiRequest<Ticket>(`/tickets/${id}`)
  },

  create(payload: CreateTicketPayload): Promise<Ticket> {
    return apiRequest<Ticket>('/tickets/', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  patch(id: string, payload: PatchTicketPayload): Promise<Ticket> {
    return apiRequest<Ticket>(`/tickets/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
  },

  listUsers(): Promise<AuthUser[]> {
    return apiRequest<AuthUser[]>('/users/')
  },

  getComments(ticketId: string, filter: CommentFilter = {}): Promise<CommentListResult> {
    const params = new URLSearchParams()
    if (filter.sort) params.set('sort', filter.sort)
    if (filter.page) params.set('page', String(filter.page))
    if (filter.limit) params.set('limit', String(filter.limit))

    const qs = params.toString()
    return apiRequest<CommentListResult>(`/tickets/${ticketId}/comments/${qs ? `?${qs}` : ''}`)
  },

  createComment(ticketId: string, payload: CreateCommentPayload): Promise<TicketComment> {
    return apiRequest<TicketComment>(`/tickets/${ticketId}/comments/`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
}
