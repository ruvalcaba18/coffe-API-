package dto

import (
	ordermodel "coffeebase-api/internal/models/order"
	"time"
)

type CheckoutRequest struct {
	CouponCode     string     `json:"coupon_code"`
	IsPickup       bool       `json:"is_pickup"`
	PickupTime     *time.Time `json:"pickup_time"`
	PickupLocation string     `json:"pickup_location"`
}

type OrderItemDTO struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type OrderResponse struct {
	ID             string         `json:"id"`
	UserID         int            `json:"user_id"`
	Items          []OrderItemDTO `json:"items"`
	Total          float64        `json:"total"`
	CouponCode     string         `json:"coupon_code,omitempty"`
	DiscountAmount float64        `json:"discount_amount,omitempty"`
	Status         string         `json:"status"`
	IsPickup       bool           `json:"is_pickup"`
	PickupTime     *time.Time     `json:"pickup_time,omitempty"`
	PickupLocation string         `json:"pickup_location,omitempty"`
	ItemsCount     int            `json:"items_count"`
	CreatedAt      time.Time      `json:"created_at"`
}

func MapOrderToResponse(o ordermodel.Order) OrderResponse {
	items := make([]OrderItemDTO, len(o.Items))
	for i, item := range o.Items {
		items[i] = OrderItemDTO{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return OrderResponse{
		ID:             o.ID,
		UserID:         o.UserID,
		Items:          items,
		Total:          o.Total,
		CouponCode:     o.CouponCode,
		DiscountAmount: o.DiscountAmount,
		Status:         o.Status,
		IsPickup:       o.IsPickup,
		PickupTime:     o.PickupTime,
		PickupLocation: o.PickupLocation,
		ItemsCount:     o.ItemsCount,
		CreatedAt:      o.CreatedAt,
	}
}

func MapOrdersToResponse(orders []ordermodel.Order) []OrderResponse {
	dtos := make([]OrderResponse, len(orders))
	for i, o := range orders {
		dtos[i] = MapOrderToResponse(o)
	}
	return dtos
}
