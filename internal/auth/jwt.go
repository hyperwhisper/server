package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// Claims represents the JWT claims
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	UserType  string    `json:"user_type"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiry in seconds
}

// getJWTSecret returns the JWT secret from environment
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "hyperwhisper-dev-secret-change-in-production"
	}
	return []byte(secret)
}

// getAccessTokenExpiry returns access token expiry duration
func getAccessTokenExpiry() time.Duration {
	expiryStr := os.Getenv("ACCESS_TOKEN_EXPIRY")
	if expiryStr == "" {
		return 5 * time.Minute // Default 5 minutes
	}
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		return 5 * time.Minute
	}
	return time.Duration(expiry) * time.Minute
}

// getRefreshTokenExpiry returns refresh token expiry duration
func getRefreshTokenExpiry() time.Duration {
	expiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if expiryStr == "" {
		return 7 * 24 * time.Hour // Default 7 days
	}
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		return 7 * 24 * time.Hour
	}
	return time.Duration(expiry) * 24 * time.Hour
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uuid.UUID, username, email, userType string) (*TokenPair, error) {
	accessExpiry := getAccessTokenExpiry()
	refreshExpiry := getRefreshTokenExpiry()
	secret := getJWTSecret()
	now := time.Now()

	// Generate unique JTI for each token
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	// Generate access token
	accessClaims := &Claims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		UserType:  userType,
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		UserType:  userType,
		TokenType: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(accessExpiry.Seconds()),
	}, nil
}

// ValidateToken validates a token and returns the claims
func ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}
