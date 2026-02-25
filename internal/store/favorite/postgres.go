package favorite

import (
	productmodel "coffeebase-api/internal/models/product"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Add(userID, productID int) error {
	query := `INSERT INTO favorites (user_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(query, userID, productID)
	return err
}

func (s *Store) Remove(userID, productID int) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND product_id = $2`
	_, err := s.db.Exec(query, userID, productID)
	return err
}

func (s *Store) GetUserFavorites(userID int) ([]productmodel.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.category 
		FROM products p
		JOIN favorites f ON p.id = f.product_id
		WHERE f.user_id = $1`
	
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []productmodel.Product
	for rows.Next() {
		var p productmodel.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
