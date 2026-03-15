package product

import (
	productmodel "coffeebase-api/internal/models/product"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- Public ---

func (store *postgresStore) GetAll(requestContext context.Context, filter productmodel.Filter) ([]productmodel.Product, error) {
	key := fmt.Sprintf("products:q=%s:c=%s:min=%.2f:max=%.2f", filter.Query, filter.Category, filter.MinPrice, filter.MaxPrice)

	if cached, error := store.cacheService.Get(requestContext, key); error == nil {
		var products []productmodel.Product
		if error := json.Unmarshal([]byte(cached), &products); error == nil {
			return products, nil
		}
	}

	query := `
		SELECT p.id, p.name, p.description, p.price, p.category, 
		       p.average_rating, p.review_count
		FROM products p
		WHERE 1=1`
	var args []interface{}
	counter := 1

	if filter.Query != "" {
		query += fmt.Sprintf(" AND (p.name ILIKE $%d OR p.description ILIKE $%d)", counter, counter+1)
		args = append(args, "%"+filter.Query+"%", "%"+filter.Query+"%")
		counter += 2
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND p.category = $%d", counter)
		args = append(args, filter.Category)
		counter++
	}
	if filter.MinPrice > 0 {
		query += fmt.Sprintf(" AND p.price >= $%d", counter)
		args = append(args, filter.MinPrice)
		counter++
	}
	if filter.MaxPrice > 0 {
		query += fmt.Sprintf(" AND p.price <= $%d", counter)
		args = append(args, filter.MaxPrice)
		counter++
	}

	rows, error := store.databaseConnection.QueryContext(requestContext, query, args...)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var products []productmodel.Product
	for rows.Next() {
		var productInstance productmodel.Product
		if error := rows.Scan(&productInstance.ID, &productInstance.Name, &productInstance.Description, &productInstance.Price, &productInstance.Category, &productInstance.AverageRating, &productInstance.ReviewCount); error != nil {
			return nil, error
		}
		products = append(products, productInstance)
	}

	if data, error := json.Marshal(products); error == nil {
		store.cacheService.Set(requestContext, key, data, 5*time.Minute)
	}

	return products, nil
}

func (store *postgresStore) GetByID(requestContext context.Context, id int) (productmodel.Product, error) {
	key := fmt.Sprintf("product:%d", id)

	if cached, error := store.cacheService.Get(requestContext, key); error == nil {
		var productInstance productmodel.Product
		if error := json.Unmarshal([]byte(cached), &productInstance); error == nil {
			return productInstance, nil
		}
	}

	var productInstance productmodel.Product
	query := `SELECT id, name, description, price, category, average_rating, review_count FROM products WHERE id = $1`
	error := store.databaseConnection.QueryRowContext(requestContext, query, id).Scan(&productInstance.ID, &productInstance.Name, &productInstance.Description, &productInstance.Price, &productInstance.Category, &productInstance.AverageRating, &productInstance.ReviewCount)
	
	if error == nil {
		if data, error := json.Marshal(productInstance); error == nil {
			store.cacheService.Set(requestContext, key, data, 10*time.Minute)
		}
	}

	return productInstance, error
}

func (store *postgresStore) Create(requestContext context.Context, productInstance *productmodel.Product) error {
	query := `INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4) RETURNING id`
	error := store.databaseConnection.QueryRowContext(requestContext, query, productInstance.Name, productInstance.Description, productInstance.Price, productInstance.Category).Scan(&productInstance.ID)
	if error == nil {
		store.cacheService.Del(requestContext, "all_products")
	}
	return error
}

func (store *postgresStore) Update(requestContext context.Context, productInstance *productmodel.Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, category = $4 WHERE id = $5`
	_, error := store.databaseConnection.ExecContext(requestContext, query, productInstance.Name, productInstance.Description, productInstance.Price, productInstance.Category, productInstance.ID)
	if error == nil {
		store.cacheService.Del(requestContext, "all_products")
		store.cacheService.Del(requestContext, fmt.Sprintf("product:%d", productInstance.ID))
	}
	return error
}

func (store *postgresStore) Delete(requestContext context.Context, id int) error {
	query := "DELETE FROM products WHERE id = $1"
	_, error := store.databaseConnection.ExecContext(requestContext, query, id)
	if error == nil {
		store.cacheService.Del(requestContext, "all_products")
		store.cacheService.Del(requestContext, fmt.Sprintf("product:%d", id))
	}
	return error
}

func (store *postgresStore) CreateBulk(requestContext context.Context, products []productmodel.Product) error {
	transaction, error := store.databaseConnection.BeginTx(requestContext, nil)
	if error != nil {
		return error
	}
	defer transaction.Rollback()

	stmt, error := transaction.PrepareContext(requestContext, `INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4)`)
	if error != nil {
		return error
	}
	defer stmt.Close()

	for _, productInstance := range products {
		if _, error := stmt.ExecContext(requestContext, productInstance.Name, productInstance.Description, productInstance.Price, productInstance.Category); error != nil {
			return error
		}
	}

	if error := transaction.Commit(); error != nil {
		return error
	}

	store.cacheService.FlushDB(requestContext)
	return nil
}

func (store *postgresStore) GetCategories(requestContext context.Context) ([]string, error) {
	query := `SELECT DISTINCT category FROM products WHERE category != '' ORDER BY category ASC`
	rows, error := store.databaseConnection.QueryContext(requestContext, query)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		if error := rows.Scan(&name); error != nil {
			return nil, error
		}
		categories = append(categories, name)
	}

	return categories, nil
}
