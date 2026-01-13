export interface APIKey {
  id: string
  name: string
  key_prefix: string
  created_at: string
  last_used_at: string | null
  revoked_at?: string | null
}

export interface APIKeyCreated extends APIKey {
  key: string // Full key, only shown once
}

export interface UsageSummary {
  total_sessions: number
  total_duration_seconds: number
  total_bytes_sent: number
  period_start: string
  period_end: string
}

export interface TranscriptionLog {
  id: string
  started_at: string
  ended_at: string | null
  duration_seconds: number | null
  status: 'active' | 'completed' | 'error' | 'timeout'
  error_message?: string
  deepgram_params: Record<string, string>
  bytes_sent: number
}

// Admin types
export interface AdminTranscriptionLog extends TranscriptionLog {
  user_id: string
  username: string
  email: string
  api_key_name: string
}

export interface AdminAPIKey extends APIKey {
  user_id: string
  username: string
  email: string
}

export interface SystemUsageSummary extends UsageSummary {
  unique_users: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  per_page: number
  total_pages: number
}
