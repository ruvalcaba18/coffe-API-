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

func MapCouponToResponse(couponInstance couponmodel.Coupon) CouponResponse {
	return CouponResponse{
		ID:                couponInstance.ID,
		Code:              couponInstance.Code,
		DiscountType:      couponInstance.DiscountType,
		DiscountValue:     couponInstance.DiscountValue,
		MinPurchaseAmount: couponInstance.MinPurchaseAmount,
		MaxDiscountAmount: couponInstance.MaxDiscountAmount,
		StartDate:         couponInstance.StartDate,
		EndDate:           couponInstance.EndDate,
		UsageLimit:        couponInstance.UsageLimit,
		UsedCount:         couponInstance.UsedCount,
		IsActive:          couponInstance.IsActive,
		CreatedAt:         couponInstance.CreatedAt,
	}
}

func MapCouponsToResponse(coupons []couponmodel.Coupon) []CouponResponse {
	dtos := make([]CouponResponse, len(coupons))
	for index, couponInstance := range coupons {
		dtos[index] = MapCouponToResponse(couponInstance)
	}
	return dtos
}
