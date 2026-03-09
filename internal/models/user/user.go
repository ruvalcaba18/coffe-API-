package user

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Language  string    `json:"language"` // es, en, fr, de, gsw
	AvatarURL string    `json:"avatar_url"`
	Role                 string    `json:"role"`
	TotalOrdersCompleted  int       `json:"total_orders_completed"`
	TotalSpent           float64   `json:"total_spent"`
	CreatedAt            time.Time `json:"created_at"`
}
