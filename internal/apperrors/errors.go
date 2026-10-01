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
	ErrCannotModifySuperAdmin    = errors.New("cannot modify or delete a super admin")

	// Refresh token
	ErrRefreshTokenInvalid  = errors.New("invalid or expired refresh token")
	ErrRefreshTokenRevoked  = errors.New("session revoked, please login again")

	// QR attendance
	ErrQRInvalid            = errors.New("invalid or expired QR code")
	ErrQRAlreadyUsed        = errors.New("QR code already used")
	ErrCheckInAlreadyDone   = errors.New("check-in already registered for today")
	ErrNoCheckInForCheckOut = errors.New("no check-in found for today, cannot check-out")

	// Subordinados
	ErrNotASubordinate      = errors.New("user is not a subordinate")
)
