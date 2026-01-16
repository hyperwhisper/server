package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"hyperwhisper/internal/auth"
	"hyperwhisper/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdminHandler handles admin endpoints
type AdminHandler struct {
	queries *sqlc.Queries
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{
		queries: sqlc.New(db),
	}
}

// Request types
type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type"`
}

type RevokeTokenRequest struct {
	TokenJTI string `json:"token_jti"`
	Reason   string `json:"reason"`
}

// Response types
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

type TokenResponse struct {
	ID            string  `json:"id"`
	TokenJTI      string  `json:"token_jti"`
	UserID        string  `json:"user_id"`
	IssuedAt      string  `json:"issued_at"`
	ExpiresAt     string  `json:"expires_at"`
	RevokedAt     *string `json:"revoked_at"`
	RevokedReason *string `json:"revoked_reason"`
}

// ========== USER MANAGEMENT ==========

// ListUsers returns a paginated list of users
func (h *AdminHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	ctx := context.Background()

	// Get total count
	total, err := h.queries.CountUsers(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Get users
	users, err := h.queries.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = toUserResponse(user)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       userResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// CreateUser creates a new user (admin only)
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "username, email, and password are required"})
	}

	// Validate user type
	if req.UserType == "" {
		req.UserType = "user"
	}
	if req.UserType != "user" && req.UserType != "admin" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user_type must be 'user' or 'admin'"})
	}

	// Validate password
	if err := auth.ValidatePassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "password validation failed",
			Details: map[string]string{"password": err.Error()},
		})
	}

	ctx := context.Background()

	// Check if email exists
	emailExists, err := h.queries.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}
	if emailExists {
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "email already taken",
			Details: map[string]string{"email": "this email is already registered"},
		})
	}

	// Check if username exists
	usernameExists, err := h.queries.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}
	if usernameExists {
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "username already taken",
			Details: map[string]string{"username": "this username is already taken"},
		})
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process password"})
	}

	// Create user
	user, err := h.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		UserType:     req.UserType,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
	}

	return c.JSON(http.StatusCreated, toUserResponse(user))
}

// DeleteUser deletes a user by ID
func (h *AdminHandler) DeleteUser(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	// Prevent self-deletion
	claims := auth.GetUserFromContext(c)
	if claims != nil && claims.UserID == userID {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "cannot delete your own account"})
	}

	ctx := context.Background()

	// Check if user exists
	_, err = h.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Delete user
	if err := h.queries.DeleteUser(ctx, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user deleted successfully"})
}

// ========== TOKEN MANAGEMENT ==========

// ListRefreshTokens returns a paginated list of all tokens
func (h *AdminHandler) ListRefreshTokens(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	ctx := context.Background()

	// Get total count
	total, err := h.queries.CountRefreshTokens(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Get tokens
	tokens, err := h.queries.ListRefreshTokens(ctx, sqlc.ListRefreshTokensParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Convert to response format
	tokenResponses := make([]TokenResponse, len(tokens))
	for i, token := range tokens {
		tokenResponses[i] = toTokenResponse(token)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       tokenResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// RevokeToken revokes a token by JTI
func (h *AdminHandler) RevokeToken(c echo.Context) error {
	var req RevokeTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.TokenJTI == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "token_jti is required"})
	}

	reason := req.Reason
	if reason == "" {
		reason = "admin"
	}

	ctx := context.Background()

	// Check if token exists
	token, err := h.queries.GetRefreshTokenByJTI(ctx, req.TokenJTI)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "token not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Check if already revoked
	if token.RevokedAt.Valid {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: "token already revoked"})
	}

	// Revoke the token
	err = h.queries.RevokeRefreshToken(ctx, sqlc.RevokeRefreshTokenParams{
		TokenJti:      req.TokenJTI,
		RevokedReason: sql.NullString{String: reason, Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "token revoked successfully"})
}

// RevokeUserRefreshTokens revokes all tokens for a user
func (h *AdminHandler) RevokeUserRefreshTokens(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	ctx := context.Background()

	// Check if user exists
	_, err = h.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Revoke all tokens for user
	err = h.queries.RevokeUserRefreshTokens(ctx, sqlc.RevokeUserRefreshTokensParams{
		UserID:        userID,
		RevokedReason: sql.NullString{String: "admin", Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke tokens"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user tokens revoked successfully"})
}

// CleanupTokens removes expired tokens
func (h *AdminHandler) CleanupTokens(c echo.Context) error {
	ctx := context.Background()

	if err := h.queries.CleanupExpiredRefreshTokens(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to cleanup tokens"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "expired tokens cleaned up successfully"})
}

// ========== DEEPGRAM ADMIN ENDPOINTS ==========

// AdminTranscriptionLogResponse extends TranscriptionLogResponse with user info
type AdminTranscriptionLogResponse struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	Email           string  `json:"email"`
	APIKeyName      string  `json:"api_key_name"`
	StartedAt       string  `json:"started_at"`
	EndedAt         *string `json:"ended_at"`
	DurationSeconds *string `json:"duration_seconds"`
	Status          string  `json:"status"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	BytesSent       int64   `json:"bytes_sent"`
}

// AdminAPIKeyResponse extends APIKeyResponse with user info
type AdminAPIKeyResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	KeyPrefix string  `json:"key_prefix"`
	CreatedAt string  `json:"created_at"`
	LastUsed  *string `json:"last_used_at"`
	RevokedAt *string `json:"revoked_at,omitempty"`
}

// SystemUsageSummaryResponse is the response for system-wide usage
type SystemUsageSummaryResponse struct {
	UniqueUsers          int64   `json:"unique_users"`
	TotalSessions        int64   `json:"total_sessions"`
	TotalDurationSeconds float64 `json:"total_duration_seconds"`
	TotalBytesSent       int64   `json:"total_bytes_sent"`
	PeriodStart          string  `json:"period_start"`
	PeriodEnd            string  `json:"period_end"`
}

// ListAllTranscriptionLogs returns all transcription logs (admin only)
func (h *AdminHandler) ListAllTranscriptionLogs(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	ctx := context.Background()

	total, err := h.queries.CountAllTranscriptionLogs(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	logs, err := h.queries.ListAllTranscriptionLogs(ctx, sqlc.ListAllTranscriptionLogsParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	responses := make([]AdminTranscriptionLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = toAdminTranscriptionLogResponse(log)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// ListAllAPIKeys returns all API keys with user info (admin only)
func (h *AdminHandler) ListAllAPIKeys(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	ctx := context.Background()

	total, err := h.queries.CountAllAPIKeys(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	keys, err := h.queries.ListAllAPIKeys(ctx, sqlc.ListAllAPIKeysParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	responses := make([]AdminAPIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = toAdminAPIKeyResponse(key)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// GetSystemUsageSummary returns system-wide usage statistics (admin only)
func (h *AdminHandler) GetSystemUsageSummary(c echo.Context) error {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	if startParam := c.QueryParam("start"); startParam != "" {
		if t, err := time.Parse(time.RFC3339, startParam); err == nil {
			startOfMonth = t
		}
	}
	if endParam := c.QueryParam("end"); endParam != "" {
		if t, err := time.Parse(time.RFC3339, endParam); err == nil {
			endOfMonth = t
		}
	}

	ctx := context.Background()

	summary, err := h.queries.GetSystemUsageSummary(ctx, sqlc.GetSystemUsageSummaryParams{
		StartDate: startOfMonth,
		EndDate:   endOfMonth,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Convert decimal string to float64
	durationFloat := parseDecimalStringAdmin(summary.TotalDurationSeconds)
	bytesSent := parseBytesSentAdmin(summary.TotalBytesSent)

	return c.JSON(http.StatusOK, SystemUsageSummaryResponse{
		UniqueUsers:          summary.UniqueUsers,
		TotalSessions:        summary.TotalSessions,
		TotalDurationSeconds: durationFloat,
		TotalBytesSent:       bytesSent,
		PeriodStart:          startOfMonth.Format(time.RFC3339),
		PeriodEnd:            endOfMonth.Format(time.RFC3339),
	})
}

// Helper function for admin transcription logs
func toAdminTranscriptionLogResponse(log sqlc.ListAllTranscriptionLogsRow) AdminTranscriptionLogResponse {
	resp := AdminTranscriptionLogResponse{
		ID:         log.ID.String(),
		UserID:     log.UserID.String(),
		Username:   log.Username,
		Email:      log.Email,
		APIKeyName: log.ApiKeyName,
		StartedAt:  log.StartedAt.Format(time.RFC3339),
		Status:     log.Status,
		BytesSent:  log.BytesSent,
	}

	if log.EndedAt.Valid {
		t := log.EndedAt.Time.Format(time.RFC3339)
		resp.EndedAt = &t
	}

	if log.DurationSeconds.Valid {
		s := log.DurationSeconds.String
		resp.DurationSeconds = &s
	}

	if log.ErrorMessage.Valid {
		resp.ErrorMessage = &log.ErrorMessage.String
	}

	return resp
}

// Helper function for admin API keys
func toAdminAPIKeyResponse(key sqlc.ListAllAPIKeysRow) AdminAPIKeyResponse {
	resp := AdminAPIKeyResponse{
		ID:        key.ID.String(),
		UserID:    key.UserID.String(),
		Username:  key.Username,
		Email:     key.Email,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		CreatedAt: key.CreatedAt.Time.Format(time.RFC3339),
	}

	if key.LastUsedAt.Valid {
		t := key.LastUsedAt.Time.Format(time.RFC3339)
		resp.LastUsed = &t
	}

	if key.RevokedAt.Valid {
		t := key.RevokedAt.Time.Format(time.RFC3339)
		resp.RevokedAt = &t
	}

	return resp
}

// Helper functions
func toTokenResponse(token sqlc.Token) TokenResponse {
	issuedAt := ""
	if token.IssuedAt.Valid {
		issuedAt = token.IssuedAt.Time.Format(time.RFC3339)
	}

	var revokedAt *string
	if token.RevokedAt.Valid {
		t := token.RevokedAt.Time.Format(time.RFC3339)
		revokedAt = &t
	}

	var revokedReason *string
	if token.RevokedReason.Valid {
		revokedReason = &token.RevokedReason.String
	}

	return TokenResponse{
		ID:            token.ID.String(),
		TokenJTI:      token.TokenJti,
		UserID:        token.UserID.String(),
		IssuedAt:      issuedAt,
		ExpiresAt:     token.ExpiresAt.Format(time.RFC3339),
		RevokedAt:     revokedAt,
		RevokedReason: revokedReason,
	}
}

// parseDecimalStringAdmin converts a decimal string to float64
func parseDecimalStringAdmin(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseBytesSentAdmin converts interface{} to int64
func parseBytesSentAdmin(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	case []byte:
		i, _ := strconv.ParseInt(string(val), 10, 64)
		return i
	default:
		return 0
	}
}

// ========== TRIAL ADMIN ENDPOINTS ==========

// TrialAPIKeyResponse is the response for trial API key admin queries
type TrialAPIKeyResponse struct {
	ID                   string  `json:"id"`
	KeyPrefix            string  `json:"key_prefix"`
	DeviceFingerprint    string  `json:"device_fingerprint"`
	CreatedAt            string  `json:"created_at"`
	ExpiresAt            string  `json:"expires_at"`
	LastUsedAt           *string `json:"last_used_at"`
	RevokedAt            *string `json:"revoked_at"`
	TotalSessions        int64   `json:"total_sessions"`
	TotalDurationSeconds float64 `json:"total_duration_seconds"`
}

// TrialUsageSummaryResponse is the response for trial usage summary
type TrialUsageSummaryResponse struct {
	TotalTrialKeys       int64   `json:"total_trial_keys"`
	ActiveTrialKeys      int64   `json:"active_trial_keys"`
	TotalSessions        int64   `json:"total_sessions"`
	TotalDurationSeconds float64 `json:"total_duration_seconds"`
	TotalBytesSent       int64   `json:"total_bytes_sent"`
	PeriodStart          string  `json:"period_start"`
	PeriodEnd            string  `json:"period_end"`
}

// TrialLimitsResponse is the response for trial limits
type TrialLimitsResponse struct {
	MaxDurationSeconds        int    `json:"max_duration_seconds"`
	MaxSessions               int    `json:"max_sessions"`
	MaxSessionDurationSeconds int    `json:"max_session_duration_seconds"`
	ExpiryDays                int    `json:"expiry_days"`
	UpdatedAt                 string `json:"updated_at"`
}

// UpdateTrialLimitsRequest is the request for updating trial limits
type UpdateTrialLimitsRequest struct {
	MaxDurationSeconds        int `json:"max_duration_seconds"`
	MaxSessions               int `json:"max_sessions"`
	MaxSessionDurationSeconds int `json:"max_session_duration_seconds"`
	ExpiryDays                int `json:"expiry_days"`
}

// ListTrialAPIKeys returns all trial API keys with usage stats (admin only)
func (h *AdminHandler) ListTrialAPIKeys(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	ctx := context.Background()

	total, err := h.queries.CountTrialAPIKeys(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	keys, err := h.queries.ListAllTrialAPIKeys(ctx, sqlc.ListAllTrialAPIKeysParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	responses := make([]TrialAPIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = toTrialAPIKeyResponse(key)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// GetTrialUsageSummary returns system-wide trial usage statistics (admin only)
func (h *AdminHandler) GetTrialUsageSummary(c echo.Context) error {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	if startParam := c.QueryParam("start"); startParam != "" {
		if t, err := time.Parse(time.RFC3339, startParam); err == nil {
			startOfMonth = t
		}
	}
	if endParam := c.QueryParam("end"); endParam != "" {
		if t, err := time.Parse(time.RFC3339, endParam); err == nil {
			endOfMonth = t
		}
	}

	ctx := context.Background()

	summary, err := h.queries.GetAllTrialUsageSummary(ctx, sqlc.GetAllTrialUsageSummaryParams{
		StartDate: startOfMonth,
		EndDate:   endOfMonth,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	durationFloat := parseDecimalStringAdmin(summary.TotalDurationSeconds)
	bytesSent := parseBytesSentAdmin(summary.TotalBytesSent)

	return c.JSON(http.StatusOK, TrialUsageSummaryResponse{
		TotalTrialKeys:       summary.TotalTrialKeys,
		ActiveTrialKeys:      summary.ActiveTrialKeys,
		TotalSessions:        summary.TotalSessions,
		TotalDurationSeconds: durationFloat,
		TotalBytesSent:       bytesSent,
		PeriodStart:          startOfMonth.Format(time.RFC3339),
		PeriodEnd:            endOfMonth.Format(time.RFC3339),
	})
}

// GetTrialLimits returns the current trial limits (admin only)
func (h *AdminHandler) GetTrialLimits(c echo.Context) error {
	ctx := context.Background()

	limits, err := h.queries.GetTrialLimits(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	return c.JSON(http.StatusOK, TrialLimitsResponse{
		MaxDurationSeconds:        int(limits.MaxDurationSeconds),
		MaxSessions:               int(limits.MaxSessions),
		MaxSessionDurationSeconds: int(limits.MaxSessionDurationSeconds),
		ExpiryDays:                int(limits.ExpiryDays),
		UpdatedAt:                 limits.UpdatedAt.Time.Format(time.RFC3339),
	})
}

// UpdateTrialLimits updates the trial limits (admin only)
func (h *AdminHandler) UpdateTrialLimits(c echo.Context) error {
	var req UpdateTrialLimitsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate limits
	if req.MaxDurationSeconds <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "max_duration_seconds must be positive"})
	}
	if req.MaxSessions <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "max_sessions must be positive"})
	}
	if req.MaxSessionDurationSeconds <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "max_session_duration_seconds must be positive"})
	}
	if req.ExpiryDays <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "expiry_days must be positive"})
	}

	ctx := context.Background()

	limits, err := h.queries.UpdateTrialLimits(ctx, sqlc.UpdateTrialLimitsParams{
		MaxDurationSeconds:        int32(req.MaxDurationSeconds),
		MaxSessions:               int32(req.MaxSessions),
		MaxSessionDurationSeconds: int32(req.MaxSessionDurationSeconds),
		ExpiryDays:                int32(req.ExpiryDays),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update limits"})
	}

	return c.JSON(http.StatusOK, TrialLimitsResponse{
		MaxDurationSeconds:        int(limits.MaxDurationSeconds),
		MaxSessions:               int(limits.MaxSessions),
		MaxSessionDurationSeconds: int(limits.MaxSessionDurationSeconds),
		ExpiryDays:                int(limits.ExpiryDays),
		UpdatedAt:                 limits.UpdatedAt.Time.Format(time.RFC3339),
	})
}

// RevokeTrialKey revokes a trial API key (admin only)
func (h *AdminHandler) RevokeTrialKey(c echo.Context) error {
	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid key ID"})
	}

	ctx := context.Background()

	// Check if key exists
	_, err = h.queries.GetTrialAPIKeyByID(ctx, keyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "trial key not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Revoke the key
	if err := h.queries.RevokeTrialAPIKey(ctx, keyID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke key"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "trial key revoked"})
}

// CleanupExpiredTrialKeys revokes all expired trial keys (admin only)
func (h *AdminHandler) CleanupExpiredTrialKeys(c echo.Context) error {
	ctx := context.Background()

	if err := h.queries.CleanupExpiredTrialKeys(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to cleanup expired keys"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "expired trial keys cleaned up"})
}

// UnrevokeTrialKey unrevokes a trial API key (admin only)
func (h *AdminHandler) UnrevokeTrialKey(c echo.Context) error {
	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid key ID"})
	}

	ctx := context.Background()

	// Check if key exists
	key, err := h.queries.GetTrialAPIKeyByID(ctx, keyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "trial key not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Check if key is actually revoked
	if !key.RevokedAt.Valid {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "trial key is not revoked"})
	}

	// Unrevoke the key
	if err := h.queries.UnrevokeTrialAPIKey(ctx, keyID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to unrevoke key"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "trial key unrevoked"})
}

// DeleteTrialKey permanently deletes a trial API key (admin only)
func (h *AdminHandler) DeleteTrialKey(c echo.Context) error {
	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid key ID"})
	}

	ctx := context.Background()

	// Check if key exists
	_, err = h.queries.GetTrialAPIKeyByID(ctx, keyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "trial key not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Delete the key (cascade will delete usage logs)
	if err := h.queries.DeleteTrialAPIKey(ctx, keyID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete key"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "trial key deleted"})
}

// Helper function for trial API key response
func toTrialAPIKeyResponse(key sqlc.ListAllTrialAPIKeysRow) TrialAPIKeyResponse {
	resp := TrialAPIKeyResponse{
		ID:                   key.ID.String(),
		KeyPrefix:            key.KeyPrefix,
		DeviceFingerprint:    key.DeviceFingerprint,
		CreatedAt:            key.CreatedAt.Time.Format(time.RFC3339),
		ExpiresAt:            key.ExpiresAt.Format(time.RFC3339),
		TotalSessions:        key.TotalSessions,
		TotalDurationSeconds: parseDecimalStringAdmin(key.TotalDurationSeconds),
	}

	if key.LastUsedAt.Valid {
		t := key.LastUsedAt.Time.Format(time.RFC3339)
		resp.LastUsedAt = &t
	}

	if key.RevokedAt.Valid {
		t := key.RevokedAt.Time.Format(time.RFC3339)
		resp.RevokedAt = &t
	}

	return resp
}
