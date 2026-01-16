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
	"os"
	"strconv"
	"sync"
	"time"

	"hyperwhisper/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// TrialHandler handles trial API key endpoints
type TrialHandler struct {
	queries  *sqlc.Queries
	upgrader websocket.Upgrader
}

// NewTrialHandler creates a new trial handler
func NewTrialHandler(db *sql.DB) *TrialHandler {
	return &TrialHandler{
		queries: sqlc.New(db),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
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

// ProvisionTrialKeyRequest is the request body for provisioning a trial key
type ProvisionTrialKeyRequest struct {
	DeviceFingerprint string `json:"device_fingerprint"`
}

// TrialKeyResponse is the response for trial key operations
type TrialKeyResponse struct {
	Key                      string  `json:"key,omitempty"` // Only returned on first provision
	KeyPrefix                string  `json:"key_prefix"`
	RemainingDurationSeconds float64 `json:"remaining_duration_seconds"`
	RemainingSessions        int64   `json:"remaining_sessions"`
	MaxSessionDuration       int     `json:"max_session_duration_seconds"`
	ExpiresAt                string  `json:"expires_at"`
	QuotaExceeded            bool    `json:"quota_exceeded"`
	Expired                  bool    `json:"expired"`
}

// TrialUsageResponse is the response for trial usage queries
type TrialUsageResponse struct {
	TotalDurationSeconds     float64 `json:"total_duration_seconds"`
	TotalSessions            int64   `json:"total_sessions"`
	RemainingDurationSeconds float64 `json:"remaining_duration_seconds"`
	RemainingSessions        int64   `json:"remaining_sessions"`
	MaxDurationSeconds       int     `json:"max_duration_seconds"`
	MaxSessions              int     `json:"max_sessions"`
	MaxSessionDuration       int     `json:"max_session_duration_seconds"`
	QuotaExceeded            bool    `json:"quota_exceeded"`
}

// TrialStatusResponse is the response for trial status endpoint
type TrialStatusResponse struct {
	Active                   bool    `json:"active"`
	RemainingDurationSeconds float64 `json:"remaining_duration_seconds"`
	RemainingSessions        int64   `json:"remaining_sessions"`
	ExpiresAt                string  `json:"expires_at"`
	Expired                  bool    `json:"expired"`
	QuotaExceeded            bool    `json:"quota_exceeded"`
	UpgradeURL               string  `json:"upgrade_url,omitempty"`
}

// ========== TRIAL KEY PROVISIONING ==========

// ProvisionTrialKey creates or returns a trial key for a device
func (h *TrialHandler) ProvisionTrialKey(c echo.Context) error {
	var req ProvisionTrialKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.DeviceFingerprint == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "device_fingerprint is required"})
	}

	ctx := context.Background()

	// Get trial limits
	limits, err := h.queries.GetTrialLimits(ctx)
	if err != nil {
		log.Printf("[Trial] Failed to get trial limits: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get trial limits"})
	}

	// Check if a trial key already exists for this fingerprint
	existingKey, err := h.queries.GetTrialAPIKeyByFingerprint(ctx, req.DeviceFingerprint)
	if err == nil {
		// Key exists, return usage info
		return h.returnExistingTrialKey(c, ctx, existingKey, limits)
	}

	if err != sql.ErrNoRows {
		log.Printf("[Trial] Database error checking fingerprint: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Generate new trial API key: hw_trial_<32 random hex chars>
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate key"})
	}

	keyRandom := hex.EncodeToString(randomBytes)
	fullKey := fmt.Sprintf("hw_trial_%s", keyRandom)
	keyPrefix := fullKey[:16] // "hw_trial_ab12cd34"

	// Hash the key for storage
	keyHash := hashTrialAPIKey(fullKey)

	// Calculate expiration
	expiresAt := time.Now().AddDate(0, 0, int(limits.ExpiryDays))

	trialKey, err := h.queries.CreateTrialAPIKey(ctx, sqlc.CreateTrialAPIKeyParams{
		KeyHash:           keyHash,
		KeyPrefix:         keyPrefix,
		DeviceFingerprint: req.DeviceFingerprint,
		ExpiresAt:         expiresAt,
	})
	if err != nil {
		log.Printf("[Trial] Failed to create trial key: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create trial key"})
	}

	log.Printf("[Trial] Created new trial key for fingerprint: %s (prefix: %s)", req.DeviceFingerprint[:8], keyPrefix)

	return c.JSON(http.StatusCreated, TrialKeyResponse{
		Key:                      fullKey, // Only returned on creation
		KeyPrefix:                keyPrefix,
		RemainingDurationSeconds: float64(limits.MaxDurationSeconds),
		RemainingSessions:        int64(limits.MaxSessions),
		MaxSessionDuration:       int(limits.MaxSessionDurationSeconds),
		ExpiresAt:                trialKey.ExpiresAt.Format(time.RFC3339),
		QuotaExceeded:            false,
		Expired:                  false,
	})
}

// returnExistingTrialKey regenerates and returns the key for an existing trial
func (h *TrialHandler) returnExistingTrialKey(c echo.Context, ctx context.Context, key sqlc.TrialApiKey, limits sqlc.TrialLimit) error {
	// Check if key is expired
	expired := time.Now().After(key.ExpiresAt)

	// Check if key is revoked
	if key.RevokedAt.Valid {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "trial key revoked",
			Details: map[string]string{"upgrade_url": getUpgradeURL()},
		})
	}

	// Generate a new key for this device (since we can't retrieve the hashed one)
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate key"})
	}

	keyRandom := hex.EncodeToString(randomBytes)
	fullKey := fmt.Sprintf("hw_trial_%s", keyRandom)
	keyPrefix := fullKey[:16]
	keyHash := hashTrialAPIKey(fullKey)

	// Update the key hash in the database
	updatedKey, err := h.queries.RegenerateTrialAPIKey(ctx, sqlc.RegenerateTrialAPIKeyParams{
		ID:        key.ID,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
	})
	if err != nil {
		log.Printf("[Trial] Failed to regenerate key: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to regenerate key"})
	}

	log.Printf("[Trial] Regenerated trial key for fingerprint (prefix: %s)", keyPrefix)

	// Get usage summary
	summary, err := h.queries.GetTrialUsageSummary(ctx, key.ID)
	if err != nil {
		log.Printf("[Trial] Failed to get usage summary: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get usage"})
	}

	usedDuration := parseDecimalString(summary.TotalDurationSeconds)
	remainingDuration := float64(limits.MaxDurationSeconds) - usedDuration
	if remainingDuration < 0 {
		remainingDuration = 0
	}

	remainingSessions := int64(limits.MaxSessions) - summary.TotalSessions
	if remainingSessions < 0 {
		remainingSessions = 0
	}

	quotaExceeded := remainingDuration <= 0 || remainingSessions <= 0

	return c.JSON(http.StatusOK, TrialKeyResponse{
		Key:                      fullKey, // Return the regenerated key
		KeyPrefix:                updatedKey.KeyPrefix,
		RemainingDurationSeconds: remainingDuration,
		RemainingSessions:        remainingSessions,
		MaxSessionDuration:       int(limits.MaxSessionDurationSeconds),
		ExpiresAt:                updatedKey.ExpiresAt.Format(time.RFC3339),
		QuotaExceeded:            quotaExceeded,
		Expired:                  expired,
	})
}

// GetTrialUsage returns usage statistics for a trial key
func (h *TrialHandler) GetTrialUsage(c echo.Context) error {
	// Get trial key from query param or header
	apiKey := c.QueryParam("api_key")
	if apiKey == "" {
		apiKey = c.Request().Header.Get("X-API-Key")
	}
	if apiKey == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "api_key required"})
	}

	ctx := context.Background()

	// Validate trial key
	keyHash := hashTrialAPIKey(apiKey)
	trialKey, err := h.queries.GetTrialAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid trial key"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Get trial limits
	limits, err := h.queries.GetTrialLimits(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get limits"})
	}

	// Get usage summary
	summary, err := h.queries.GetTrialUsageSummary(ctx, trialKey.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get usage"})
	}

	usedDuration := parseDecimalString(summary.TotalDurationSeconds)
	remainingDuration := float64(limits.MaxDurationSeconds) - usedDuration
	if remainingDuration < 0 {
		remainingDuration = 0
	}

	remainingSessions := int64(limits.MaxSessions) - summary.TotalSessions
	if remainingSessions < 0 {
		remainingSessions = 0
	}

	quotaExceeded := remainingDuration <= 0 || remainingSessions <= 0

	return c.JSON(http.StatusOK, TrialUsageResponse{
		TotalDurationSeconds:     usedDuration,
		TotalSessions:            summary.TotalSessions,
		RemainingDurationSeconds: remainingDuration,
		RemainingSessions:        remainingSessions,
		MaxDurationSeconds:       int(limits.MaxDurationSeconds),
		MaxSessions:              int(limits.MaxSessions),
		MaxSessionDuration:       int(limits.MaxSessionDurationSeconds),
		QuotaExceeded:            quotaExceeded,
	})
}

// GetTrialStatus returns the status of a trial key with upgrade prompt
func (h *TrialHandler) GetTrialStatus(c echo.Context) error {
	// Get trial key from query param or header
	apiKey := c.QueryParam("api_key")
	if apiKey == "" {
		apiKey = c.Request().Header.Get("X-API-Key")
	}
	if apiKey == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "api_key required"})
	}

	ctx := context.Background()

	// Validate trial key
	keyHash := hashTrialAPIKey(apiKey)
	trialKey, err := h.queries.GetTrialAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid trial key"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Check if expired
	expired := time.Now().After(trialKey.ExpiresAt)

	// Get trial limits
	limits, err := h.queries.GetTrialLimits(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get limits"})
	}

	// Get usage summary
	summary, err := h.queries.GetTrialUsageSummary(ctx, trialKey.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get usage"})
	}

	usedDuration := parseDecimalString(summary.TotalDurationSeconds)
	remainingDuration := float64(limits.MaxDurationSeconds) - usedDuration
	if remainingDuration < 0 {
		remainingDuration = 0
	}

	remainingSessions := int64(limits.MaxSessions) - summary.TotalSessions
	if remainingSessions < 0 {
		remainingSessions = 0
	}

	quotaExceeded := remainingDuration <= 0 || remainingSessions <= 0

	response := TrialStatusResponse{
		Active:                   !expired && !quotaExceeded && !trialKey.RevokedAt.Valid,
		RemainingDurationSeconds: remainingDuration,
		RemainingSessions:        remainingSessions,
		ExpiresAt:                trialKey.ExpiresAt.Format(time.RFC3339),
		Expired:                  expired,
		QuotaExceeded:            quotaExceeded,
	}

	// Add upgrade URL if quota exceeded or expired
	if quotaExceeded || expired {
		response.UpgradeURL = getUpgradeURL()
	}

	return c.JSON(http.StatusOK, response)
}

// ========== TRIAL WEBSOCKET PROXY ==========

// TrialDeepgramProxy handles WebSocket connections for trial users
func (h *TrialHandler) TrialDeepgramProxy(c echo.Context) error {
	// Extract trial API key from query param or header
	apiKey := c.QueryParam("api_key")
	if apiKey == "" {
		apiKey = c.Request().Header.Get("X-API-Key")
	}
	if apiKey == "" {
		log.Printf("[Trial Deepgram] No API key provided")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "API key required"})
	}
	log.Printf("[Trial Deepgram] API key received (prefix: %s...)", apiKey[:16])

	ctx := context.Background()

	// Validate trial API key
	keyHash := hashTrialAPIKey(apiKey)
	trialKey, err := h.queries.GetTrialAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[Trial Deepgram] Invalid trial API key - not found")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid trial key"})
		}
		log.Printf("[Trial Deepgram] Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Check if key is expired
	if time.Now().After(trialKey.ExpiresAt) {
		log.Printf("[Trial Deepgram] Trial key expired")
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "trial key expired",
			Details: map[string]string{"upgrade_url": getUpgradeURL()},
		})
	}

	// Check if key is revoked
	if trialKey.RevokedAt.Valid {
		log.Printf("[Trial Deepgram] Trial key revoked")
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "trial key revoked",
			Details: map[string]string{"upgrade_url": getUpgradeURL()},
		})
	}

	// Get trial limits
	limits, err := h.queries.GetTrialLimits(ctx)
	if err != nil {
		log.Printf("[Trial Deepgram] Failed to get limits: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get limits"})
	}

	// Get current usage
	summary, err := h.queries.GetTrialUsageSummary(ctx, trialKey.ID)
	if err != nil {
		log.Printf("[Trial Deepgram] Failed to get usage: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get usage"})
	}

	usedDuration := parseDecimalString(summary.TotalDurationSeconds)
	remainingDuration := float64(limits.MaxDurationSeconds) - usedDuration
	remainingSessions := int64(limits.MaxSessions) - summary.TotalSessions

	// Check quota
	if remainingDuration <= 0 || remainingSessions <= 0 {
		log.Printf("[Trial Deepgram] Quota exceeded - duration: %.2f, sessions: %d", remainingDuration, remainingSessions)
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "trial quota exceeded",
			Details: map[string]string{"upgrade_url": getUpgradeURL()},
		})
	}

	// Calculate session timeout: min(per-session limit, remaining quota)
	sessionTimeout := time.Duration(limits.MaxSessionDurationSeconds) * time.Second
	if remainingDuration < float64(limits.MaxSessionDurationSeconds) {
		sessionTimeout = time.Duration(remainingDuration) * time.Second
	}
	log.Printf("[Trial Deepgram] Session timeout: %v (remaining: %.2fs)", sessionTimeout, remainingDuration)

	// Update last used timestamp (async)
	go func() {
		_ = h.queries.UpdateTrialAPIKeyLastUsed(context.Background(), trialKey.ID)
	}()

	// Extract Deepgram params from query string
	deepgramParams := extractDeepgramParams(c.Request().URL.Query())

	// Get Deepgram API key from environment
	deepgramAPIKey := os.Getenv("DEEPGRAM_API_KEY")
	if deepgramAPIKey == "" {
		log.Printf("[Trial Deepgram] ERROR: DEEPGRAM_API_KEY not set")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Deepgram not configured"})
	}

	// Create usage log
	paramsJSON, _ := json.Marshal(deepgramParams)
	clientIP := c.RealIP()

	usageLog, err := h.queries.CreateTrialUsageLog(ctx, sqlc.CreateTrialUsageLogParams{
		TrialKeyID:     trialKey.ID,
		DeepgramParams: paramsJSON,
		ClientIp:       sql.NullString{String: clientIP, Valid: clientIP != ""},
	})
	if err != nil {
		log.Printf("[Trial Deepgram] Failed to create usage log: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create log"})
	}

	// Upgrade to WebSocket
	clientConn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		_ = h.queries.UpdateTrialUsageError(ctx, sqlc.UpdateTrialUsageErrorParams{
			ID:           usageLog.ID,
			ErrorMessage: sql.NullString{String: "websocket upgrade failed", Valid: true},
			BytesSent:    0,
		})
		return err
	}
	defer clientConn.Close()

	// Connect to Deepgram
	deepgramURL := buildDeepgramURL(deepgramParams)
	log.Printf("[Trial Deepgram] Connecting to: %s", deepgramURL)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Token %s", deepgramAPIKey))

	deepgramConn, resp, err := dialer.Dial(deepgramURL, headers)
	if err != nil {
		log.Printf("[Trial Deepgram] Connection failed: %v", err)
		if resp != nil {
			log.Printf("[Trial Deepgram] Response status: %d", resp.StatusCode)
		}
		_ = h.queries.UpdateTrialUsageError(ctx, sqlc.UpdateTrialUsageErrorParams{
			ID:           usageLog.ID,
			ErrorMessage: sql.NullString{String: fmt.Sprintf("deepgram connection failed: %v", err), Valid: true},
			BytesSent:    0,
		})
		_ = clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to connect to Deepgram"))
		return nil
	}
	defer deepgramConn.Close()
	log.Printf("[Trial Deepgram] Connected successfully")

	// Create trial proxy session
	session := &trialProxySession{
		clientConn:     clientConn,
		deepgramConn:   deepgramConn,
		logID:          usageLog.ID,
		queries:        h.queries,
		bytesSent:      0,
		duration:       0,
		maxDuration:    sessionTimeout,
		startTime:      time.Now(),
		trialKeyPrefix: trialKey.KeyPrefix,
	}

	// Start bidirectional proxy with timeout
	session.run()

	return nil
}

// trialProxySession manages a trial WebSocket proxy session with timeout
type trialProxySession struct {
	clientConn     *websocket.Conn
	deepgramConn   *websocket.Conn
	logID          uuid.UUID
	queries        *sqlc.Queries
	trialKeyPrefix string

	mu          sync.Mutex
	bytesSent   int64
	duration    float64
	maxDuration time.Duration
	startTime   time.Time
	closed      bool
}

func (s *trialProxySession) run() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Set up session timeout
	timeout := time.AfterFunc(s.maxDuration, func() {
		log.Printf("[Trial Deepgram] Session timeout reached for %s", s.trialKeyPrefix)
		s.closeWithTimeout()
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

	// Update final log
	s.finalize()
}

func (s *trialProxySession) proxyClientToDeepgram() {
	for {
		messageType, data, err := s.clientConn.ReadMessage()
		if err != nil {
			log.Printf("[Trial Deepgram] Client read error: %v", err)
			_ = s.deepgramConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"CloseStream"}`))
			return
		}

		// Track bytes sent (only for binary audio data)
		if messageType == websocket.BinaryMessage {
			s.mu.Lock()
			s.bytesSent += int64(len(data))
			s.mu.Unlock()
		}

		// Forward to Deepgram
		if err := s.deepgramConn.WriteMessage(messageType, data); err != nil {
			log.Printf("[Trial Deepgram] Error forwarding to Deepgram: %v", err)
			return
		}
	}
}

func (s *trialProxySession) proxyDeepgramToClient() {
	clientClosed := false

	for {
		messageType, data, err := s.deepgramConn.ReadMessage()
		if err != nil {
			log.Printf("[Trial Deepgram] Deepgram read error: %v", err)
			return
		}

		// Parse Deepgram response to extract duration
		if messageType == websocket.TextMessage {
			s.extractDurationFromResponse(data)

			var msg struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(data, &msg) == nil && msg.Type == "Metadata" {
				if !clientClosed {
					if err := s.clientConn.WriteMessage(messageType, data); err != nil {
						clientClosed = true
					}
				}
				continue
			}
		}

		// Forward to client
		if !clientClosed {
			if err := s.clientConn.WriteMessage(messageType, data); err != nil {
				log.Printf("[Trial Deepgram] Error forwarding to client: %v", err)
				clientClosed = true
			}
		}
	}
}

func (s *trialProxySession) extractDurationFromResponse(data []byte) {
	var response struct {
		Type     string  `json:"type"`
		Duration float64 `json:"duration"`
		Metadata *struct {
			Duration float64 `json:"duration"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &response); err == nil {
		if response.Type == "Metadata" && response.Duration > 0 {
			s.mu.Lock()
			s.duration = response.Duration
			s.mu.Unlock()
		}
		if response.Metadata != nil && response.Metadata.Duration > 0 {
			s.mu.Lock()
			s.duration = response.Metadata.Duration
			s.mu.Unlock()
		}
	}
}

func (s *trialProxySession) closeWithTimeout() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	_ = s.clientConn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Trial session time limit reached"))
	s.clientConn.Close()
	s.deepgramConn.Close()
}

func (s *trialProxySession) finalize() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	log.Printf("[Trial Deepgram] Finalizing session - duration: %.3f, bytes: %d", s.duration, s.bytesSent)

	ctx := context.Background()

	if s.duration > 0 {
		durationStr := fmt.Sprintf("%.3f", s.duration)
		_ = s.queries.UpdateTrialUsageComplete(ctx, sqlc.UpdateTrialUsageCompleteParams{
			ID:              s.logID,
			DurationSeconds: stringToNumeric(durationStr),
			BytesSent:       s.bytesSent,
		})
	} else {
		// No duration captured - treat as timeout
		_ = s.queries.UpdateTrialUsageTimeout(ctx, sqlc.UpdateTrialUsageTimeoutParams{
			ID:        s.logID,
			BytesSent: s.bytesSent,
		})
	}
}

// ========== HELPER FUNCTIONS ==========

func hashTrialAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func getUpgradeURL() string {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://hyperwhisper.dev"
	}
	return baseURL + "/signup"
}

// IsTrialKey checks if an API key is a trial key (hw_trial_ prefix)
func IsTrialKey(apiKey string) bool {
	return len(apiKey) >= 9 && apiKey[:9] == "hw_trial_"
}

// getTrialExpiryDays returns the trial expiry days from env or default
func getTrialExpiryDays() int {
	expiryStr := os.Getenv("TRIAL_EXPIRY_DAYS")
	if expiryStr == "" {
		return 90
	}
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		return 90
	}
	return expiry
}
