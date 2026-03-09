package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/models/coupon"
	couponstore "coffeebase-api/internal/store/coupon"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CouponHandler struct {
	CouponStore *couponstore.Store
}

func (couponHandler *CouponHandler) Create(responseWriter http.ResponseWriter, request *http.Request) {
	var couponInput dto.CouponRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&couponInput); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	couponInstance := &coupon.Coupon{
		Code:              couponInput.Code,
		DiscountType:      couponInput.DiscountType,
		DiscountValue:     couponInput.DiscountValue,
		MinPurchaseAmount: couponInput.MinPurchaseAmount,
		MaxDiscountAmount: couponInput.MaxDiscountAmount,
		StartDate:         couponInput.StartDate,
		EndDate:           couponInput.EndDate,
		UsageLimit:        couponInput.UsageLimit,
		IsActive:          couponInput.IsActive,
	}

	if createError := couponHandler.CouponStore.Create(couponInstance); createError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)
	json.NewEncoder(responseWriter).Encode(dto.MapCouponToResponse(*couponInstance))
}

func (couponHandler *CouponHandler) GetAll(responseWriter http.ResponseWriter, request *http.Request) {
	couponList, fetchError := couponHandler.CouponStore.GetAll()
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapCouponsToResponse(couponList))
}

func (couponHandler *CouponHandler) ToggleStatus(responseWriter http.ResponseWriter, request *http.Request) {
	couponIdentifier, _ := strconv.Atoi(chi.URLParam(request, "id"))
	var toggleRequest struct {
		IsActive bool `json:"is_active"`
	}
	if decodeError := json.NewDecoder(request.Body).Decode(&toggleRequest); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	if toggleError := couponHandler.CouponStore.ToggleStatus(couponIdentifier, toggleRequest.IsActive); toggleError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func (couponHandler *CouponHandler) Delete(responseWriter http.ResponseWriter, request *http.Request) {
	couponIdentifier, _ := strconv.Atoi(chi.URLParam(request, "id"))
	if deletionError := couponHandler.CouponStore.Delete(couponIdentifier); deletionError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusNoContent)
}
