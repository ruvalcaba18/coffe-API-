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
	db           *sql.DB
	rdb          *redis.Client
	orderStore   *orderstore.Store
	cartStore    *cartstore.Store
	productStore *productstore.Store
	couponStore  *couponstore.Store
}

func NewService(db *sql.DB, rdb *redis.Client, os *orderstore.Store, cs *cartstore.Store, ps *productstore.Store, co *couponstore.Store) *Service {
	return &Service{
		db:           db,
		rdb:          rdb,
		orderStore:   os,
		cartStore:    cs,
		productStore: ps,
		couponStore:  co,
	}
}

func (s *Service) Checkout(userID int, couponCode string) (*ordermodel.Order, error) {
	ctx := context.Background()
	lockKey := fmt.Sprintf("lock:checkout:%d", userID)

	// Idempotency: Use Redis lock to prevent double clicks/submissions (5 second window)
	success, err := s.rdb.SetNX(ctx, lockKey, "locked", 5*time.Second).Result()
	if err != nil || !success {
		return nil, errors.New("request already in progress, please wait")
	}

	// ACID: Start DB Transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Get Cart
	cart, err := s.cartStore.GetCart(userID)
	if err != nil || len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// 2. Validate products and calculate total
	var orderItems []ordermodel.OrderItem
	var total float64
	for _, item := range cart.Items {
		p, err := s.productStore.GetByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %d not found", item.ProductID)
		}
		orderItems = append(orderItems, ordermodel.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
		total += p.Price * float64(item.Quantity)
	}

	o := &ordermodel.Order{
		UserID: userID,
		Items:  orderItems,
		Total:  total,
	}

	// 3. Handle Coupon
	if couponCode != "" {
		c, err := s.couponStore.GetByCode(couponCode)
		if err != nil {
			return nil, errors.New("invalid coupon code")
		}
		if !c.IsValid(total) {
			return nil, errors.New("coupon is not valid for this purchase")
		}

		discount := c.CalculateDiscount(total)
		o.Total = total - discount
		o.CouponCode = couponCode
		o.DiscountAmount = discount

		// Increment usage in DB within transaction
		if err := s.couponStore.IncrementUsage(tx, couponCode); err != nil {
			return nil, err
		}
	}

	// 4. Create Order using the transaction
	if err := s.orderStore.CreateWithTx(tx, o); err != nil {
		return nil, err
	}

	// 4. Clear Cart using the transaction
	if err := s.cartStore.ClearCart(tx, userID); err != nil {
		return nil, err
	}

	// 5. Commit Transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 6. Invalidate Cart Cache in Redis
	s.rdb.Del(ctx, fmt.Sprintf("cart:%d", userID))

	return o, nil
}
