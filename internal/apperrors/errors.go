package apperrors

import "errors"

// --- Public ---

var (
	ErrCartEmpty                = errors.New("cart is empty")
	ErrProductNotFound           = errors.New("product not found")
	ErrInvalidCoupon             = errors.New("invalid coupon code")
	ErrCouponNotValidForPurchase = errors.New("coupon is not valid for this purchase")
	ErrRequestInProgress         = errors.New("request already in progress, please wait")
	ErrInternalServerError       = errors.New("internal server error")
	ErrInvalidRequest            = errors.New("invalid request body")
	ErrUserNotFound              = errors.New("user not found")
	ErrUnauthorized              = errors.New("unauthorized")
	ErrForbidden                 = errors.New("forbidden")
	ErrInvalidID                 = errors.New("invalid identifier")
	ErrDuplicateCard             = errors.New("this card is already registered")
	ErrCouponAlreadyUsedByUser   = errors.New("you have already used this coupon")
)
