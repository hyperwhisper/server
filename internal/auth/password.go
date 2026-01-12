package auth

import (
	"errors"
	"os"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters")
	ErrPasswordNoUppercase = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoNumber    = errors.New("password must contain at least one number")
	ErrPasswordNoSpecial   = errors.New("password must contain at least one special character")
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePassword validates password strength
// In dev mode (APP_ENV=dev), no validation is performed
// In prod mode, strong validation is enforced
func ValidatePassword(password string) error {
	// Skip validation in dev mode
	if os.Getenv("APP_ENV") == "dev" {
		return nil
	}

	// Strong validation for production
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return ErrPasswordNoUppercase
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return ErrPasswordNoLowercase
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return ErrPasswordNoNumber
	}

	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return ErrPasswordNoSpecial
	}

	return nil
}
