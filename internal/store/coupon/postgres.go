package coupon

import (
	"coffeebase-api/internal/models/coupon"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(c *coupon.Coupon) error {
	query := `INSERT INTO coupons (code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, is_active) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	return s.db.QueryRow(query, c.Code, c.DiscountType, c.DiscountValue, c.MinPurchaseAmount, c.MaxDiscountAmount, c.StartDate, c.EndDate, c.UsageLimit, c.IsActive).Scan(&c.ID)
}

func (s *Store) GetByCode(code string) (coupon.Coupon, error) {
	var c coupon.Coupon
	query := `SELECT id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, used_count, is_active, created_at FROM coupons WHERE code = $1`
	err := s.db.QueryRow(query, code).Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinPurchaseAmount, &c.MaxDiscountAmount, &c.StartDate, &c.EndDate, &c.UsageLimit, &c.UsedCount, &c.IsActive, &c.CreatedAt)
	return c, err
}

func (s *Store) IncrementUsage(tx *sql.Tx, code string) error {
	query := `UPDATE coupons SET used_count = used_count + 1 WHERE code = $1`
	var err error
	if tx != nil {
		_, err = tx.Exec(query, code)
	} else {
		_, err = s.db.Exec(query, code)
	}
	return err
}

func (s *Store) GetAll() ([]coupon.Coupon, error) {
	rows, err := s.db.Query(`SELECT id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, used_count, is_active, created_at FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []coupon.Coupon
	for rows.Next() {
		var c coupon.Coupon
		if err := rows.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinPurchaseAmount, &c.MaxDiscountAmount, &c.StartDate, &c.EndDate, &c.UsageLimit, &c.UsedCount, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}
	return coupons, nil
}
func (s *Store) ToggleStatus(id int, isActive bool) error {
	query := `UPDATE coupons SET is_active = $1 WHERE id = $2`
	_, err := s.db.Exec(query, isActive, id)
	return err
}

func (s *Store) Delete(id int) error {
	_, err := s.db.Exec(`DELETE FROM coupons WHERE id = $1`, id)
	return err
}
