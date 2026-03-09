package product

import (
	productmodel "coffeebase-api/internal/models/product"
	"context"
	"database/sql"
	"encoding/json"
	outputFormatting "fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

/**
 * Store handles all persistence operations for the product module.
 * Refactored to eliminate all shorthands and follow strictly declarative naming.
 */
type Store struct {
	databaseConnection *sql.DB
	redisClient        *redis.Client
}

func NewStore(databaseConnection *sql.DB, redisClient *redis.Client) *Store {
	return &Store{
		databaseConnection: databaseConnection,
		redisClient:        redisClient,
	}
}

func (productStore *Store) GetAll(filter productmodel.Filter) ([]productmodel.Product, error) {
	requestContext := context.Background()
	
	// Generate dynamic cache key based on filters
	cacheUniqueIdentifier := outputFormatting.Sprintf("products:q=%s:c=%s:min=%.2f:max=%.2f", filter.Query, filter.Category, filter.MinPrice, filter.MaxPrice)

	// Try cache first
	if cachedValue, cacheError := productStore.redisClient.Get(requestContext, cacheUniqueIdentifier).Result(); cacheError == nil {
		var cachedProducts []productmodel.Product
		if unmarshalError := json.Unmarshal([]byte(cachedValue), &cachedProducts); unmarshalError == nil {
			return cachedProducts, nil
		}
	}

	// Dynamic Query Construction
	baseQuery := `
		SELECT p.id, p.name, p.description, p.price, p.category, 
		       p.average_rating, p.review_count
		FROM products p
		WHERE 1=1`
	var queryArguments []interface{}
	argumentCounter := 1

	if filter.Query != "" {
		baseQuery += outputFormatting.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argumentCounter, argumentCounter+1)
		queryArguments = append(queryArguments, "%"+filter.Query+"%", "%"+filter.Query+"%")
		argumentCounter += 2
	}
	if filter.Category != "" {
		baseQuery += outputFormatting.Sprintf(" AND category = $%d", argumentCounter)
		queryArguments = append(queryArguments, filter.Category)
		argumentCounter++
	}
	if filter.MinPrice > 0 {
		baseQuery += outputFormatting.Sprintf(" AND price >= $%d", argumentCounter)
		queryArguments = append(queryArguments, filter.MinPrice)
		argumentCounter++
	}
	if filter.MaxPrice > 0 {
		baseQuery += outputFormatting.Sprintf(" AND p.price <= $%d", argumentCounter)
		queryArguments = append(queryArguments, filter.MaxPrice)
		argumentCounter++
	}

	// No longer need GROUP BY because we aren't using aggregate functions

	rows, queryError := productStore.databaseConnection.Query(baseQuery, queryArguments...)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	var productList []productmodel.Product
	for rows.Next() {
		var productInstance productmodel.Product
		if scanError := rows.Scan(
			&productInstance.ID, 
			&productInstance.Name, 
			&productInstance.Description, 
			&productInstance.Price, 
			&productInstance.Category,
			&productInstance.AverageRating,
			&productInstance.ReviewCount,
		); scanError != nil {
			return nil, scanError
		}
		productList = append(productList, productInstance)
	}

	// Save to cache for 5 minutes
	if serializedData, serializationError := json.Marshal(productList); serializationError == nil {
		productStore.redisClient.Set(requestContext, cacheUniqueIdentifier, serializedData, 5*time.Minute)
	}

	return productList, nil
}

func (productStore *Store) GetByID(productID int) (productmodel.Product, error) {
	requestContext := context.Background()
	cacheUniqueIdentifier := outputFormatting.Sprintf("product:%d", productID)

	// Try cache
	if cachedValue, cacheError := productStore.redisClient.Get(requestContext, cacheUniqueIdentifier).Result(); cacheError == nil {
		var cachedProduct productmodel.Product
		if unmarshalError := json.Unmarshal([]byte(cachedValue), &cachedProduct); unmarshalError == nil {
			return cachedProduct, nil
		}
	}

	var productInstance productmodel.Product
	query := `
		SELECT id, name, description, price, category, average_rating, review_count
		FROM products
		WHERE id = $1`
	
	fetchError := productStore.databaseConnection.QueryRow(query, productID).
		Scan(
			&productInstance.ID, 
			&productInstance.Name, 
			&productInstance.Description, 
			&productInstance.Price, 
			&productInstance.Category,
			&productInstance.AverageRating,
			&productInstance.ReviewCount,
		)
	
	if fetchError == nil {
		if serializedData, serializationError := json.Marshal(productInstance); serializationError == nil {
			productStore.redisClient.Set(requestContext, cacheUniqueIdentifier, serializedData, 10*time.Minute)
		}
	}

	return productInstance, fetchError
}

func (productStore *Store) Create(productInstance *productmodel.Product) error {
	queryStatement := `INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4) RETURNING id`
	executionError := productStore.databaseConnection.QueryRow(queryStatement, productInstance.Name, productInstance.Description, productInstance.Price, productInstance.Category).Scan(&productInstance.ID)
	if executionError == nil {
		productStore.redisClient.Del(context.Background(), "all_products")
	}
	return executionError
}

func (productStore *Store) Update(productInstance *productmodel.Product) error {
	queryStatement := `UPDATE products SET name = $1, description = $2, price = $3, category = $4 WHERE id = $5`
	_, executionError := productStore.databaseConnection.Exec(queryStatement, productInstance.Name, productInstance.Description, productInstance.Price, productInstance.Category, productInstance.ID)
	if executionError == nil {
		requestContext := context.Background()
		productStore.redisClient.Del(requestContext, "all_products")
		productStore.redisClient.Del(requestContext, outputFormatting.Sprintf("product:%d", productInstance.ID))
	}
	return executionError
}

func (productStore *Store) Delete(productID int) error {
	_, executionError := productStore.databaseConnection.Exec("DELETE FROM products WHERE id = $1", productID)
	if executionError == nil {
		requestContext := context.Background()
		productStore.redisClient.Del(requestContext, "all_products")
		productStore.redisClient.Del(requestContext, outputFormatting.Sprintf("product:%d", productID))
	}
	return executionError
}

func (productStore *Store) CreateBulk(productCollection []productmodel.Product) error {
	databaseTransaction, transactionError := productStore.databaseConnection.Begin()
	if transactionError != nil {
		return transactionError
	}
	defer databaseTransaction.Rollback()

	insertionQuery := `INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4)`
	preparedContextStatement, preparationError := databaseTransaction.Prepare(insertionQuery)
	if preparationError != nil {
		return preparationError
	}
	defer preparedContextStatement.Close()

	for _, item := range productCollection {
		if _, executionError := preparedContextStatement.Exec(item.Name, item.Description, item.Price, item.Category); executionError != nil {
			return executionError
		}
	}

	if commitError := databaseTransaction.Commit(); commitError != nil {
		return commitError
	}

	// Invalidate cache
	productStore.redisClient.FlushDB(context.Background())
	return nil
}

func (productStore *Store) GetCategories() ([]string, error) {
	queryStatement := `SELECT DISTINCT category FROM products WHERE category != '' ORDER BY category ASC`
	rows, queryError := productStore.databaseConnection.Query(queryStatement)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	var categoryList []string
	for rows.Next() {
		var categoryName string
		if scanError := rows.Scan(&categoryName); scanError != nil {
			return nil, scanError
		}
		categoryList = append(categoryList, categoryName)
	}

	return categoryList, nil
}


