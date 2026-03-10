package dto

import (
	favoritemodel "coffeebase-api/internal/models/favorite"
	"time"
)

type FavoriteRequest struct {
	ProductID int `json:"product_id" validate:"required"`
}

type FavoriteResponse struct {
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

func MapFavoriteToResponse(f favoritemodel.Favorite) FavoriteResponse {
	return FavoriteResponse{
		UserID:    f.UserID,
		ProductID: f.ProductID,
		CreatedAt: f.CreatedAt,
	}
}

func MapFavoritesToResponse(favorites []favoritemodel.Favorite) []FavoriteResponse {
	dtos := make([]FavoriteResponse, len(favorites))
	for i, f := range favorites {
		dtos[i] = MapFavoriteToResponse(f)
	}
	return dtos
}
