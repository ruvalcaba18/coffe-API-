package coupon

import (
	couponmodel "coffeebase-api/internal/models/coupon"
	"context"
	"database/sql"
)

// --- Public ---

func (store *postgresStore) Create(requestContext context.Context, coupon *couponmodel.Coupon) error {
	query := `INSERT INTO coupons (code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, is_active) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	return store.databaseConnection.QueryRowContext(requestContext, query, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MinPurchaseAmount, coupon.MaxDiscountAmount, coupon.StartDate, coupon.EndDate, coupon.UsageLimit, coupon.IsActive).Scan(&coupon.ID)
}

func (store *postgresStore) GetByCode(requestContext context.Context, code string) (couponmodel.Coupon, error) {
	var couponInstance couponmodel.Coupon
	query := `SELECT id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, used_count, is_active, created_at FROM coupons WHERE code = $1`
	error := store.databaseConnection.QueryRowContext(requestContext, query, code).Scan(&couponInstance.ID, &couponInstance.Code, &couponInstance.DiscountType, &couponInstance.DiscountValue, &couponInstance.MinPurchaseAmount, &couponInstance.MaxDiscountAmount, &couponInstance.StartDate, &couponInstance.EndDate, &couponInstance.UsageLimit, &couponInstance.UsedCount, &couponInstance.IsActive, &couponInstance.CreatedAt)
	return couponInstance, error
}

func (store *postgresStore) IncrementUsage(requestContext context.Context, transaction *sql.Tx, code string) error {
	query := `UPDATE coupons SET used_count = used_count + 1 WHERE code = $1`
	if transaction != nil {
		_, error := transaction.ExecContext(requestContext, query, code)
		return error
	}
	_, error := store.databaseConnection.ExecContext(requestContext, query, code)
	return error
}

func (store *postgresStore) GetAll(requestContext context.Context) ([]couponmodel.Coupon, error) {
	query := `SELECT id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, used_count, is_active, created_at FROM coupons ORDER BY created_at DESC`
	rows, error := store.databaseConnection.QueryContext(requestContext, query)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var coupons []couponmodel.Coupon
	for rows.Next() {
		var couponInstance couponmodel.Coupon
		if error := rows.Scan(&couponInstance.ID, &couponInstance.Code, &couponInstance.DiscountType, &couponInstance.DiscountValue, &couponInstance.MinPurchaseAmount, &couponInstance.MaxDiscountAmount, &couponInstance.StartDate, &couponInstance.EndDate, &couponInstance.UsageLimit, &couponInstance.UsedCount, &couponInstance.IsActive, &couponInstance.CreatedAt); error != nil {
			return nil, error
		}
		coupons = append(coupons, couponInstance)
	}
	return coupons, nil
}

func (store *postgresStore) ToggleStatus(requestContext context.Context, id int, isActive bool) error {
	_, error := store.databaseConnection.ExecContext(requestContext, `UPDATE coupons SET is_active = $1 WHERE id = $2`, isActive, id)
	return error
}

func (store *postgresStore) Delete(requestContext context.Context, id int) error {
	_, error := store.databaseConnection.ExecContext(requestContext, `DELETE FROM coupons WHERE id = $1`, id)
	return error
}
