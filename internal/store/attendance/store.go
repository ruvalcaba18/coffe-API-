package attendance

import (
	"context"
	"database/sql"
	"time"

	attendancemodel "coffeebase-api/internal/models/attendance"

	"github.com/google/uuid"
)

// Store define operaciones de persistencia para asistencia y tokens QR.
type Store interface {
	// --- Registros de asistencia ---

	// RecordCheckIn registra la hora de entrada. Crea el registro del día si no existe.
	RecordCheckIn(ctx context.Context, userID int, ip string, tokenID uuid.UUID) error

	// RecordCheckOut registra la hora de salida en el registro del día existente.
	RecordCheckOut(ctx context.Context, userID int, ip string) error

	// GetByUserID devuelve el historial de asistencia de un empleado.
	GetByUserID(ctx context.Context, userID int) ([]attendancemodel.AttendanceRecord, error)

	// GetAll devuelve todos los registros de asistencia (para el admin).
	GetAll(ctx context.Context) ([]attendancemodel.AttendanceRecord, error)

	// GetTodayByUserID devuelve el registro de hoy de un empleado (puede ser nil).
	GetTodayByUserID(ctx context.Context, userID int) (*attendancemodel.AttendanceRecord, error)

	// --- Tokens QR ---

	// CreateQRToken persiste un nuevo token QR.
	CreateQRToken(ctx context.Context, token *attendancemodel.QRToken) error

	// GetQRTokenByHash busca un token QR por su hash HMAC.
	GetQRTokenByHash(ctx context.Context, hash string) (*attendancemodel.QRToken, error)

	// MarkQRUsedCheckIn marca el token QR como usado para check-in.
	MarkQRUsedCheckIn(ctx context.Context, id uuid.UUID) error

	// MarkQRUsedCheckOut marca el token QR como usado para check-out.
	MarkQRUsedCheckOut(ctx context.Context, id uuid.UUID) error

	// DeleteExpiredQRTokens elimina tokens QR expirados.
	DeleteExpiredQRTokens(ctx context.Context) error
}

type postgresStore struct {
	db *sql.DB
}

// --- Public ---

func NewStore(db *sql.DB) Store {
	return &postgresStore{db: db}
}

// --- Attendance records ---

func (s *postgresStore) RecordCheckIn(ctx context.Context, userID int, ip string, tokenID uuid.UUID) error {
	query := `
		INSERT INTO attendance_records (user_id, date, check_in_at, check_in_ip, qr_token_id)
		VALUES ($1, CURRENT_DATE, NOW(), $2, $3)
		ON CONFLICT (user_id, date) DO UPDATE
			SET check_in_at  = EXCLUDED.check_in_at,
			    check_in_ip  = EXCLUDED.check_in_ip,
			    qr_token_id  = EXCLUDED.qr_token_id`
	_, err := s.db.ExecContext(ctx, query, userID, ip, tokenID)
	return err
}

func (s *postgresStore) RecordCheckOut(ctx context.Context, userID int, ip string) error {
	query := `
		UPDATE attendance_records
		SET check_out_at = NOW(), check_out_ip = $2
		WHERE user_id = $1 AND date = CURRENT_DATE`
	result, err := s.db.ExecContext(ctx, query, userID, ip)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		// No hay check-in previo hoy
		return sql.ErrNoRows
	}
	return nil
}

func (s *postgresStore) GetByUserID(ctx context.Context, userID int) ([]attendancemodel.AttendanceRecord, error) {
	query := `
		SELECT id, user_id, date, check_in_at, check_out_at,
		       COALESCE(check_in_ip, ''), COALESCE(check_out_ip, ''),
		       qr_token_id,
		       CASE
		           WHEN check_in_at IS NOT NULL AND check_out_at IS NOT NULL THEN 'checked_out'
		           WHEN check_in_at IS NOT NULL THEN 'checked_in'
		           ELSE 'pending'
		       END AS status,
		       created_at
		FROM attendance_records
		WHERE user_id = $1
		ORDER BY date DESC`
	return s.scanRows(s.db.QueryContext(ctx, query, userID))
}

func (s *postgresStore) GetAll(ctx context.Context) ([]attendancemodel.AttendanceRecord, error) {
	query := `
		SELECT id, user_id, date, check_in_at, check_out_at,
		       COALESCE(check_in_ip, ''), COALESCE(check_out_ip, ''),
		       qr_token_id,
		       CASE
		           WHEN check_in_at IS NOT NULL AND check_out_at IS NOT NULL THEN 'checked_out'
		           WHEN check_in_at IS NOT NULL THEN 'checked_in'
		           ELSE 'pending'
		       END AS status,
		       created_at
		FROM attendance_records
		ORDER BY date DESC, user_id ASC`
	return s.scanRows(s.db.QueryContext(ctx, query))
}

func (s *postgresStore) GetTodayByUserID(ctx context.Context, userID int) (*attendancemodel.AttendanceRecord, error) {
	query := `
		SELECT id, user_id, date, check_in_at, check_out_at,
		       COALESCE(check_in_ip, ''), COALESCE(check_out_ip, ''),
		       qr_token_id,
		       CASE
		           WHEN check_in_at IS NOT NULL AND check_out_at IS NOT NULL THEN 'checked_out'
		           WHEN check_in_at IS NOT NULL THEN 'checked_in'
		           ELSE 'pending'
		       END AS status,
		       created_at
		FROM attendance_records
		WHERE user_id = $1 AND date = CURRENT_DATE`
	row := s.db.QueryRowContext(ctx, query, userID)

	rec, err := s.scanRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// --- QR Tokens ---

func (s *postgresStore) CreateQRToken(ctx context.Context, token *attendancemodel.QRToken) error {
	query := `
		INSERT INTO qr_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`
	return s.db.QueryRowContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.CreatedAt)
}

func (s *postgresStore) GetQRTokenByHash(ctx context.Context, hash string) (*attendancemodel.QRToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_checkin, used_checkout, created_at
		FROM qr_tokens
		WHERE token_hash = $1`
	row := s.db.QueryRowContext(ctx, query, hash)

	var t attendancemodel.QRToken
	err := row.Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt,
		&t.UsedCheckIn, &t.UsedCheckOut, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *postgresStore) MarkQRUsedCheckIn(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE qr_tokens SET used_checkin = TRUE WHERE id = $1`, id)
	return err
}

func (s *postgresStore) MarkQRUsedCheckOut(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE qr_tokens SET used_checkout = TRUE WHERE id = $1`, id)
	return err
}

func (s *postgresStore) DeleteExpiredQRTokens(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM qr_tokens WHERE expires_at < $1`, time.Now())
	return err
}

// --- Private helpers ---

func (s *postgresStore) scanRows(rows *sql.Rows, err error) ([]attendancemodel.AttendanceRecord, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []attendancemodel.AttendanceRecord
	for rows.Next() {
		rec, scanErr := s.scanRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, *rec)
	}
	if records == nil {
		records = []attendancemodel.AttendanceRecord{}
	}
	return records, rows.Err()
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (s *postgresStore) scanRow(scanner rowScanner) (*attendancemodel.AttendanceRecord, error) {
	var rec attendancemodel.AttendanceRecord
	var checkIn, checkOut sql.NullTime
	var tokenID uuid.NullUUID

	err := scanner.Scan(
		&rec.ID, &rec.UserID, &rec.Date,
		&checkIn, &checkOut,
		&rec.CheckInIP, &rec.CheckOutIP,
		&tokenID, &rec.Status, &rec.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if checkIn.Valid {
		rec.CheckInAt = &checkIn.Time
	}
	if checkOut.Valid {
		rec.CheckOutAt = &checkOut.Time
	}
	if tokenID.Valid {
		rec.QRTokenID = &tokenID.UUID
	}
	return &rec, nil
}
