package cart

import (
	cartmodel "coffeebase-api/internal/models/cart"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// --- Public ---

func (store *postgresStore) UpdateItem(requestContext context.Context, userID, productID, quantity int) error {
	transaction, error := store.databaseConnection.BeginTx(requestContext, nil)
	if error != nil {
		return error
	}
	defer transaction.Rollback()

	query := `INSERT INTO carts (user_id, updated_at) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET updated_at = $2`
	_, error = transaction.ExecContext(requestContext, query, userID, time.Now())
	if error != nil {
		return error
	}

	if quantity <= 0 {
		_, error = transaction.ExecContext(requestContext, `DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`, userID, productID)
	} else {
		query := `INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3) 
						 ON CONFLICT (user_id, product_id) DO UPDATE SET quantity = $3`
		_, error = transaction.ExecContext(requestContext, query, userID, productID, quantity)
	}
	
	if error != nil {
		return error
	}

	if error := transaction.Commit(); error != nil {
		return error
	}

	store.cacheService.Del(requestContext, fmt.Sprintf("cart:v2:%d", userID))

	return nil
}

func (store *postgresStore) GetCart(requestContext context.Context, userID int) (*cartmodel.Cart, error) {
	key := fmt.Sprintf("cart:v2:%d", userID)

	if cached, error := store.cacheService.Get(requestContext, key); error == nil {
		var cart cartmodel.Cart
		if error := json.Unmarshal([]byte(cached), &cart); error == nil {
			return &cart, nil
		}
	}

	query := `
		SELECT ci.product_id, ci.quantity, p.name, p.price, COALESCE(p.image_url, '') 
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1
	`
	rows, error := store.databaseConnection.QueryContext(requestContext, query, userID)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	cart := &cartmodel.Cart{UserID: userID, Items: []cartmodel.Item{}}
	for rows.Next() {
		var item cartmodel.Item
		if error := rows.Scan(&item.ProductID, &item.Quantity, &item.ProductName, &item.Price, &item.ImageURL); error != nil {
			return nil, error
		}
		cart.Items = append(cart.Items, item)
	}

	if data, error := json.Marshal(cart); error == nil {
		store.cacheService.Set(requestContext, key, data, 1*time.Hour)
	}

	return cart, nil
}

func (store *postgresStore) ClearCart(requestContext context.Context, transaction *sql.Tx, userID int) error {
	_, error := transaction.ExecContext(requestContext, `DELETE FROM cart_items WHERE user_id = $1`, userID)
	if error == nil {
		store.cacheService.Del(requestContext, fmt.Sprintf("cart:v2:%d", userID))
	}
	return error
}
