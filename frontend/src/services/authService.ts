import { apiRequest } from '@/services/api'
import type {
  LoginPayload,
  LogoutResponse,
  RegisterPayload,
  TokenResponse,
} from '@/types/auth'

export const authService = {
  login(payload: LoginPayload): Promise<TokenResponse> {
    return apiRequest<TokenResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  register(payload: RegisterPayload): Promise<TokenResponse> {
    return apiRequest<TokenResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  logout(): Promise<LogoutResponse> {
    return apiRequest<LogoutResponse>('/auth/logout', {
      method: 'POST',
    })
  },
}
