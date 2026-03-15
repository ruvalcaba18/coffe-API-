package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/middleware"
	ordermodel "coffeebase-api/internal/models/order"
	"coffeebase-api/internal/store/order"
	"coffeebase-api/internal/store/product"
	"context"
	"net/http"
	"time"
)

type OrderCheckoutService interface {
	Checkout(requestContext context.Context, userID int, couponCode string, isPickup bool, pickupTime *time.Time, pickupLocation string) (*ordermodel.Order, error)
}

type OrderHandler struct {
	orderStore   order.Store
	productStore product.Store
	orderService OrderCheckoutService
}

// --- Public ---

func NewOrderHandler(orderStore order.Store, productStore product.Store, orderService OrderCheckoutService) *OrderHandler {
	return &OrderHandler{
		orderStore:   orderStore,
		productStore: productStore,
		orderService: orderService,
	}
}

func (orderHandler *OrderHandler) Checkout(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var request dto.CheckoutRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	orderResult, error := orderHandler.orderService.Checkout(
		httpRequest.Context(), 
		userID, 
		request.CouponCode, 
		request.IsPickup, 
		request.PickupTime, 
		request.PickupLocation,
	)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	response.SendJSON(responseWriter, http.StatusCreated, dto.MapOrderToResponse(*orderResult))
}

func (orderHandler *OrderHandler) GetHistory(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)
	
	orderHistory, error := orderHandler.orderStore.GetByUserID(httpRequest.Context(), userID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapOrdersToResponse(orderHistory))
}

func (orderHandler *OrderHandler) GetLatest(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)
	
	latestOrder, error := orderHandler.orderStore.GetLatestByUserID(httpRequest.Context(), userID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapOrderToResponse(latestOrder))
}

func (orderHandler *OrderHandler) GetPickups(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID := httpRequest.Context().Value(middleware.UserIDKey).(int)
	
	pickupOrders, error := orderHandler.orderStore.GetPickupsByUserID(httpRequest.Context(), userID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}
	
	response.SendJSON(responseWriter, http.StatusOK, dto.MapOrdersToResponse(pickupOrders))
}
