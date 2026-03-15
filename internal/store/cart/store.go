package cart

import (
	"coffeebase-api/internal/cache"
	cartmodel "coffeebase-api/internal/models/cart"
	"context"
	"database/sql"
)

type Store interface {
	UpdateItem(requestContext context.Context, userID, productID, quantity int) error
	GetCart(requestContext context.Context, userID int) (*cartmodel.Cart, error)
	ClearCart(requestContext context.Context, transaction *sql.Tx, userID int) error
}

type postgresStore struct {
	databaseConnection *sql.DB
	cacheService       cache.Service
}

// --- Public ---

func NewStore(databaseConnection *sql.DB, cacheService cache.Service) Store {
	return &postgresStore{
		databaseConnection: databaseConnection,
		cacheService:       cacheService,
	}
}
