package review

import (
	reviewmodel "coffeebase-api/internal/models/review"
	"context"
	"database/sql"
)

type Store interface {
	Create(requestContext context.Context, reviewInstance *reviewmodel.Review) error
	GetByProductID(requestContext context.Context, productID int) ([]reviewmodel.Review, error)
}

type postgresStore struct {
	databaseConnection *sql.DB
}

// --- Public ---

func NewStore(databaseConnection *sql.DB) Store {
	return &postgresStore{databaseConnection: databaseConnection}
}
