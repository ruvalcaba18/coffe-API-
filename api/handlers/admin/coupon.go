package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/models/coupon"
	couponstore "coffeebase-api/internal/store/coupon"
	"coffeebase-api/internal/validation"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CouponHandler struct {
	couponStore couponstore.Store
}

// --- Public ---

func NewCouponHandler(couponStore couponstore.Store) *CouponHandler {
	return &CouponHandler{
		couponStore: couponStore,
	}
}

func (couponHandler *CouponHandler) Create(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var request dto.CouponRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	// OWASP A03 - Validate coupon code format
	cleanCode, error := validation.CouponCode(request.Code)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidRequest)
		return
	}

	couponInstance := &coupon.Coupon{
		Code:              cleanCode,
		DiscountType:      request.DiscountType,
		DiscountValue:     request.DiscountValue,
		MinPurchaseAmount: request.MinPurchaseAmount,
		MaxDiscountAmount: request.MaxDiscountAmount,
		StartDate:         request.StartDate,
		EndDate:           request.EndDate,
		UsageLimit:        request.UsageLimit,
		IsActive:          request.IsActive,
	}

	if error := couponHandler.couponStore.Create(httpRequest.Context(), couponInstance); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusCreated, dto.MapCouponToResponse(*couponInstance))
}

func (couponHandler *CouponHandler) GetAll(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	couponList, error := couponHandler.couponStore.GetAll(httpRequest.Context())
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapCouponsToResponse(couponList))
}

func (couponHandler *CouponHandler) ToggleStatus(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	couponIDString := chi.URLParam(httpRequest, "id")
	couponID, conversionError := strconv.Atoi(couponIDString)
	if conversionError != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	var request struct {
		IsActive bool `json:"is_active"`
	}
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if error := couponHandler.couponStore.ToggleStatus(httpRequest.Context(), couponID, request.IsActive); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func (couponHandler *CouponHandler) Delete(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	couponIDString := chi.URLParam(httpRequest, "id")
	couponID, conversionError := strconv.Atoi(couponIDString)
	if conversionError != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	if error := couponHandler.couponStore.Delete(httpRequest.Context(), couponID); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusNoContent)
}
