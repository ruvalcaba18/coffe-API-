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

func MapProductToResponse(productInstance productmodel.Product) ProductResponse {
	return ProductResponse{
		ID:            productInstance.ID,
		Name:          productInstance.Name,
		Description:   productInstance.Description,
		Price:         productInstance.Price,
		Category:      productInstance.Category,
		AverageRating: productInstance.AverageRating,
		ReviewCount:   productInstance.ReviewCount,
	}
}

func MapProductsToResponse(products []productmodel.Product) []ProductResponse {
	var dtos []ProductResponse
	for _, productInstance := range products {
		dtos = append(dtos, MapProductToResponse(productInstance))
	}
	return dtos
}
