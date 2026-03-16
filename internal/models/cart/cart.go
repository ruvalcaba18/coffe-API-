package cart

import "time"

type Item struct {
	ProductID   int     `json:"product_id"`
	Quantity    int     `json:"quantity"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

type Cart struct {
	UserID    int       `json:"user_id"`
	Items     []Item    `json:"items"`
	UpdatedAt time.Time `json:"updated_at"`
}
