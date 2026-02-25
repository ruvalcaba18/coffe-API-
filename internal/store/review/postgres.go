package review

import (
	reviewmodel "coffeebase-api/internal/models/review"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(r *reviewmodel.Review) error {
	query := `INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return s.db.QueryRow(query, r.ProductID, r.UserID, r.Rating, r.Comment).Scan(&r.ID, &r.CreatedAt)
}

func (s *Store) GetByProductID(productID int) ([]reviewmodel.Review, error) {
	rows, err := s.db.Query("SELECT id, product_id, user_id, rating, comment, created_at FROM reviews WHERE product_id = $1", productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []reviewmodel.Review
	for rows.Next() {
		var r reviewmodel.Review
		if err := rows.Scan(&r.ID, &r.ProductID, &r.UserID, &r.Rating, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, nil
}
