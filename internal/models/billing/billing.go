package billing

import "time"

type PaymentMethod struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Last4     string    `json:"last4"`
	Expiry    string    `json:"expiry"`
	Brand     string    `json:"brand"`
	Holder    string    `json:"holder"`
	IsDefault bool      `json:"isDefault"`
	CreatedAt time.Time `json:"createdAt"`
}

type Wallet struct {
	Balance float64 `json:"balance"`
}
