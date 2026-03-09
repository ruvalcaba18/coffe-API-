package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/middleware"
	reviewmodel "coffeebase-api/internal/models/review"
	reviewstore "coffeebase-api/internal/store/review"
	"encoding/json"
	"net/http"
	"strconv"
)

type ReviewHandler struct {
	ReviewStore *reviewstore.Store
}

func (reviewHandler *ReviewHandler) Create(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)

	var reviewInput dto.ReviewRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&reviewInput); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	productReview := &reviewmodel.Review{
		ProductID: reviewInput.ProductID,
		UserID:    userID,
		Rating:    reviewInput.Rating,
		Comment:   reviewInput.Comment,
	}

	if creationError := reviewHandler.ReviewStore.Create(productReview); creationError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)
	json.NewEncoder(responseWriter).Encode(dto.MapReviewToResponse(*productReview))
}

func (reviewHandler *ReviewHandler) GetByProduct(responseWriter http.ResponseWriter, request *http.Request) {
	productIdentifier, _ := strconv.Atoi(request.URL.Query().Get("product_id"))
	if productIdentifier == 0 {
		http.Error(responseWriter, "product_id query param is required", http.StatusBadRequest)
		return
	}

	reviewList, fetchError := reviewHandler.ReviewStore.GetByProductID(productIdentifier)
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapReviewsToResponse(reviewList))
}
