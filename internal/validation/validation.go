package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

var strictPolicy = bluemonday.StrictPolicy() // strips ALL HTML/JS

// --- Errors ---
var (
	ErrFieldTooLong     = errors.New("field exceeds maximum allowed length")
	ErrFieldEmpty       = errors.New("required field is empty")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidUsername  = errors.New("username contains invalid characters")
	ErrPasswordTooWeak  = errors.New("password must be at least 8 characters")
	ErrInvalidCouponCode = errors.New("coupon code contains invalid characters")
)

var emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{2,50}$`)
var couponRegex   = regexp.MustCompile(`^[A-Z0-9_\-]{1,50}$`)

// --- Public ---

func Sanitize(input string) string {
	return strings.TrimSpace(strictPolicy.Sanitize(input))
}

func SafeText(input string, maxLen int) (string, error) {
	clean := Sanitize(input)
	if utf8.RuneCountInString(clean) > maxLen {
		return "", ErrFieldTooLong
	}
	return clean, nil
}

func Email(email string) (string, error) {
	clean := strings.ToLower(strings.TrimSpace(email))
	if clean == "" {
		return "", ErrFieldEmpty
	}
	if len(clean) > 254 {
		return "", ErrFieldTooLong
	}
	if !emailRegex.MatchString(clean) {
		return "", ErrInvalidEmail
	}
	return clean, nil
}

func Username(username string) (string, error) {
	clean := strings.TrimSpace(username)
	if clean == "" {
		return "", ErrFieldEmpty
	}
	if !usernameRegex.MatchString(clean) {
		return "", ErrInvalidUsername
	}
	return clean, nil
}

func Password(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return ErrPasswordTooWeak
	}
	if len(password) > 128 {
		return ErrFieldTooLong
	}
	return nil
}

func CouponCode(code string) (string, error) {
	clean := strings.ToUpper(strings.TrimSpace(code))
	if clean == "" {
		return "", ErrFieldEmpty
	}
	if !couponRegex.MatchString(clean) {
		return "", ErrInvalidCouponCode
	}
	return clean, nil
}
