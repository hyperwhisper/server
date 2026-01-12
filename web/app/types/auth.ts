export interface User {
  id: string
  username: string
  email: string
  first_name: string
  last_name: string
  user_type: 'admin' | 'user'
  created_at: string
}

export interface AuthResponse {
  user: User
  access_token: string
  expires_in: number
}

export interface SignUpPayload {
  username: string
  email: string
  password: string
  first_name?: string
  last_name?: string
}

export interface SignInPayload {
  identifier: string // email or username
  password: string
}

export interface TokenRefreshResponse {
  access_token: string
  expires_in: number
}

export interface ApiError {
  error: string
  details?: Record<string, string>
}
