package response

import (
	"coffeebase-api/internal/apperrors"
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// --- Public ---

func SendJSON(responseWriter http.ResponseWriter, statusCode int, payload interface{}) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	json.NewEncoder(responseWriter).Encode(payload)
}

func SendError(responseWriter http.ResponseWriter, providedError error) {
	statusCode := http.StatusInternalServerError
	message := "an unexpected error occurred"

	switch providedError {
	case apperrors.ErrInvalidRequest, apperrors.ErrInvalidCoupon, apperrors.ErrCouponNotValidForPurchase, apperrors.ErrInvalidID, apperrors.ErrCartEmpty:
		statusCode = http.StatusBadRequest
		message = providedError.Error()
	case apperrors.ErrUnauthorized:
		statusCode = http.StatusUnauthorized
		message = providedError.Error()
	case apperrors.ErrForbidden, apperrors.ErrCannotModifySuperAdmin:
		statusCode = http.StatusForbidden
		message = providedError.Error()
	case apperrors.ErrUserNotFound, apperrors.ErrProductNotFound:
		statusCode = http.StatusNotFound
		message = providedError.Error()
	case apperrors.ErrRequestInProgress:
		statusCode = http.StatusConflict
		message = providedError.Error()
	case apperrors.ErrDuplicateCard, apperrors.ErrCouponAlreadyUsedByUser:
		statusCode = http.StatusConflict
		message = providedError.Error()
	}

	SendJSON(responseWriter, statusCode, ErrorResponse{
		Message: message,
		Code:    statusCode,
	})
}

func DecodeJSON(httpRequest *http.Request, target interface{}) error {
	if error := json.NewDecoder(httpRequest.Body).Decode(target); error != nil {
		return apperrors.ErrInvalidRequest
	}
	return nil
}

func BadRequest(responseWriter http.ResponseWriter, message string) {
	SendJSON(responseWriter, http.StatusBadRequest, ErrorResponse{Message: message, Code: http.StatusBadRequest})
}

func Unauthorized(responseWriter http.ResponseWriter, message string) {
	SendJSON(responseWriter, http.StatusUnauthorized, ErrorResponse{Message: message, Code: http.StatusUnauthorized})
}

func Forbidden(responseWriter http.ResponseWriter, message string) {
	SendJSON(responseWriter, http.StatusForbidden, ErrorResponse{Message: message, Code: http.StatusForbidden})
}

func NotFound(responseWriter http.ResponseWriter, message string) {
	SendJSON(responseWriter, http.StatusNotFound, ErrorResponse{Message: message, Code: http.StatusNotFound})
}

func InternalError(responseWriter http.ResponseWriter, message string) {
	SendJSON(responseWriter, http.StatusInternalServerError, ErrorResponse{Message: message, Code: http.StatusInternalServerError})
}
