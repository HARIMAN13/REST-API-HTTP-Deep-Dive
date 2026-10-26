-- ---------------------------------------------------------------
-- permissions untuk students
-- ---------------------------------------------------------------
INSERT INTO permissions (name, description) VALUES
    ('student:list', 'Melihat daftar seluruh student'),
    ('student:read:any', 'Melihat data student mana pun'),
    ('student:create', 'Mendaftar student baru'),
    ('student:update:any', 'Mengubah data student mana pun'),
    ('student:delete', 'Menghapus student')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------
-- Menambahkan kolom owner_id pada tabel students
-- ---------------------------------------------------------------
ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER;

-- Hapus students yang belum memiliki owner_id sebelum memasang constraint
DELETE FROM students WHERE owner_id IS NULL;

-- Pasang foreign key dan set not null
ALTER TABLE students DROP CONSTRAINT IF EXISTS students_owner_id_fkey;
ALTER TABLE students
    ADD CONSTRAINT students_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE students ALTER COLUMN owner_id SET NOT NULL;
