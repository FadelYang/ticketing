export type AttachmentType = 'ticket' | 'comment'

export interface Attachment {
  id: string
  type: AttachmentType
  target_id: string
  uploader_id: string
  file_name: string
  content_type: string
  size_bytes: number
  created_at: string
}
