package token

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken representa un refresh token almacenado en DB.
// El campo TokenHash contiene el HMAC-SHA256 del token raw — nunca el token en sí.
// El campo FamilyID agrupa todos los tokens de una misma sesión de login.
// Si un token ya rotado (Used=true) vuelve a presentarse, se invalida toda la familia (detección de robo).
type RefreshToken struct {
	ID        uuid.UUID `json:"-"`
	FamilyID  uuid.UUID `json:"-"`
	UserID    int       `json:"-"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
	Used      bool      `json:"-"`
	CreatedAt time.Time `json:"-"`
}
