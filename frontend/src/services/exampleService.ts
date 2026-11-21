import { apiRequest } from '@/services/api'
import type {
  CreateExamplePayload,
  ExampleItem,
  HealthStatus,
  UpdateExamplePayload,
} from '@/types/example'

export const exampleService = {
  checkHealth(): Promise<HealthStatus> {
    return apiRequest<HealthStatus>('/health')
  },

  list(search?: string): Promise<ExampleItem[]> {
    const params = new URLSearchParams()
    if (search?.trim()) params.set('search', search.trim())
    const queryString = params.toString()
    return apiRequest<ExampleItem[]>(`/examples${queryString ? `?${queryString}` : ''}`)
  },

  getById(id: string): Promise<ExampleItem> {
    return apiRequest<ExampleItem>(`/examples/${id}`)
  },

  create(payload: CreateExamplePayload): Promise<ExampleItem> {
    return apiRequest<ExampleItem>('/examples', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  update(id: string, payload: UpdateExamplePayload): Promise<ExampleItem> {
    return apiRequest<ExampleItem>(`/examples/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },

  remove(id: string): Promise<void> {
    return apiRequest<void>(`/examples/${id}`, {
      method: 'DELETE',
    })
  },
}
