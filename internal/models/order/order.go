package order

import "time"

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Order struct {
	ID             string      `json:"id"`
	UserID         int         `json:"user_id"`
	Items          []OrderItem `json:"items"`
	Total          float64     `json:"total"`
	CouponCode     string      `json:"coupon_code,omitempty"`
	DiscountAmount float64     `json:"discount_amount,omitempty"`
	Status         string      `json:"status"`
	IsPickup       bool        `json:"is_pickup"`
	PickupTime     *time.Time  `json:"pickup_time,omitempty"`
	PickupLocation string      `json:"pickup_location,omitempty"`
	ItemsCount     int         `json:"items_count"`
	CreatedAt      time.Time   `json:"created_at"`
}
