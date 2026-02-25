package cart

import "time"

type Item struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Cart struct {
	UserID    int       `json:"user_id"`
	Items     []Item    `json:"items"`
	UpdatedAt time.Time `json:"updated_at"`
}
