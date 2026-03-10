package dto

import (
	reviewmodel "coffeebase-api/internal/models/review"
	"time"
)

type ReviewRequest struct {
	ProductID int    `json:"product_id" validate:"required"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
	Comment   string `json:"comment"`
}

type ReviewResponse struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

func MapReviewToResponse(r reviewmodel.Review) ReviewResponse {
	return ReviewResponse{
		ID:        r.ID,
		ProductID: r.ProductID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt,
	}
}

func MapReviewsToResponse(reviews []reviewmodel.Review) []ReviewResponse {
	dtos := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		dtos[i] = MapReviewToResponse(r)
	}
	return dtos
}
