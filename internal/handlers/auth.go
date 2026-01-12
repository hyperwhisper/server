package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"hyperwhisper/internal/auth"
	"hyperwhisper/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Request types
type SignUpRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SignInRequest struct {
	Identifier string `json:"identifier"` // email or username
	Password   string `json:"password"`
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Response types
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type"`
	CreatedAt string `json:"created_at"`
}

type AuthResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"access_token"`
	ExpiresIn   int64        `json:"expires_in"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	queries *sqlc.Queries
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		queries: sqlc.New(db),
	}
}

// SignUp handles user registration
func (h *AuthHandler) SignUp(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "username, email, and password are required"})
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

	// Check if this is the first user (make them admin)
	userCount, err := h.queries.CountUsers(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	userType := "user"
	if userCount == 0 {
		userType = "admin"
	}

	// Create user
	user, err := h.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		UserType:     userType,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
	}

	// Generate tokens
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username, user.Email, user.UserType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate tokens"})
	}

	// Store tokens in database
	if err := h.storeRefreshToken(ctx, user.ID, tokens); err != nil {
		// Log error but don't fail - tokens are still valid
		// In production, you might want to handle this differently
	}

	// Set cookies
	setAuthCookies(c, tokens)

	return c.JSON(http.StatusCreated, AuthResponse{
		User:        toUserResponse(user),
		AccessToken: tokens.AccessToken,
		ExpiresIn:   tokens.ExpiresIn,
	})
}

// SignIn handles user login
func (h *AuthHandler) SignIn(c echo.Context) error {
	var req SignInRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Identifier == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "identifier and password are required"})
	}

	ctx := context.Background()

	// Find user by email or username
	user, err := h.queries.GetUserByEmailOrUsername(ctx, req.Identifier)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database error"})
	}

	// Verify password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
	}

	// Generate tokens
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username, user.Email, user.UserType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate tokens"})
	}

	// Store tokens in database
	if err := h.storeRefreshToken(ctx, user.ID, tokens); err != nil {
		// Log error but don't fail - tokens are still valid
	}

	// Set cookies
	setAuthCookies(c, tokens)

	return c.JSON(http.StatusOK, AuthResponse{
		User:        toUserResponse(user),
		AccessToken: tokens.AccessToken,
		ExpiresIn:   tokens.ExpiresIn,
	})
}

// TokenRefresh handles refreshing access tokens
func (h *AuthHandler) TokenRefresh(c echo.Context) error {
	var refreshToken string

	// Debug: log all cookies received
	cookies := c.Cookies()
	fmt.Printf("[TokenRefresh] Received %d cookies\n", len(cookies))
	for _, cookie := range cookies {
		fmt.Printf("[TokenRefresh] Cookie: %s (path: %s)\n", cookie.Name, cookie.Path)
	}

	// Get refresh token from cookie first
	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
		fmt.Printf("[TokenRefresh] Found refresh_token cookie\n")
	} else {
		fmt.Printf("[TokenRefresh] No refresh_token cookie: %v\n", err)
	}

	// Fall back to request body
	if refreshToken == "" {
		var req TokenRefreshRequest
		if err := c.Bind(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "refresh token required"})
	}

	// Validate refresh token
	claims, err := auth.ValidateToken(refreshToken, auth.RefreshToken)
	if err != nil {
		clearAuthCookies(c)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	}

	ctx := context.Background()

	// Check if refresh token is revoked
	isRevoked, err := h.queries.IsRefreshTokenRevoked(ctx, claims.ID)
	if err == nil && isRevoked {
		clearAuthCookies(c)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "token has been revoked"})
	}

	// Generate new token pair
	tokens, err := auth.GenerateTokenPair(claims.UserID, claims.Username, claims.Email, claims.UserType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate tokens"})
	}

	// Revoke the old refresh token (single-use)
	_ = h.queries.RevokeRefreshToken(ctx, sqlc.RevokeRefreshTokenParams{
		TokenJti:      claims.ID,
		RevokedReason: sql.NullString{String: "refreshed", Valid: true},
	})

	// Store new tokens in database
	if err := h.storeRefreshToken(ctx, claims.UserID, tokens); err != nil {
		// Log error but don't fail
	}

	// Set new cookies
	setAuthCookies(c, tokens)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": tokens.AccessToken,
		"expires_in":   tokens.ExpiresIn,
	})
}

// SignOut handles logout
func (h *AuthHandler) SignOut(c echo.Context) error {
	clearAuthCookies(c)
	return c.JSON(http.StatusOK, map[string]string{"message": "signed out successfully"})
}

// Me returns current user info
func (h *AuthHandler) Me(c echo.Context) error {
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
	}

	ctx := context.Background()
	user, err := h.queries.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
}

// Helper functions
func toUserResponse(user sqlc.User) UserResponse {
	createdAt := ""
	if user.CreatedAt.Valid {
		createdAt = user.CreatedAt.Time.Format(time.RFC3339)
	}

	return UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserType:  user.UserType,
		CreatedAt: createdAt,
	}
}

func isSecureMode() bool {
	return os.Getenv("APP_ENV") != "dev"
}

func getRefreshTokenExpiryDays() int {
	expiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if expiryStr == "" {
		return 7
	}
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		return 7
	}
	return expiry
}

func getAccessTokenExpiryMinutes() int {
	expiryStr := os.Getenv("ACCESS_TOKEN_EXPIRY")
	if expiryStr == "" {
		return 5
	}
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		return 5
	}
	return expiry
}

func setAuthCookies(c echo.Context, tokens *auth.TokenPair) {
	secure := isSecureMode()
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteStrictMode
	}

	// Access token cookie (backup, primary is in response body)
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   getAccessTokenExpiryMinutes() * 60,
	})

	// Refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/api/v1",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   getRefreshTokenExpiryDays() * 24 * 60 * 60,
	})
}

func clearAuthCookies(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// storeRefreshToken saves the refresh token to the database for tracking
func (h *AuthHandler) storeRefreshToken(ctx context.Context, userID uuid.UUID, tokens *auth.TokenPair) error {
	// Parse refresh token to get JTI and expiry
	refreshClaims, err := auth.ValidateToken(tokens.RefreshToken, auth.RefreshToken)
	if err != nil {
		return err
	}

	// Store refresh token
	_, err = h.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		TokenJti:  refreshClaims.ID,
		UserID:    userID,
		ExpiresAt: refreshClaims.ExpiresAt.Time,
	})
	if err != nil {
		return err
	}

	return nil
}
