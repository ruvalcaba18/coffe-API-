package dto

import (
	productmodel "coffeebase-api/internal/models/product"
)

type ProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

type ProductResponse struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	Category      string  `json:"category"`
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int     `json:"review_count"`
}

func MapProductToResponse(p productmodel.Product) ProductResponse {
	return ProductResponse{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		Category:      p.Category,
		AverageRating: p.AverageRating,
		ReviewCount:   p.ReviewCount,
	}
}

func MapProductsToResponse(products []productmodel.Product) []ProductResponse {
	var dtos []ProductResponse
	for _, p := range products {
		dtos = append(dtos, MapProductToResponse(p))
	}
	return dtos
}
