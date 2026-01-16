package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"hyperwhisper/internal/auth"
	"hyperwhisper/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// DeepgramHandler handles Deepgram proxy and API key management
type DeepgramHandler struct {
	queries  *sqlc.Queries
	upgrader websocket.Upgrader
}

// NewDeepgramHandler creates a new Deepgram handler
func NewDeepgramHandler(db *sql.DB) *DeepgramHandler {
	return &DeepgramHandler{
		queries: sqlc.New(db),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in dev, restrict in production
				if os.Getenv("APP_ENV") == "dev" {
					return true
				}
				return checkAllowedOrigin(r)
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// ========== REQUEST/RESPONSE TYPES ==========

// CreateAPIKeyRequest is the request body for creating an API key
type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

// APIKeyResponse is the response for API key operations
type APIKeyResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	KeyPrefix string  `json:"key_prefix"`
	CreatedAt string  `json:"created_at"`
	LastUsed  *string `json:"last_used_at"`
	RevokedAt *string `json:"revoked_at,omitempty"`
}

// APIKeyCreatedResponse includes the full key (only shown once)
type APIKeyCreatedResponse struct {
	APIKeyResponse
	Key string `json:"key"` // Full key, only shown once on creation
}

// UsageSummaryResponse is the response for usage summary
type UsageSummaryResponse struct {
	TotalSessions        int64   `json:"total_sessions"`
	TotalDurationSeconds float64 `json:"total_duration_seconds"`
	TotalBytesSent       int64   `json:"total_bytes_sent"`
	PeriodStart          string  `json:"period_start"`
	PeriodEnd            string  `json:"period_end"`
}

// TranscriptionLogResponse is the response for transcription logs
type TranscriptionLogResponse struct {
	ID              string          `json:"id"`
	StartedAt       string          `json:"started_at"`
	EndedAt         *string         `json:"ended_at"`
	DurationSeconds *float64        `json:"duration_seconds"`
	Status          string          `json:"status"`
	ErrorMessage    *string         `json:"error_message,omitempty"`
	DeepgramParams  json.RawMessage `json:"deepgram_params"`
	BytesSent       int64           `json:"bytes_sent"`
}

// ========== API KEY MANAGEMENT ==========

// GenerateAPIKey creates a new API key for the authenticated user
func (h *DeepgramHandler) GenerateAPIKey(c echo.Context) error {
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
	}

	var req CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Name == "" {
		req.Name = "Default Key"
	}

	// Generate random API key: hw_live_<32 random hex chars>
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate key"})
	}

	keyRandom := hex.EncodeToString(randomBytes)
	fullKey := fmt.Sprintf("hw_live_%s", keyRandom)
	keyPrefix := fullKey[:12] // "hw_live_abcd"

	// Hash the key for storage
	keyHash := hashAPIKey(fullKey)

	ctx := context.Background()

	apiKey, err := h.queries.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		UserID:    claims.UserID,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Name:      req.Name,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create API key"})
	}

	return c.JSON(http.StatusCreated, APIKeyCreatedResponse{
		APIKeyResponse: toAPIKeyResponse(apiKey),
		Key:            fullKey, // Only time the full key is returned
	})
}

// ListAPIKeys returns all API keys for the authenticated user
func (h *DeepgramHandler) ListAPIKeys(c echo.Context) error {
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
	}

	page, perPage, offset := getPaginationParams(c)
	ctx := context.Background()

	total, err := h.queries.CountUserAPIKeys(ctx, claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	keys, err := h.queries.ListUserAPIKeys(ctx, sqlc.ListUserAPIKeysParams{
		UserID: claims.UserID,
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	responses := make([]APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = toAPIKeyResponse(key)
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	})
}

// RevokeAPIKey revokes an API key
func (h *DeepgramHandler) RevokeAPIKey(c echo.Context) error {
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
	}

	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid key ID"})
	}

	ctx := context.Background()

	err = h.queries.RevokeAPIKey(ctx, sqlc.RevokeAPIKeyParams{
		ID:     keyID,
		UserID: claims.UserID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke key"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "API key revoked"})
}

// ========== USAGE TRACKING ==========

// GetUsageSummary returns usage statistics for the authenticated user
func (h *DeepgramHandler) GetUsageSummary(c echo.Context) error {
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
	}

	// Default to current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// Allow custom date range via query params
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

	summary, err := h.queries.GetUserUsageSummary(ctx, sqlc.GetUserUsageSummaryParams{
		UserID:    claims.UserID,
		StartDate: startOfMonth,
		EndDate:   endOfMonth,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Convert decimal string to float64
	durationFloat := parseDecimalString(summary.TotalDurationSeconds)
	bytesSent := parseBytesSent(summary.TotalBytesSent)

	return c.JSON(http.StatusOK, UsageSummaryResponse{
		TotalSessions:        summary.TotalSessions,
		TotalDurationSeconds: durationFloat,
		TotalBytesSent:       bytesSent,
		PeriodStart:          startOfMonth.Format(time.RFC3339),
		PeriodEnd:            endOfMonth.Format(time.RFC3339),
	})
}

// ListTranscriptionLogs returns usage logs for the authenticated user
func (h *DeepgramHandler) ListTranscriptionLogs(c echo.Context) error {
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
	}

	page, perPage, offset := getPaginationParams(c)
	ctx := context.Background()

	total, err := h.queries.CountUserTranscriptionLogs(ctx, claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	logs, err := h.queries.ListUserTranscriptionLogs(ctx, sqlc.ListUserTranscriptionLogsParams{
		UserID: claims.UserID,
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	responses := make([]TranscriptionLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = toTranscriptionLogResponse(log)
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	})
}

// ========== WEBSOCKET PROXY ==========

// DeepgramProxy handles WebSocket connections and proxies to Deepgram
// This endpoint handles both regular API keys (hw_live_) and trial keys (hw_trial_)
func (h *DeepgramHandler) DeepgramProxy(c echo.Context) error {
	// Extract API key from query param or header
	apiKey := c.QueryParam("api_key")
	if apiKey == "" {
		apiKey = c.Request().Header.Get("X-API-Key")
	}
	if apiKey == "" {
		log.Printf("[Deepgram] No API key provided")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "API key required"})
	}

	// Check if this is a trial key - use the trial handler stored in context
	if IsTrialKey(apiKey) {
		log.Printf("[Deepgram] Detected trial key, routing to trial handler")
		trialHandler := c.Get("trial_handler")
		if trialHandler == nil {
			log.Printf("[Deepgram] Trial handler not configured")
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "trial handler not configured"})
		}
		return trialHandler.(*TrialHandler).TrialDeepgramProxy(c)
	}

	log.Printf("[Deepgram] API key received (prefix: %s...)", apiKey[:12])

	// Validate API key and get user
	ctx := context.Background()
	keyHash := hashAPIKey(apiKey)

	apiKeyRecord, err := h.queries.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[Deepgram] Invalid API key - not found in database")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid API key"})
		}
		log.Printf("[Deepgram] Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}
	log.Printf("[Deepgram] API key validated, user: %s", apiKeyRecord.UserID)

	// Update last used timestamp (async, don't block)
	go func() {
		_ = h.queries.UpdateAPIKeyLastUsed(context.Background(), apiKeyRecord.ID)
	}()

	// Extract Deepgram params from query string
	deepgramParams := extractDeepgramParams(c.Request().URL.Query())

	// Get Deepgram API key from environment
	deepgramAPIKey := os.Getenv("DEEPGRAM_API_KEY")
	if deepgramAPIKey == "" {
		log.Printf("[Deepgram] ERROR: DEEPGRAM_API_KEY not set in environment")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Deepgram not configured"})
	}
	log.Printf("[Deepgram] API key configured (length: %d)", len(deepgramAPIKey))

	// Create transcription log
	paramsJSON, _ := json.Marshal(deepgramParams)
	clientIP := c.RealIP()

	txLog, err := h.queries.CreateTranscriptionLog(ctx, sqlc.CreateTranscriptionLogParams{
		UserID:         apiKeyRecord.UserID,
		ApiKeyID:       apiKeyRecord.ID,
		DeepgramParams: paramsJSON,
		ClientIp:       sql.NullString{String: clientIP, Valid: clientIP != ""},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create log"})
	}

	// Upgrade to WebSocket
	clientConn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		_ = h.queries.UpdateTranscriptionLogError(ctx, sqlc.UpdateTranscriptionLogErrorParams{
			ID:           txLog.ID,
			ErrorMessage: sql.NullString{String: "websocket upgrade failed", Valid: true},
			BytesSent:    0,
		})
		return err
	}
	defer clientConn.Close()

	// Connect to Deepgram
	deepgramURL := buildDeepgramURL(deepgramParams)
	log.Printf("[Deepgram] Connecting to: %s", deepgramURL)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Token %s", deepgramAPIKey))

	deepgramConn, resp, err := dialer.Dial(deepgramURL, headers)
	if err != nil {
		log.Printf("[Deepgram] Connection failed: %v", err)
		if resp != nil {
			log.Printf("[Deepgram] Response status: %d", resp.StatusCode)
		}
		_ = h.queries.UpdateTranscriptionLogError(ctx, sqlc.UpdateTranscriptionLogErrorParams{
			ID:           txLog.ID,
			ErrorMessage: sql.NullString{String: fmt.Sprintf("deepgram connection failed: %v", err), Valid: true},
			BytesSent:    0,
		})
		_ = clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to connect to Deepgram"))
		return nil
	}
	defer deepgramConn.Close()
	log.Printf("[Deepgram] Connected successfully")

	// Create proxy session
	session := &proxySession{
		clientConn:   clientConn,
		deepgramConn: deepgramConn,
		logID:        txLog.ID,
		queries:      h.queries,
		bytesSent:    0,
		duration:     0,
	}

	// Start bidirectional proxy
	session.run()

	return nil
}

// DeepgramProxyDashboard handles WebSocket connections for dashboard users using JWT auth
// This endpoint doesn't require an API key and doesn't log to transcription_logs
// Rate limiting: max 5 minutes per session, max 10 sessions per hour per user
func (h *DeepgramHandler) DeepgramProxyDashboard(c echo.Context) error {
	// Get user from JWT (set by middleware)
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		log.Printf("[Deepgram Dashboard] No JWT claims found")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}
	log.Printf("[Deepgram Dashboard] User authenticated: %s", claims.UserID)

	// Extract Deepgram params from query string
	deepgramParams := extractDeepgramParams(c.Request().URL.Query())

	// Get Deepgram API key from environment
	deepgramAPIKey := os.Getenv("DEEPGRAM_API_KEY")
	if deepgramAPIKey == "" {
		log.Printf("[Deepgram Dashboard] ERROR: DEEPGRAM_API_KEY not set in environment")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Deepgram not configured"})
	}

	// Upgrade to WebSocket
	clientConn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("[Deepgram Dashboard] WebSocket upgrade failed: %v", err)
		return err
	}
	defer clientConn.Close()

	// Connect to Deepgram
	deepgramURL := buildDeepgramURL(deepgramParams)
	log.Printf("[Deepgram Dashboard] Connecting to: %s", deepgramURL)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Token %s", deepgramAPIKey))

	deepgramConn, resp, err := dialer.Dial(deepgramURL, headers)
	if err != nil {
		log.Printf("[Deepgram Dashboard] Connection failed: %v", err)
		if resp != nil {
			log.Printf("[Deepgram Dashboard] Response status: %d", resp.StatusCode)
		}
		_ = clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to connect to Deepgram"))
		return nil
	}
	defer deepgramConn.Close()
	log.Printf("[Deepgram Dashboard] Connected successfully")

	// Create a simple proxy session (no logging)
	dashboardSession := &dashboardProxySession{
		clientConn:   clientConn,
		deepgramConn: deepgramConn,
		userID:       claims.UserID.String(),
		maxDuration:  5 * time.Minute, // Max 5 minutes per session
		startTime:    time.Now(),
	}

	// Start bidirectional proxy
	dashboardSession.run()

	return nil
}

// dashboardProxySession manages a dashboard WebSocket proxy session (no logging)
type dashboardProxySession struct {
	clientConn   *websocket.Conn
	deepgramConn *websocket.Conn
	userID       string
	maxDuration  time.Duration
	startTime    time.Time

	mu     sync.Mutex
	closed bool
}

func (s *dashboardProxySession) run() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Set up timeout
	timeout := time.AfterFunc(s.maxDuration, func() {
		log.Printf("[Deepgram Dashboard] Session timeout reached for user %s", s.userID)
		s.close()
	})
	defer timeout.Stop()

	// Client -> Deepgram (audio data)
	go func() {
		defer wg.Done()
		s.proxyClientToDeepgram()
	}()

	// Deepgram -> Client (transcriptions)
	go func() {
		defer wg.Done()
		s.proxyDeepgramToClient()
	}()

	wg.Wait()
}

func (s *dashboardProxySession) proxyClientToDeepgram() {
	for {
		messageType, data, err := s.clientConn.ReadMessage()
		if err != nil {
			log.Printf("[Deepgram Dashboard] Client read error: %v", err)
			_ = s.deepgramConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"CloseStream"}`))
			return
		}

		if err := s.deepgramConn.WriteMessage(messageType, data); err != nil {
			log.Printf("[Deepgram Dashboard] Error forwarding to Deepgram: %v", err)
			return
		}
	}
}

func (s *dashboardProxySession) proxyDeepgramToClient() {
	for {
		messageType, data, err := s.deepgramConn.ReadMessage()
		if err != nil {
			log.Printf("[Deepgram Dashboard] Deepgram read error: %v", err)
			return
		}

		if err := s.clientConn.WriteMessage(messageType, data); err != nil {
			log.Printf("[Deepgram Dashboard] Error forwarding to client: %v", err)
			return
		}
	}
}

func (s *dashboardProxySession) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	_ = s.clientConn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Session time limit reached"))
	s.clientConn.Close()
	s.deepgramConn.Close()
}

// proxySession manages a single WebSocket proxy session
type proxySession struct {
	clientConn   *websocket.Conn
	deepgramConn *websocket.Conn
	logID        uuid.UUID
	queries      *sqlc.Queries

	mu        sync.Mutex
	bytesSent int64
	duration  float64
	closed    bool
}

func (s *proxySession) run() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Deepgram (audio data)
	go func() {
		defer wg.Done()
		s.proxyClientToDeepgram()
	}()

	// Deepgram -> Client (transcriptions)
	go func() {
		defer wg.Done()
		s.proxyDeepgramToClient()
	}()

	wg.Wait()

	// Update final log
	s.finalize()
}

func (s *proxySession) proxyClientToDeepgram() {
	for {
		messageType, data, err := s.clientConn.ReadMessage()
		if err != nil {
			log.Printf("[Deepgram] Client read error: %v", err)
			// Client disconnected - send CloseStream to Deepgram
			_ = s.deepgramConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"CloseStream"}`))
			return
		}

		// Track bytes sent (only for binary audio data)
		if messageType == websocket.BinaryMessage {
			s.mu.Lock()
			s.bytesSent += int64(len(data))
			s.mu.Unlock()
			log.Printf("[Deepgram] Sent %d bytes of audio to Deepgram (total: %d)", len(data), s.bytesSent)
		} else {
			log.Printf("[Deepgram] Client sent text message: %s", string(data))
		}

		// Forward to Deepgram
		if err := s.deepgramConn.WriteMessage(messageType, data); err != nil {
			log.Printf("[Deepgram] Error forwarding to Deepgram: %v", err)
			return
		}
	}
}

func (s *proxySession) proxyDeepgramToClient() {
	clientClosed := false

	for {
		messageType, data, err := s.deepgramConn.ReadMessage()
		if err != nil {
			log.Printf("[Deepgram] Deepgram read error: %v", err)
			return
		}

		// Parse Deepgram response to extract duration from final metadata
		if messageType == websocket.TextMessage {
			log.Printf("[Deepgram] Received from Deepgram: %s", string(data))
			s.extractDurationFromResponse(data)

			// Check if this is the final metadata (Deepgram closes after this)
			var msg struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(data, &msg) == nil && msg.Type == "Metadata" {
				// This could be the final metadata after CloseStream
				// Try to forward but don't exit if it fails
				if !clientClosed {
					if err := s.clientConn.WriteMessage(messageType, data); err != nil {
						log.Printf("[Deepgram] Client closed, but captured final metadata")
						clientClosed = true
					}
				}
				continue
			}
		}

		// Forward to client (if still connected)
		if !clientClosed {
			if err := s.clientConn.WriteMessage(messageType, data); err != nil {
				log.Printf("[Deepgram] Error forwarding to client: %v", err)
				clientClosed = true
				// Don't return - keep reading from Deepgram to get final metadata
			}
		}
	}
}

func (s *proxySession) extractDurationFromResponse(data []byte) {
	// Deepgram sends duration in Metadata messages
	// The final Metadata message (after CloseStream) contains the total duration
	var response struct {
		Type     string  `json:"type"`
		Duration float64 `json:"duration"`
		Metadata *struct {
			Duration float64 `json:"duration"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &response); err == nil {
		// Check for duration at top level (final metadata message)
		if response.Type == "Metadata" && response.Duration > 0 {
			s.mu.Lock()
			s.duration = response.Duration
			s.mu.Unlock()
			log.Printf("[Deepgram] Duration extracted from Metadata: %.3f seconds", response.Duration)
		}
		// Also check nested metadata (in Results messages)
		if response.Metadata != nil && response.Metadata.Duration > 0 {
			s.mu.Lock()
			s.duration = response.Metadata.Duration
			s.mu.Unlock()
			log.Printf("[Deepgram] Duration extracted from nested metadata: %.3f seconds", response.Metadata.Duration)
		}
	}
}

func (s *proxySession) finalize() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	log.Printf("[Deepgram] Finalizing session - duration: %.3f, bytes: %d", s.duration, s.bytesSent)

	ctx := context.Background()

	if s.duration > 0 {
		// Convert float64 to pgtype.Numeric
		durationStr := fmt.Sprintf("%.3f", s.duration)
		log.Printf("[Deepgram] Updating log as completed with duration: %s", durationStr)
		_ = s.queries.UpdateTranscriptionLogComplete(ctx, sqlc.UpdateTranscriptionLogCompleteParams{
			ID:              s.logID,
			DurationSeconds: stringToNumeric(durationStr),
			BytesSent:       s.bytesSent,
		})
	} else {
		// No duration means possibly a timeout or error
		log.Printf("[Deepgram] Updating log as timeout (no duration captured)")
		_ = s.queries.UpdateTranscriptionLogTimeout(ctx, sqlc.UpdateTranscriptionLogTimeoutParams{
			ID:        s.logID,
			BytesSent: s.bytesSent,
		})
	}
}

// ========== HELPER FUNCTIONS ==========

func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func extractDeepgramParams(query url.Values) map[string]string {
	params := make(map[string]string)

	// Whitelist of allowed Deepgram parameters
	allowedParams := []string{
		"model", "language", "encoding", "sample_rate", "channels",
		"punctuate", "diarize", "smart_format", "interim_results",
		"utterances", "vad_events", "filler_words", "multichannel",
		"alternatives", "numerals", "profanity_filter", "redact",
		"search", "replace", "keywords", "endpointing", "tier",
		"detect_entities", "dictation", "utterance_end_ms", "version",
	}

	for _, param := range allowedParams {
		if value := query.Get(param); value != "" {
			params[param] = value
		}
	}

	return params
}

func buildDeepgramURL(params map[string]string) string {
	base := "wss://api.deepgram.com/v1/listen"

	if len(params) == 0 {
		return base
	}

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	return base + "?" + query.Encode()
}

func checkAllowedOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	// Add your allowed origins here
	allowedOrigins := []string{
		"https://hyperwhisper.dev",
		"https://www.hyperwhisper.dev",
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func getPaginationParams(c echo.Context) (page, perPage, offset int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ = strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset = (page - 1) * perPage
	return
}

func calculateTotalPages(total int64, perPage int) int {
	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}
	return pages
}

func toAPIKeyResponse(key sqlc.ApiKey) APIKeyResponse {
	resp := APIKeyResponse{
		ID:        key.ID.String(),
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

func toTranscriptionLogResponse(log sqlc.TranscriptionLog) TranscriptionLogResponse {
	resp := TranscriptionLogResponse{
		ID:             log.ID.String(),
		StartedAt:      log.StartedAt.Format(time.RFC3339),
		Status:         log.Status,
		DeepgramParams: log.DeepgramParams,
		BytesSent:      log.BytesSent,
	}

	if log.EndedAt.Valid {
		t := log.EndedAt.Time.Format(time.RFC3339)
		resp.EndedAt = &t
	}

	if log.DurationSeconds.Valid {
		f := parseDecimalString(log.DurationSeconds.String)
		resp.DurationSeconds = &f
	}

	if log.ErrorMessage.Valid {
		resp.ErrorMessage = &log.ErrorMessage.String
	}

	return resp
}

// stringToNumeric converts a string to sql.NullString for decimal fields
func stringToNumeric(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// parseDecimalString converts a decimal string to float64
func parseDecimalString(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseBytesSent converts interface{} to int64
func parseBytesSent(v interface{}) int64 {
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
		log.Printf("[Usage] parseBytesSent: unknown type %T with value %v", v, v)
		return 0
	}
}
