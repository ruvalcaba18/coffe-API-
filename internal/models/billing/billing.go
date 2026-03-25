package billing

import "time"

type PaymentMethod struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Last4     string    `json:"last4"`
	Expiry    string    `json:"expiry"`
	Brand     string    `json:"brand"`
	Holder    string    `json:"holder"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type Wallet struct {
	Balance float64 `json:"balance"`
}
