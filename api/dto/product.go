package dto

import (
	productmodel "coffeebase-api/internal/models/product"
)

// ProductRequest defines the expected payload for creating or updating a product
type ProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// ProductResponse is the object returned to the client
type ProductResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// MapProductToResponse securely converts an internal Product model into an API DTO
func MapProductToResponse(p productmodel.Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Category:    p.Category,
	}
}

// MapProductsToResponse converts a slice of Product models into a slice of DTOs
func MapProductsToResponse(products []productmodel.Product) []ProductResponse {
	var dtos []ProductResponse
	for _, p := range products {
		dtos = append(dtos, MapProductToResponse(p))
	}
	return dtos
}
