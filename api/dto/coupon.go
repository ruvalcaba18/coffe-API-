package dto

import (
	couponmodel "coffeebase-api/internal/models/coupon"
	"time"
)

type CouponRequest struct {
	Code              string    `json:"code"`
	DiscountType      string    `json:"discount_type"`
	DiscountValue     float64   `json:"discount_value"`
	MinPurchaseAmount float64   `json:"min_purchase_amount"`
	MaxDiscountAmount *float64  `json:"max_discount_amount,omitempty"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	UsageLimit        int       `json:"usage_limit"`
	IsActive          bool      `json:"is_active"`
}

type CouponResponse struct {
	ID                int       `json:"id"`
	Code              string    `json:"code"`
	DiscountType      string    `json:"discount_type"`
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

func MapCouponToResponse(c couponmodel.Coupon) CouponResponse {
	return CouponResponse{
		ID:                c.ID,
		Code:              c.Code,
		DiscountType:      c.DiscountType,
		DiscountValue:     c.DiscountValue,
		MinPurchaseAmount: c.MinPurchaseAmount,
		MaxDiscountAmount: c.MaxDiscountAmount,
		StartDate:         c.StartDate,
		EndDate:           c.EndDate,
		UsageLimit:        c.UsageLimit,
		UsedCount:         c.UsedCount,
		IsActive:          c.IsActive,
		CreatedAt:         c.CreatedAt,
	}
}

func MapCouponsToResponse(coupons []couponmodel.Coupon) []CouponResponse {
	dtos := make([]CouponResponse, len(coupons))
	for i, c := range coupons {
		dtos[i] = MapCouponToResponse(c)
	}
	return dtos
}
