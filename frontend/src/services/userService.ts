import { apiRequest } from '@/services/api'
import type { AuthUser, RegisterPayload, UserRole } from '@/types/auth'

export const userService = {
  list(search?: string): Promise<AuthUser[]> {
    const trimmed = search?.trim()
    const qs = trimmed ? `?search=${encodeURIComponent(trimmed)}` : ''
    return apiRequest<AuthUser[]>(`/users/${qs}`)
  },

  create(payload: RegisterPayload): Promise<AuthUser> {
    return apiRequest<AuthUser>('/users/', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  updateRole(id: string, role: UserRole): Promise<AuthUser> {
    return apiRequest<AuthUser>(`/users/${id}/role`, {
      method: 'PATCH',
      body: JSON.stringify({ role }),
    })
  },
}
