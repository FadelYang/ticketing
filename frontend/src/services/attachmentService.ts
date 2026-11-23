import { ApiError, getStoredToken } from '@/services/api'
import type { Attachment, AttachmentType } from '@/types/attachment'
import type { ApiResponse } from '@/types/example'

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/$/, '')

export const attachmentService = {
  async upload(type: AttachmentType, targetId: string, files: File[]): Promise<Attachment[]> {
    if (files.length === 0) return []

    const formData = new FormData()
    formData.append('type', type)
    formData.append('target_id', targetId)
    for (const file of files) {
      formData.append('files', file)
    }

    const token = getStoredToken()
    const headers: Record<string, string> = {}
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    const response = await fetch(`${API_BASE_URL}/attachments/`, {
      method: 'POST',
      headers,
      body: formData,
    })

    let body: ApiResponse<Attachment[]> | null = null
    try {
      body = (await response.json()) as ApiResponse<Attachment[]>
    } catch {
      body = null
    }

    if (!response.ok || !body?.success) {
      throw new ApiError(body?.error || `Upload failed with status ${response.status}`, response.status)
    }

    return body.data ?? []
  },

  async list(type: AttachmentType, targetIds: string[]): Promise<Attachment[]> {
    const cleaned = targetIds.map((id) => id.trim()).filter(Boolean)
    if (cleaned.length === 0) return []

    const params = new URLSearchParams({
      type,
      target_ids: cleaned.join(','),
    })

    const token = getStoredToken()
    const headers: Record<string, string> = {}
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    const response = await fetch(`${API_BASE_URL}/attachments/?${params.toString()}`, {
      headers,
    })

    let body: ApiResponse<Attachment[]> | null = null
    try {
      body = (await response.json()) as ApiResponse<Attachment[]>
    } catch {
      body = null
    }

    if (!response.ok || !body?.success) {
      throw new ApiError(body?.error || `Failed to load attachments`, response.status)
    }

    return body.data ?? []
  },

  async fetchFileBlobUrl(attachmentId: string): Promise<string> {
    const token = getStoredToken()
    const headers: Record<string, string> = {}
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    const response = await fetch(`${API_BASE_URL}/attachments/${attachmentId}/file`, {
      headers,
    })

    if (!response.ok) {
      throw new ApiError(`Failed to load image (${response.status})`, response.status)
    }

    const blob = await response.blob()
    return URL.createObjectURL(blob)
  },
}
