package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/cache"
	ordermodel "coffeebase-api/internal/models/order"
	cartstore "coffeebase-api/internal/store/cart"
	couponstore "coffeebase-api/internal/store/coupon"
	orderstore "coffeebase-api/internal/store/order"
	productstore "coffeebase-api/internal/store/product"
)

type Service struct {
	databaseConnection *sql.DB
	cacheService       cache.Service
	orderStore         orderstore.Store
	cartStore          cartstore.Store
	productStore       productstore.Store
	couponStore        couponstore.Store
}

// --- Public ---

func NewService(
	databaseConnection *sql.DB,
	cacheService cache.Service,
	orderStore orderstore.Store,
	cartStore cartstore.Store,
	productStore productstore.Store,
	couponStore couponstore.Store,
) *Service {
	return &Service{
		databaseConnection: databaseConnection,
		cacheService:       cacheService,
		orderStore:         orderStore,
		cartStore:          cartStore,
		productStore:       productStore,
		couponStore:        couponStore,
	}
}

func (service *Service) Checkout(
	requestContext context.Context,
	userID int,
	couponCode string,
	isPickup bool,
	pickupTime *time.Time,
	pickupLocation string,
) (*ordermodel.Order, error) {
	if error := service.acquireCheckoutLock(requestContext, userID); error != nil {
		return nil, error
	}

	databaseTransaction, error := service.databaseConnection.BeginTx(requestContext, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if error != nil {
		return nil, error
	}
	defer databaseTransaction.Rollback()

	orderInstance, error := service.executeCheckoutLogic(requestContext, databaseTransaction, userID, couponCode, isPickup, pickupTime, pickupLocation)
	if error != nil {
		return nil, error
	}

	if error := databaseTransaction.Commit(); error != nil {
		return nil, error
	}

	service.invalidateUserCaches(requestContext, userID)

	return orderInstance, nil
}

// --- Private ---

func (service *Service) acquireCheckoutLock(requestContext context.Context, userID int) error {
	lockKey := fmt.Sprintf("lock:checkout:%d", userID)
	lockAcquired, error := service.cacheService.SetNX(requestContext, lockKey, "locked", 5*time.Second)
	if error != nil || !lockAcquired {
		return apperrors.ErrRequestInProgress
	}
	return nil
}

func (service *Service) executeCheckoutLogic(
	requestContext context.Context,
	transaction *sql.Tx,
	userID int,
	couponCode string,
	isPickup bool,
	pickupTime *time.Time,
	pickupLocation string,
) (*ordermodel.Order, error) {
	
	items, total, error := service.prepareOrderItemsFromCart(requestContext, userID)
	if error != nil {
		return nil, error
	}

	orderInstance := &ordermodel.Order{
		UserID:         userID,
		Items:          items,
		Total:          total,
		IsPickup:       isPickup,
		PickupTime:     pickupTime,
		PickupLocation: pickupLocation,
	}

	if couponCode != "" {
		if error := service.applyCouponToOrder(requestContext, transaction, orderInstance, couponCode, userID); error != nil {
			return nil, error
		}
	}

	if error := service.orderStore.CreateWithTx(requestContext, transaction, orderInstance); error != nil {
		return nil, error
	}

	if couponCode != "" && orderInstance.ID != "" {
		if error := service.couponStore.RecordUserCouponUsage(requestContext, transaction, userID, couponCode, orderInstance.ID); error != nil {
			return nil, error
		}
	}

	if error := service.cartStore.ClearCart(requestContext, transaction, userID); error != nil {
		return nil, error
	}

	return orderInstance, nil
}

func (service *Service) prepareOrderItemsFromCart(requestContext context.Context, userID int) ([]ordermodel.OrderItem, float64, error) {
	userCart, error := service.cartStore.GetCart(requestContext, userID)
	if error != nil || len(userCart.Items) == 0 {
		return nil, 0, apperrors.ErrCartEmpty
	}

	var orderItems []ordermodel.OrderItem
	var totalAmount float64

	for _, cartItem := range userCart.Items {
		productInstance, error := service.productStore.GetByID(requestContext, cartItem.ProductID)
		if error != nil {
			return nil, 0, fmt.Errorf("%w: %d", apperrors.ErrProductNotFound, cartItem.ProductID)
		}
		
		orderItems = append(orderItems, ordermodel.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
		})
		totalAmount += productInstance.Price * float64(cartItem.Quantity)
	}

	return orderItems, totalAmount, nil
}

func (service *Service) applyCouponToOrder(requestContext context.Context, transaction *sql.Tx, order *ordermodel.Order, code string, userID int) error {
	couponInstance, error := service.couponStore.GetByCode(requestContext, code)
	if error != nil {
		return apperrors.ErrInvalidCoupon
	}

	if !couponInstance.IsValid(order.Total) {
		return apperrors.ErrCouponNotValidForPurchase
	}

	alreadyUsed, error := service.couponStore.HasUserUsedCoupon(requestContext, userID, code)
	if error != nil {
		return apperrors.ErrInternalServerError
	}
	if alreadyUsed {
		return apperrors.ErrCouponAlreadyUsedByUser
	}

	discountAmount := couponInstance.CalculateDiscount(order.Total)
	order.Total -= discountAmount
	order.CouponCode = code
	order.DiscountAmount = discountAmount

	return service.couponStore.IncrementUsage(requestContext, transaction, code)
}

func (service *Service) invalidateUserCaches(requestContext context.Context, userID int) {
	cacheKey := fmt.Sprintf("cart:%d", userID)
	service.cacheService.Del(requestContext, cacheKey)
}
