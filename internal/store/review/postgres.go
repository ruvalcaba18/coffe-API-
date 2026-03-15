package review

import (
	reviewmodel "coffeebase-api/internal/models/review"
	"context"
)

// --- Public ---

func (store *postgresStore) Create(requestContext context.Context, review *reviewmodel.Review) error {
	query := `INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return store.databaseConnection.QueryRowContext(requestContext, query, review.ProductID, review.UserID, review.Rating, review.Comment).Scan(&review.ID, &review.CreatedAt)
}

func (store *postgresStore) GetByProductID(requestContext context.Context, productID int) ([]reviewmodel.Review, error) {
	query := "SELECT id, product_id, user_id, rating, comment, created_at FROM reviews WHERE product_id = $1"
	rows, error := store.databaseConnection.QueryContext(requestContext, query, productID)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var reviews []reviewmodel.Review
	for rows.Next() {
		var reviewInstance reviewmodel.Review
		if error := rows.Scan(&reviewInstance.ID, &reviewInstance.ProductID, &reviewInstance.UserID, &reviewInstance.Rating, &reviewInstance.Comment, &reviewInstance.CreatedAt); error != nil {
			return nil, error
		}
		reviews = append(reviews, reviewInstance)
	}
	return reviews, nil
}
