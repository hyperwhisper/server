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
