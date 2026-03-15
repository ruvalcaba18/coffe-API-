package favorite

import (
	productmodel "coffeebase-api/internal/models/product"
	"context"
	"database/sql"
)

type Store interface {
	Add(requestContext context.Context, userID, productID int) error
	Remove(requestContext context.Context, userID, productID int) error
	GetUserFavorites(requestContext context.Context, userID int) ([]productmodel.Product, error)
}

type postgresStore struct {
	databaseConnection *sql.DB
}

// --- Public ---

func NewStore(databaseConnection *sql.DB) Store {
	return &postgresStore{databaseConnection: databaseConnection}
}
