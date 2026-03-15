package favorite

import (
	productmodel "coffeebase-api/internal/models/product"
	"context"
)

// --- Public ---

func (store *postgresStore) Add(requestContext context.Context, userID, productID int) error {
	query := `INSERT INTO favorites (user_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, error := store.databaseConnection.ExecContext(requestContext, query, userID, productID)
	return error
}

func (store *postgresStore) Remove(requestContext context.Context, userID, productID int) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND product_id = $2`
	_, error := store.databaseConnection.ExecContext(requestContext, query, userID, productID)
	return error
}

func (store *postgresStore) GetUserFavorites(requestContext context.Context, userID int) ([]productmodel.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.category 
		FROM products p
		JOIN favorites f ON p.id = f.product_id
		WHERE f.user_id = $1`
	
	rows, error := store.databaseConnection.QueryContext(requestContext, query, userID)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var products []productmodel.Product
	for rows.Next() {
		var productInstance productmodel.Product
		if error := rows.Scan(&productInstance.ID, &productInstance.Name, &productInstance.Description, &productInstance.Price, &productInstance.Category); error != nil {
			return nil, error
		}
		products = append(products, productInstance)
	}
	return products, nil
}
