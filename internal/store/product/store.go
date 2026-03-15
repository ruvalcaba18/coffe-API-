package product

import (
	"coffeebase-api/internal/cache"
	productmodel "coffeebase-api/internal/models/product"
	"context"
	"database/sql"
)

type Store interface {
	GetAll(requestContext context.Context, filter productmodel.Filter) ([]productmodel.Product, error)
	GetByID(requestContext context.Context, id int) (productmodel.Product, error)
	Create(requestContext context.Context, productInstance *productmodel.Product) error
	Update(requestContext context.Context, productInstance *productmodel.Product) error
	Delete(requestContext context.Context, id int) error
	CreateBulk(requestContext context.Context, products []productmodel.Product) error
	GetCategories(requestContext context.Context) ([]string, error)
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
