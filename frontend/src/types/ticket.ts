export type TicketStatus = 'open' | 'in_progress' | 'resolved' | 'closed'
export type TicketPriority = 'low' | 'medium' | 'high' | 'urgent'

export interface Ticket {
  id: string
  title: string
  description: string
  category: string
  priority: TicketPriority
  status: TicketStatus
  requester_id: string
  assignee_id: string | null
  created_at: string
  updated_at: string
  deadline_at: string
}

export interface TicketListResult {
  items: Ticket[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export type TicketStatusCounts = Record<TicketStatus | 'all', number>

export interface TicketFilter {
  search?: string
  status?: TicketStatus | 'all'
  priority?: TicketPriority | 'all'
  assignee_id?: string
  page?: number
  limit?: number
  sort?: 'newest' | 'oldest'
}

export interface CreateTicketPayload {
  title: string
  description: string
  category: string
  priority: TicketPriority
  requester_id: string
  assignee_id?: string | null
  deadline_at: string
}

export interface PatchTicketPayload {
  priority?: TicketPriority
  status?: TicketStatus
  assignee_id?: string | null
  comment?: string
}

export interface TicketComment {
  id: string
  ticket_id: string
  author_id: string
  body: string
  is_internal: boolean
  created_at: string
}

export interface CommentListResult {
  items: TicketComment[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface CommentFilter {
  sort?: 'newest' | 'oldest'
  page?: number
  limit?: number
}

export interface CreateCommentPayload {
  body: string
  is_internal?: boolean
}
