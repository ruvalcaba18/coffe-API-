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

func (coupon *Coupon) IsValid(total float64) bool {
	now := time.Now()
	if !coupon.IsActive {
		return false
	}
	if now.Before(coupon.StartDate) || now.After(coupon.EndDate) {
		return false
	}
	if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
		return false
	}
	if total < coupon.MinPurchaseAmount {
		return false
	}
	return true
}

func (coupon *Coupon) CalculateDiscount(total float64) float64 {
	var discount float64
	if coupon.DiscountType == "percentage" {
		discount = total * (coupon.DiscountValue / 100)
		if coupon.MaxDiscountAmount != nil && discount > *coupon.MaxDiscountAmount {
			discount = *coupon.MaxDiscountAmount
		}
	} else if coupon.DiscountType == "fixed" {
		discount = coupon.DiscountValue
	}
	
	if discount > total {
		return total
	}
	return discount
}
