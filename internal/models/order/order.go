package order

import "time"

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Order struct {
	ID         string      `json:"id"`
	UserID     int         `json:"user_id"`
	Items      []OrderItem `json:"items"`
	Total          float64     `json:"total"`
	CouponCode     string      `json:"coupon_code,omitempty"`
	DiscountAmount float64     `json:"discount_amount,omitempty"`
	Status         string      `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
}
