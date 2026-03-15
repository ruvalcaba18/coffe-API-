package handlers

import (
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/middleware"
	billingmodel "coffeebase-api/internal/models/billing"
	"coffeebase-api/internal/store/billing"
	"net/http"
)

type BillingHandler struct {
	billingStore billing.Store
}

// --- Public ---

func NewBillingHandler(billingStore billing.Store) *BillingHandler {
	return &BillingHandler{
		billingStore: billingStore,
	}
}

func (billingHandler *BillingHandler) GetWallet(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	response.SendJSON(responseWriter, http.StatusOK, map[string]interface{}{
		"balance": 150.50, 
	})
}

func (billingHandler *BillingHandler) GetPaymentMethods(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)
	
	paymentMethods, error := billingHandler.billingStore.GetPaymentMethodsByUserID(httpRequest.Context(), userID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, paymentMethods)
}

func (billingHandler *BillingHandler) AddPaymentMethod(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var request struct {
		Last4  string `json:"last4"`
		Expiry string `json:"expiry"`
		Brand  string `json:"brand"`
		Holder string `json:"holder"`
	}

	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	paymentMethodInstance := &billingmodel.PaymentMethod{
		UserID: userID,
		Last4:  request.Last4,
		Expiry: request.Expiry,
		Brand:  request.Brand,
		Holder: request.Holder,
	}

	if error := billingHandler.billingStore.AddPaymentMethod(httpRequest.Context(), paymentMethodInstance); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusCreated, paymentMethodInstance)
}
