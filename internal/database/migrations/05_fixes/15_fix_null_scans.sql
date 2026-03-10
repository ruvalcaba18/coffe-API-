
UPDATE users SET avatar_url = '' WHERE avatar_url IS NULL;
UPDATE users SET language = 'es' WHERE language IS NULL;

ALTER TABLE users ALTER COLUMN avatar_url SET DEFAULT '';
ALTER TABLE users ALTER COLUMN language SET DEFAULT 'es';
