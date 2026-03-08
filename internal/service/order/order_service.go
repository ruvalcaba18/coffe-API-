package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	ordermodel "coffeebase-api/internal/models/order"
	cartstore "coffeebase-api/internal/store/cart"
	couponstore "coffeebase-api/internal/store/coupon"
	orderstore "coffeebase-api/internal/store/order"
	productstore "coffeebase-api/internal/store/product"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	databaseConnection *sql.DB
	redisClient        *redis.Client
	orderStore         *orderstore.Store
	cartStore          *cartstore.Store
	productStore       *productstore.Store
	couponStore        *couponstore.Store
}

func NewService(
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	orderStore *orderstore.Store,
	cartStore *cartstore.Store,
	productStore *productstore.Store,
	couponStore *couponstore.Store,
) *Service {
	return &Service{
		databaseConnection: databaseConnection,
		redisClient:        redisClient,
		orderStore:         orderStore,
		cartStore:          cartStore,
		productStore:       productStore,
		couponStore:        couponStore,
	}
}

func (orderService *Service) Checkout(
	requestContext context.Context,
	userID int,
	couponCode string,
	isPickup bool,
	pickupTime *time.Time,
	pickupLocation string,
) (*ordermodel.Order, error) {
	lockKey := fmt.Sprintf("lock:checkout:%d", userID)

	// Idempotency: Use Redis lock to prevent double clicks/submissions (5 second window)
	lockAcquired, lockError := orderService.redisClient.SetNX(requestContext, lockKey, "locked", 5*time.Second).Result()
	if lockError != nil || !lockAcquired {
		return nil, errors.New("request already in progress, please wait")
	}

	// ACID: Start DB Transaction
	databaseTransaction, transactionBeginError := orderService.databaseConnection.BeginTx(requestContext, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if transactionBeginError != nil {
		return nil, transactionBeginError
	}
	defer databaseTransaction.Rollback()

	// 1. Get Cart
	userCart, cartFetchError := orderService.cartStore.GetCart(userID)
	if cartFetchError != nil || len(userCart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// 2. Validate products and calculate total
	var orderItems []ordermodel.OrderItem
	var totalAmount float64
	for _, item := range userCart.Items {
		productInstance, productFetchError := orderService.productStore.GetByID(item.ProductID)
		if productFetchError != nil {
			return nil, fmt.Errorf("product %d not found", item.ProductID)
		}
		orderItems = append(orderItems, ordermodel.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
		totalAmount += productInstance.Price * float64(item.Quantity)
	}

	orderInstance := &ordermodel.Order{
		UserID:         userID,
		Items:          orderItems,
		Total:          totalAmount,
		IsPickup:       isPickup,
		PickupTime:     pickupTime,
		PickupLocation: pickupLocation,
	}

	// 3. Handle Coupon
	if couponCode != "" {
		couponInstance, couponFetchError := orderService.couponStore.GetByCode(couponCode)
		if couponFetchError != nil {
			return nil, errors.New("invalid coupon code")
		}
		if !couponInstance.IsValid(totalAmount) {
			return nil, errors.New("coupon is not valid for this purchase")
		}

		discountAmount := couponInstance.CalculateDiscount(totalAmount)
		orderInstance.Total = totalAmount - discountAmount
		orderInstance.CouponCode = couponCode
		orderInstance.DiscountAmount = discountAmount

		// Increment usage in DB within transaction
		if incrementError := orderService.couponStore.IncrementUsage(databaseTransaction, couponCode); incrementError != nil {
			return nil, incrementError
		}
	}

	// 4. Create Order using the transaction
	if createError := orderService.orderStore.CreateWithTx(databaseTransaction, orderInstance); createError != nil {
		return nil, createError
	}

	// 4. Clear Cart using the transaction
	if clearError := orderService.cartStore.ClearCart(databaseTransaction, userID); clearError != nil {
		return nil, clearError
	}

	// 5. Commit Transaction
	if commitError := databaseTransaction.Commit(); commitError != nil {
		return nil, commitError
	}

	// 6. Invalidate Cart Cache in Redis
	orderService.redisClient.Del(requestContext, fmt.Sprintf("cart:%d", userID))

	return orderInstance, nil
}
