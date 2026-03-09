package dto

import (
	favoritemodel "coffeebase-api/internal/models/favorite"
	"time"
)

// FavoriteRequest defines the payload for adding/removing a favorite
type FavoriteRequest struct {
	ProductID int `json:"product_id" validate:"required"`
}

// FavoriteResponse represents a favorite record returned to the client
type FavoriteResponse struct {
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

// MapFavoriteToResponse converts an internal Favorite model into an API DTO
func MapFavoriteToResponse(f favoritemodel.Favorite) FavoriteResponse {
	return FavoriteResponse{
		UserID:    f.UserID,
		ProductID: f.ProductID,
		CreatedAt: f.CreatedAt,
	}
}

// MapFavoritesToResponse converts a slice of internal Favorite models into API DTOs
func MapFavoritesToResponse(favorites []favoritemodel.Favorite) []FavoriteResponse {
	dtos := make([]FavoriteResponse, len(favorites))
	for i, f := range favorites {
		dtos[i] = MapFavoriteToResponse(f)
	}
	return dtos
}
