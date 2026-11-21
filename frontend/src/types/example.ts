export interface ExampleItem {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface CreateExamplePayload {
  name: string
  description: string
}

export interface UpdateExamplePayload {
  name: string
  description: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}

export interface HealthStatus {
  status: string
  database: string
}
