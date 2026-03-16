package dto

import (
	cartmodel "coffeebase-api/internal/models/cart"
	"time"
)

// CartUpdateRequest defines the payload for updating a cart item
type CartUpdateRequest struct {
	ProductID int `json:"product_id" validate:"required"`
	Quantity  int `json:"quantity" validate:"min=0"`
}

// CartItemDTO represents a cart item in the API
type CartItemDTO struct {
	ProductID   int     `json:"product_id"`
	Quantity    int     `json:"quantity"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

// CartResponse represents the cart data returned to the client
type CartResponse struct {
	UserID    int           `json:"user_id"`
	Items     []CartItemDTO `json:"items"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// MapCartToResponse converts an internal Cart model into an API DTO
func MapCartToResponse(c cartmodel.Cart) CartResponse {
	items := make([]CartItemDTO, len(c.Items))
	for i, item := range c.Items {
		items[i] = CartItemDTO{
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
			ProductName: item.ProductName,
			Price:       item.Price,
			ImageURL:    item.ImageURL,
		}
	}

	return CartResponse{
		UserID:    c.UserID,
		Items:     items,
		UpdatedAt: c.UpdatedAt,
	}
}
