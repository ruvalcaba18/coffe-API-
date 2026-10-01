package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	attendancestore "coffeebase-api/internal/store/attendance"
	qrservice "coffeebase-api/internal/service/attendance"
	"coffeebase-api/internal/middleware"
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type AttendanceHandler struct {
	attendanceStore attendancestore.Store
}

// AttendanceStoreIface permite testear con mocks.
type AttendanceStoreIface interface {
	attendancestore.Store
}

// --- Public ---

func NewAttendanceHandler(attendanceStore attendancestore.Store) *AttendanceHandler {
	return &AttendanceHandler{attendanceStore: attendanceStore}
}

// CheckIn registra la entrada del empleado autenticado usando el payload del QR.
// POST /attendance/check-in
func (h *AttendanceHandler) CheckIn(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID, err := getUserIDFromContext(httpRequest.Context())
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrUnauthorized)
		return
	}

	var req dto.QRCheckRequest
	if err := response.DecodeJSON(httpRequest, &req); err != nil {
		response.SendError(responseWriter, err)
		return
	}
	if req.Payload == "" {
		response.SendError(responseWriter, apperrors.ErrInvalidRequest)
		return
	}

	// Validar firma HMAC del QR
	payload, err := qrservice.ValidateQRToken(req.Payload)
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrQRInvalid)
		return
	}

	// El QR debe corresponder al usuario autenticado
	if payload.UserID != userID {
		response.SendError(responseWriter, apperrors.ErrForbidden)
		return
	}

	// Buscar el token en DB por hash (verifica que no sea falso)
	tokenHash := qrservice.HashQRToken(payload.Token)
	qrToken, err := h.attendanceStore.GetQRTokenByHash(httpRequest.Context(), tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(responseWriter, apperrors.ErrQRInvalid)
		} else {
			response.SendError(responseWriter, apperrors.ErrInternalServerError)
		}
		return
	}

	// Verificar que no fue usado ya para check-in
	if qrToken.UsedCheckIn {
		response.SendError(responseWriter, apperrors.ErrQRAlreadyUsed)
		return
	}

	ip := extractClientIP(httpRequest)

	// Registrar check-in (UPSERT — si ya hay uno hoy, lo sobreescribe)
	if err := h.attendanceStore.RecordCheckIn(httpRequest.Context(), userID, ip, qrToken.ID); err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	// Marcar QR como usado para check-in
	_ = h.attendanceStore.MarkQRUsedCheckIn(httpRequest.Context(), qrToken.ID)

	record, _ := h.attendanceStore.GetTodayByUserID(httpRequest.Context(), userID)
	if record != nil {
		response.SendJSON(responseWriter, http.StatusOK, dto.MapAttendanceToResponse(*record))
		return
	}
	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"message": "check-in registrado"})
}

// CheckOut registra la salida del empleado autenticado.
// POST /attendance/check-out
func (h *AttendanceHandler) CheckOut(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID, err := getUserIDFromContext(httpRequest.Context())
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrUnauthorized)
		return
	}

	var req dto.QRCheckRequest
	if err := response.DecodeJSON(httpRequest, &req); err != nil {
		response.SendError(responseWriter, err)
		return
	}
	if req.Payload == "" {
		response.SendError(responseWriter, apperrors.ErrInvalidRequest)
		return
	}

	payload, err := qrservice.ValidateQRToken(req.Payload)
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrQRInvalid)
		return
	}

	if payload.UserID != userID {
		response.SendError(responseWriter, apperrors.ErrForbidden)
		return
	}

	tokenHash := qrservice.HashQRToken(payload.Token)
	qrToken, err := h.attendanceStore.GetQRTokenByHash(httpRequest.Context(), tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(responseWriter, apperrors.ErrQRInvalid)
		} else {
			response.SendError(responseWriter, apperrors.ErrInternalServerError)
		}
		return
	}

	if qrToken.UsedCheckOut {
		response.SendError(responseWriter, apperrors.ErrQRAlreadyUsed)
		return
	}

	ip := extractClientIP(httpRequest)

	if err := h.attendanceStore.RecordCheckOut(httpRequest.Context(), userID, ip); err != nil {
		if err == sql.ErrNoRows {
			response.SendError(responseWriter, apperrors.ErrNoCheckInForCheckOut)
		} else {
			response.SendError(responseWriter, apperrors.ErrInternalServerError)
		}
		return
	}

	_ = h.attendanceStore.MarkQRUsedCheckOut(httpRequest.Context(), qrToken.ID)

	record, _ := h.attendanceStore.GetTodayByUserID(httpRequest.Context(), userID)
	if record != nil {
		response.SendJSON(responseWriter, http.StatusOK, dto.MapAttendanceToResponse(*record))
		return
	}
	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"message": "check-out registrado"})
}

// GetMyAttendance devuelve el historial de asistencia del empleado autenticado.
// GET /attendance/me
func (h *AttendanceHandler) GetMyAttendance(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userID, err := getUserIDFromContext(httpRequest.Context())
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrUnauthorized)
		return
	}

	records, err := h.attendanceStore.GetByUserID(httpRequest.Context(), userID)
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapAttendanceListToResponse(records))
}

// --- Private helpers ---

func getUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		return 0, apperrors.ErrUnauthorized
	}
	return userID, nil
}

func extractClientIP(httpRequest *http.Request) string {
	ip := httpRequest.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = httpRequest.RemoteAddr
	}
	return strings.Split(strings.Split(ip, ",")[0], ":")[0]
}
