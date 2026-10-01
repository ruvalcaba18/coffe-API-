package attendance

import (
	"time"

	"github.com/google/uuid"
)

// AttendanceRecord representa la asistencia de un empleado en un día.
// Solo existe un registro por usuario por fecha (UNIQUE constraint en DB).
type AttendanceRecord struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Date        time.Time  `json:"date"`
	CheckInAt   *time.Time `json:"check_in_at,omitempty"`
	CheckOutAt  *time.Time `json:"check_out_at,omitempty"`
	CheckInIP   string     `json:"check_in_ip,omitempty"`
	CheckOutIP  string     `json:"check_out_ip,omitempty"`
	QRTokenID   *uuid.UUID `json:"qr_token_id,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

// QRToken representa un token QR efímero generado por el admin.
// El TokenHash es el HMAC-SHA256 del token raw — nunca se persiste el token en sí.
type QRToken struct {
	ID           uuid.UUID `json:"-"`
	UserID       int       `json:"-"`
	TokenHash    string    `json:"-"`
	ExpiresAt    time.Time `json:"-"`
	UsedCheckIn  bool      `json:"-"`
	UsedCheckOut bool      `json:"-"`
	CreatedAt    time.Time `json:"-"`
}
