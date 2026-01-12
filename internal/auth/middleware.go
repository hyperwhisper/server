package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const UserContextKey = "user"

// JWTMiddleware creates an Echo middleware for JWT authentication
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var tokenString string

			// Try to get token from Authorization header first
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// Fall back to cookie if no header
			if tokenString == "" {
				cookie, err := c.Cookie("access_token")
				if err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing authentication token",
				})
			}

			// Validate the token
			claims, err := ValidateToken(tokenString, AccessToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": err.Error(),
				})
			}

			// Store claims in context
			c.Set(UserContextKey, claims)

			return next(c)
		}
	}
}

// GetUserFromContext retrieves user claims from Echo context
func GetUserFromContext(c echo.Context) *Claims {
	claims, ok := c.Get(UserContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
