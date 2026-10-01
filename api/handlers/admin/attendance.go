package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	usermodel "coffeebase-api/internal/models/user"
	attendancestore "coffeebase-api/internal/store/attendance"
	userstore "coffeebase-api/internal/store/user"
	qrservice "coffeebase-api/internal/service/attendance"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// AdminAttendanceHandler gestiona la trazabilidad de asistencia desde el panel admin.
type AdminAttendanceHandler struct {
	attendanceStore attendancestore.Store
	userStore       userstore.Store
}

// --- Public ---

func NewAdminAttendanceHandler(attendanceStore attendancestore.Store, userStore userstore.Store) *AdminAttendanceHandler {
	return &AdminAttendanceHandler{
		attendanceStore: attendanceStore,
		userStore:       userStore,
	}
}

// GenerateQR genera un código QR (PNG base64) para un subordinado.
// El QR tiene vigencia de 1 día y es único por empleado.
// POST /admin/attendance/qr/generate
func (h *AdminAttendanceHandler) GenerateQR(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var req dto.QRGenerateRequest
	if err := response.DecodeJSON(httpRequest, &req); err != nil {
		response.SendError(responseWriter, err)
		return
	}

	if req.UserID <= 0 {
		response.SendError(responseWriter, apperrors.ErrInvalidRequest)
		return
	}

	// Verificar que el usuario objetivo existe y es un subordinado (no superadmin)
	targetUser, err := h.userStore.GetByID(httpRequest.Context(), req.UserID)
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrUserNotFound)
		return
	}

	if !isSubordinate(targetUser.Role) {
		response.SendError(responseWriter, apperrors.ErrNotASubordinate)
		return
	}

	// Generar QR PNG + token para persistir
	pngBytes, qrToken, err := qrservice.GenerateQRForUser(req.UserID)
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	// Persistir el token QR en DB
	if err := h.attendanceStore.CreateQRToken(httpRequest.Context(), qrToken); err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusCreated, dto.QRGenerateResponse{
		UserID:    req.UserID,
		QRImage:   "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes),
		ExpiresAt: qrToken.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetAll devuelve todos los registros de asistencia.
// GET /admin/attendance
func (h *AdminAttendanceHandler) GetAll(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	records, err := h.attendanceStore.GetAll(httpRequest.Context())
	if err != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}
	response.SendJSON(responseWriter, http.StatusOK, dto.MapAttendanceListToResponse(records))
}

// GetByUserID devuelve el historial de asistencia de un empleado específico.
// GET /admin/attendance/{user_id}
func (h *AdminAttendanceHandler) GetByUserID(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userIDStr := chi.URLParam(httpRequest, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	// Verificar que el usuario existe y es subordinado
	targetUser, err := h.userStore.GetByID(httpRequest.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(responseWriter, apperrors.ErrUserNotFound)
		} else {
			response.SendError(responseWriter, apperrors.ErrInternalServerError)
		}
		return
	}

	if !isSubordinate(targetUser.Role) {
		response.SendError(responseWriter, apperrors.ErrNotASubordinate)
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

// isSubordinate determina si un rol puede ser gestionado por el admin.
// Los subordinados son: barista y admin (pero NO superadmin).
func isSubordinate(role usermodel.UserRole) bool {
	return role == usermodel.RoleBarista || role == usermodel.RoleAdmin
}
