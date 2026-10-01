-- Sistema de trazabilidad de asistencia via QR

-- Tabla principal de registros de asistencia (un registro por empleado por día)
CREATE TABLE IF NOT EXISTS attendance_records (
    id           SERIAL PRIMARY KEY,
    user_id      INT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date         DATE NOT NULL DEFAULT CURRENT_DATE,
    check_in_at  TIMESTAMP WITH TIME ZONE,
    check_out_at TIMESTAMP WITH TIME ZONE,
    check_in_ip  TEXT,
    check_out_ip TEXT,
    qr_token_id  UUID,                                   -- UUID del token QR usado para auditoría
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)                                -- un único registro por empleado por día
);

CREATE INDEX IF NOT EXISTS idx_attendance_user_id ON attendance_records(user_id);
CREATE INDEX IF NOT EXISTS idx_attendance_date    ON attendance_records(date);

-- Tokens QR efímeros (separados del JWT)
CREATE TABLE IF NOT EXISTS qr_tokens (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       INT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash    TEXT NOT NULL UNIQUE,                  -- HMAC-SHA256 del token
    expires_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    used_checkin  BOOLEAN NOT NULL DEFAULT FALSE,
    used_checkout BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_qr_tokens_user   ON qr_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_qr_tokens_hash   ON qr_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_qr_tokens_expiry ON qr_tokens(expires_at);
