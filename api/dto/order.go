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
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
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

func MapOrderToResponse(orderInstance ordermodel.Order) OrderResponse {
	items := make([]OrderItemDTO, len(orderInstance.Items))
	for index, item := range orderInstance.Items {
		items[index] = OrderItemDTO{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		}
	}

	return OrderResponse{
		ID:             orderInstance.ID,
		UserID:         orderInstance.UserID,
		Items:          items,
		Total:          orderInstance.Total,
		CouponCode:     orderInstance.CouponCode,
		DiscountAmount: orderInstance.DiscountAmount,
		Status:         orderInstance.Status,
		IsPickup:       orderInstance.IsPickup,
		PickupTime:     orderInstance.PickupTime,
		PickupLocation: orderInstance.PickupLocation,
		ItemsCount:     orderInstance.ItemsCount,
		CreatedAt:      orderInstance.CreatedAt,
	}
}

func MapOrdersToResponse(orders []ordermodel.Order) []OrderResponse {
	dtos := make([]OrderResponse, len(orders))
	for index, orderInstance := range orders {
		dtos[index] = MapOrderToResponse(orderInstance)
	}
	return dtos
}
