package billing

import (
	billingmodel "coffeebase-api/internal/models/billing"
	"context"
	"database/sql"
)

type Store interface {
	GetPaymentMethodsByUserID(requestContext context.Context, userID int) ([]billingmodel.PaymentMethod, error)
	AddPaymentMethod(requestContext context.Context, method *billingmodel.PaymentMethod) error
	ExistsPaymentMethod(requestContext context.Context, userID int, last4, brand string) (bool, error)
}

type postgresStore struct {
	databaseConnection *sql.DB
}

// --- Public ---

func NewStore(databaseConnection *sql.DB) Store {
	return &postgresStore{databaseConnection: databaseConnection}
}
