export type UserRole = 'requester' | 'agent' | 'admin'

export interface AuthUser {
  id: string
  name: string
  email: string
  phonenumber: string
  role: UserRole
  created_at: string
  updated_at: string
  created_by: string | null
}

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  name: string
  email: string
  phonenumber: string
  password: string
  role?: UserRole
}

export interface TokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: AuthUser
}

export interface LogoutResponse {
  message: string
}
