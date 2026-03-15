package coupon

import (
	couponmodel "coffeebase-api/internal/models/coupon"
	"context"
	"database/sql"
)

type Store interface {
	Create(requestContext context.Context, couponInstance *couponmodel.Coupon) error
	GetByCode(requestContext context.Context, code string) (couponmodel.Coupon, error)
	IncrementUsage(requestContext context.Context, transaction *sql.Tx, code string) error
	GetAll(requestContext context.Context) ([]couponmodel.Coupon, error)
	ToggleStatus(requestContext context.Context, id int, isActive bool) error
	Delete(requestContext context.Context, id int) error
}

type postgresStore struct {
	databaseConnection *sql.DB
}

// --- Public ---

func NewStore(databaseConnection *sql.DB) Store {
	return &postgresStore{databaseConnection: databaseConnection}
}
