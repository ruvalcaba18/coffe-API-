package billing

import (
	billingmodel "coffeebase-api/internal/models/billing"
	"context"
)

// --- Public ---

func (store *postgresStore) GetPaymentMethodsByUserID(requestContext context.Context, userID int) ([]billingmodel.PaymentMethod, error) {
	query := `SELECT id, user_id, last4, expiry, brand, holder, is_default, created_at 
			  FROM payment_methods WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC`
	rows, error := store.databaseConnection.QueryContext(requestContext, query, userID)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	var methods []billingmodel.PaymentMethod
	for rows.Next() {
		var paymentMethod billingmodel.PaymentMethod
		if error := rows.Scan(&paymentMethod.ID, &paymentMethod.UserID, &paymentMethod.Last4, &paymentMethod.Expiry, &paymentMethod.Brand, &paymentMethod.Holder, &paymentMethod.IsDefault, &paymentMethod.CreatedAt); error != nil {
			return nil, error
		}
		methods = append(methods, paymentMethod)
	}
	return methods, nil
}

func (store *postgresStore) AddPaymentMethod(requestContext context.Context, paymentMethodInstance *billingmodel.PaymentMethod) error {
	query := `INSERT INTO payment_methods (user_id, last4, expiry, brand, holder, is_default) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return store.databaseConnection.QueryRowContext(requestContext, query, paymentMethodInstance.UserID, paymentMethodInstance.Last4, paymentMethodInstance.Expiry, paymentMethodInstance.Brand, paymentMethodInstance.Holder, paymentMethodInstance.IsDefault).Scan(&paymentMethodInstance.ID, &paymentMethodInstance.CreatedAt)
}

func (store *postgresStore) ExistsPaymentMethod(requestContext context.Context, userID int, last4, brand string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM payment_methods WHERE user_id = $1 AND last4 = $2 AND brand = $3)`
	error := store.databaseConnection.QueryRowContext(requestContext, query, userID, last4, brand).Scan(&exists)
	return exists, error
}
