package coupon

import (
	"time"
)

type Coupon struct {
	ID                int       `json:"id"`
	Code              string    `json:"code"`
	DiscountType      string    `json:"discount_type"` // "percentage" or "fixed"
	DiscountValue     float64   `json:"discount_value"`
	MinPurchaseAmount float64   `json:"min_purchase_amount"`
	MaxDiscountAmount *float64  `json:"max_discount_amount,omitempty"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	UsageLimit        int       `json:"usage_limit"`
	UsedCount         int       `json:"used_count"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

func (c *Coupon) IsValid(total float64) bool {
	now := time.Now()
	if !c.IsActive {
		return false
	}
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return false
	}
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return false
	}
	if total < c.MinPurchaseAmount {
		return false
	}
	return true
}

func (c *Coupon) CalculateDiscount(total float64) float64 {
	var discount float64
	if c.DiscountType == "percentage" {
		discount = total * (c.DiscountValue / 100)
		if c.MaxDiscountAmount != nil && discount > *c.MaxDiscountAmount {
			discount = *c.MaxDiscountAmount
		}
	} else if c.DiscountType == "fixed" {
		discount = c.DiscountValue
	}
	
	if discount > total {
		return total
	}
	return discount
}
