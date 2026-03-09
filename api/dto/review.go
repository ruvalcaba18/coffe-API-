package dto

import (
	reviewmodel "coffeebase-api/internal/models/review"
	"time"
)

// ReviewRequest defines the payload for submitting a review
type ReviewRequest struct {
	ProductID int    `json:"product_id" validate:"required"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
	Comment   string `json:"comment"`
}

// ReviewResponse represents a review returned to the client
type ReviewResponse struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// MapReviewToResponse converts an internal Review model into an API DTO
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

// MapReviewsToResponse converts a slice of internal Review models into API DTOs
func MapReviewsToResponse(reviews []reviewmodel.Review) []ReviewResponse {
	dtos := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		dtos[i] = MapReviewToResponse(r)
	}
	return dtos
}
