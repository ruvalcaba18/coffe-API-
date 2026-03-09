package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/middleware"
	ordermodel "coffeebase-api/internal/models/order"
	productmodel "coffeebase-api/internal/models/product"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type OrderRepository interface {
	GetByUserID(userID int) ([]ordermodel.Order, error)
	GetLatestByUserID(userID int) (ordermodel.Order, error)
	GetPickupsByUserID(userID int) ([]ordermodel.Order, error)
	Create(orderInstance *ordermodel.Order) error
}

type OrderCheckoutService interface {
	Checkout(requestContext context.Context, userID int, couponCode string, isPickup bool, pickupTime *time.Time, pickupLocation string) (*ordermodel.Order, error)
}

type ProductReader interface {
	GetByID(productID int) (productmodel.Product, error)
}

type OrderHandler struct {
	OrderStore   OrderRepository
	ProductStore ProductReader
	OrderService OrderCheckoutService
}

func (orderHandler *OrderHandler) Checkout(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)

	var checkoutInput dto.CheckoutRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&checkoutInput); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	orderResult, checkoutError := orderHandler.OrderService.Checkout(request.Context(), userID, checkoutInput.CouponCode, checkoutInput.IsPickup, checkoutInput.PickupTime, checkoutInput.PickupLocation)
	if checkoutError != nil {
		http.Error(responseWriter, checkoutError.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)
	json.NewEncoder(responseWriter).Encode(dto.MapOrderToResponse(*orderResult))
}

func (orderHandler *OrderHandler) GetHistory(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)
	orderHistory, fetchError := orderHandler.OrderStore.GetByUserID(userID)
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapOrdersToResponse(orderHistory))
}

func (orderHandler *OrderHandler) GetLatest(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)
	latestOrder, fetchError := orderHandler.OrderStore.GetLatestByUserID(userID)
	if fetchError != nil {
		http.Error(responseWriter, "Order not found", http.StatusNotFound)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapOrderToResponse(latestOrder))
}

func (orderHandler *OrderHandler) GetPickups(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)
	pickupOrders, fetchError := orderHandler.OrderStore.GetPickupsByUserID(userID)
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapOrdersToResponse(pickupOrders))
}

func (orderHandler *OrderHandler) Create(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.UserIDKey).(int)
	var orderItems []ordermodel.OrderItem
	if decodeError := json.NewDecoder(request.Body).Decode(&orderItems); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	var totalAmount float64
	for _, item := range orderItems {
		productInstance, _ := orderHandler.ProductStore.GetByID(item.ProductID)
		totalAmount += productInstance.Price * float64(item.Quantity)
	}

	orderInstance := &ordermodel.Order{
		UserID: userID,
		Items:  orderItems,
		Total:  totalAmount,
	}

	if createError := orderHandler.OrderStore.Create(orderInstance); createError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)
	json.NewEncoder(responseWriter).Encode(orderInstance)
}
