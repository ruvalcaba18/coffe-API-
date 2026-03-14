package dto

import (
	usermodel "coffeebase-api/internal/models/user"
	"time"
)

type UserResponse struct {
	ID                   int       `json:"id"`
	Username             string    `json:"username"`
	Email                string    `json:"email"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	Birthday             time.Time `json:"birthday"`
	Language             string    `json:"language"`
	AvatarURL            string    `json:"avatar_url"`
	Role                 string    `json:"role"`
	TotalOrdersCompleted int       `json:"total_orders_completed"`
	TotalSpent           float64   `json:"total_spent"`
	CreatedAt            time.Time `json:"created_at"`
}

func MapUserToResponse(u usermodel.User) UserResponse {
	return UserResponse{
		ID:                   u.ID,
		Username:             u.Username,
		Email:                u.Email,
		FirstName:            u.FirstName,
		LastName:             u.LastName,
		Birthday:             u.Birthday,
		Language:             u.Language,
		AvatarURL:            u.AvatarURL,
		Role:                 u.Role,
		TotalOrdersCompleted: u.TotalOrdersCompleted,
		TotalSpent:           u.TotalSpent,
		CreatedAt:            u.CreatedAt,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Language string `json:"language,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateLanguageRequest struct {
	Language string `json:"language"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Birthday  string `json:"birthday,omitempty"`
	Language  string `json:"language,omitempty"`
}

// MapUsersToResponse converts a slice of User models into a slice of DTOs
func MapUsersToResponse(users []usermodel.User) []UserResponse {
	dtos := make([]UserResponse, 0)
	for _, u := range users {
		dtos = append(dtos, MapUserToResponse(u))
	}
	return dtos
}
