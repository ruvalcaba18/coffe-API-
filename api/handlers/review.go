package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/middleware"
	reviewmodel "coffeebase-api/internal/models/review"
	"coffeebase-api/internal/store/review"
	"net/http"
	"strconv"
)

type ReviewHandler struct {
	reviewStore review.Store
}

// --- Public ---

func NewReviewHandler(reviewStore review.Store) *ReviewHandler {
	return &ReviewHandler{
		reviewStore: reviewStore,
	}
}

func (reviewHandler *ReviewHandler) Create(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var request dto.ReviewRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	reviewInstance := &reviewmodel.Review{
		ProductID: request.ProductID,
		UserID:    userID,
		Rating:    request.Rating,
		Comment:   request.Comment,
	}

	if error := reviewHandler.reviewStore.Create(httpRequest.Context(), reviewInstance); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusCreated, dto.MapReviewToResponse(*reviewInstance))
}

func (reviewHandler *ReviewHandler) GetByProduct(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	productIDString := httpRequest.URL.Query().Get("product_id")
	productID, conversionError := strconv.Atoi(productIDString)
	if conversionError != nil || productID == 0 {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	reviewList, error := reviewHandler.reviewStore.GetByProductID(httpRequest.Context(), productID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapReviewsToResponse(reviewList))
}
