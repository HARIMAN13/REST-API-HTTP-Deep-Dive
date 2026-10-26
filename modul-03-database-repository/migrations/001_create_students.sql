CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    grade VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Keunikan nim tanpa membedakan huruf besar dan kecil (case-insensitive)
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_key 
    ON students (LOWER(nim));

-- Indeks tambahan pada nama untuk mempercepat pencarian (ILIKE)
CREATE INDEX IF NOT EXISTS students_name_lower_idx 
    ON students (LOWER(name));
