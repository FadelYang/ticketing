import type { ApiResponse } from '@/types/example'

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/$/, '')

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  })

  if (response.status === 204) {
    return undefined as T
  }

  let body: ApiResponse<T> | null = null
  try {
    body = (await response.json()) as ApiResponse<T>
  } catch {
    body = null
  }

  if (!response.ok || !body?.success) {
    const message = body?.error || `Request failed with status ${response.status}`
    throw new ApiError(message, response.status)
  }

  return body.data as T
}
