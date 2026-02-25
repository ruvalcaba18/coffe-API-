-- Add language column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS language VARCHAR(10) DEFAULT 'es';

-- Comment about valid languages: español (es), ingles (en), frances (fr), aleman (de), suizo aleman (gsw)
