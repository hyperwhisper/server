export interface TrialAPIKey {
  id: string
  key_prefix: string
  device_fingerprint: string
  created_at: string
  expires_at: string
  last_used_at: string | null
  revoked_at: string | null
  total_sessions: number
  total_duration_seconds: number
}

export interface TrialUsageSummary {
  total_trial_keys: number
  active_trial_keys: number
  total_sessions: number
  total_duration_seconds: number
  total_bytes_sent: number
  period_start: string
  period_end: string
}

export interface TrialLimits {
  max_duration_seconds: number
  max_sessions: number
  max_session_duration_seconds: number
  expiry_days: number
  updated_at: string
}

export interface UpdateTrialLimitsRequest {
  max_duration_seconds: number
  max_sessions: number
  max_session_duration_seconds: number
  expiry_days: number
}
