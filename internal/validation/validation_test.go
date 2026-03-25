package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmail_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple email", input: "user@example.com", expected: "user@example.com"},
		{name: "uppercase email", input: "USER@EXAMPLE.COM", expected: "user@example.com"},
		{name: "with spaces", input: "  user@example.com  ", expected: "user@example.com"},
		{name: "plus addressing", input: "user+tag@example.com", expected: "user+tag@example.com"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, validationError := Email(testCase.input)
			assert.NoError(t, validationError)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestEmail_Invalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces only", input: "   "},
		{name: "no domain", input: "user@"},
		{name: "no at sign", input: "userexample.com"},
		{name: "no tld", input: "user@example"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, validationError := Email(testCase.input)
			assert.Error(t, validationError)
		})
	}
}

func TestUsername_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple", input: "john_doe", expected: "john_doe"},
		{name: "with dots", input: "john.doe", expected: "john.doe"},
		{name: "with dashes", input: "john-doe", expected: "john-doe"},
		{name: "alphanumeric", input: "user123", expected: "user123"},
		{name: "minimum length", input: "ab", expected: "ab"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, validationError := Username(testCase.input)
			assert.NoError(t, validationError)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestUsername_Invalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "too short", input: "a"},
		{name: "special chars", input: "user@name"},
		{name: "spaces", input: "user name"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, validationError := Username(testCase.input)
			assert.Error(t, validationError)
		})
	}
}

func TestPassword_Valid(t *testing.T) {
	assert.NoError(t, Password("12345678"))
	assert.NoError(t, Password("StrongPass!"))
}

func TestPassword_TooShort(t *testing.T) {
	assert.ErrorIs(t, Password("short"), ErrPasswordTooWeak)
	assert.ErrorIs(t, Password("1234567"), ErrPasswordTooWeak)
}

func TestPassword_TooLong(t *testing.T) {
	longPassword := make([]byte, 129)
	for index := range longPassword {
		longPassword[index] = 'a'
	}
	assert.ErrorIs(t, Password(string(longPassword)), ErrFieldTooLong)
}

func TestCouponCode_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "uppercase", input: "SAVE20", expected: "SAVE20"},
		{name: "lowercase converts", input: "save20", expected: "SAVE20"},
		{name: "with dashes", input: "WELCOME-2024", expected: "WELCOME-2024"},
		{name: "with underscore", input: "VIP_MEMBER", expected: "VIP_MEMBER"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, validationError := CouponCode(testCase.input)
			assert.NoError(t, validationError)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestCouponCode_Invalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces", input: "   "},
		{name: "special chars", input: "SAVE@20"},
		{name: "lowercase special", input: "save!20"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, validationError := CouponCode(testCase.input)
			assert.Error(t, validationError)
		})
	}
}

func TestSanitize_RemovesHTML(t *testing.T) {
	result := Sanitize("<script>alert('xss')</script>Hello")
	assert.Equal(t, "Hello", result)
}

func TestSanitize_TrimsSpaces(t *testing.T) {
	result := Sanitize("  hello world  ")
	assert.Equal(t, "hello world", result)
}

func TestSafeText_Valid(t *testing.T) {
	result, validationError := SafeText("Hello World", 50)
	assert.NoError(t, validationError)
	assert.Equal(t, "Hello World", result)
}

func TestSafeText_TooLong(t *testing.T) {
	_, validationError := SafeText("This is a very long text", 5)
	assert.ErrorIs(t, validationError, ErrFieldTooLong)
}
