package dto

import (
	usermodel "coffeebase-api/internal/models/user"
	"time"
)

// UserResponse is the safe object that will be returned to the client
// (It omits sensitive data like Password)
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Language  string    `json:"language"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// MapUserToResponse securely converts an internal User model into an API DTO 
func MapUserToResponse(u usermodel.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Language:  u.Language,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// RegisterRequest defines the expected payload for registering a new user
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Language string `json:"language,omitempty"`
}

// LoginRequest defines the payload for generating JWT tokens
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateLanguageRequest is the DTO for users changing their preferred language
type UpdateLanguageRequest struct {
	Language string `json:"language"`
}
