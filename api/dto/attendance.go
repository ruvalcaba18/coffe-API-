package dto

import (
	attendancemodel "coffeebase-api/internal/models/attendance"
	"time"
)

// QRGenerateRequest es el body para que el admin genere un QR para un subordinado.
type QRGenerateRequest struct {
	UserID int `json:"user_id"`
}

// QRCheckRequest es el payload JSON escaneado del QR (enviado por el empleado al hacer check-in/out).
type QRCheckRequest struct {
	Payload string `json:"payload"` // JSON completo leído del QR
}

// AttendanceResponse es la respuesta de un registro de asistencia.
type AttendanceResponse struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Date       string     `json:"date"`
	CheckInAt  *time.Time `json:"check_in_at,omitempty"`
	CheckOutAt *time.Time `json:"check_out_at,omitempty"`
	Status     string     `json:"status"` // "pending", "checked_in", "checked_out"
}

// QRGenerateResponse es la respuesta cuando el admin genera un QR.
type QRGenerateResponse struct {
	UserID    int    `json:"user_id"`
	QRImage   string `json:"qr_image"`   // PNG en base64
	ExpiresAt string `json:"expires_at"` // ISO 8601
}

// --- Mappers ---

func MapAttendanceToResponse(rec attendancemodel.AttendanceRecord) AttendanceResponse {
	return AttendanceResponse{
		ID:         rec.ID,
		UserID:     rec.UserID,
		Date:       rec.Date.Format("2006-01-02"),
		CheckInAt:  rec.CheckInAt,
		CheckOutAt: rec.CheckOutAt,
		Status:     rec.Status,
	}
}

func MapAttendanceListToResponse(records []attendancemodel.AttendanceRecord) []AttendanceResponse {
	dtos := make([]AttendanceResponse, 0, len(records))
	for _, rec := range records {
		dtos = append(dtos, MapAttendanceToResponse(rec))
	}
	return dtos
}
