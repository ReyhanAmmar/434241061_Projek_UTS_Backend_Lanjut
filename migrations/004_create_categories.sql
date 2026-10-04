BEGIN;

CREATE TABLE IF NOT EXISTS categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    owner_id    INTEGER NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT categories_name_not_blank
        CHECK (CHAR_LENGTH(BTRIM(name)) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS categories_owner_name_lower_key
    ON categories (owner_id, LOWER(name));

INSERT INTO permissions (name, description) VALUES
    ('category:list', 'Melihat daftar kategori sesuai hak akses'),
    ('category:create', 'Membuat kategori milik sendiri'),
    ('category:delete', 'Mengakses fitur penghapusan kategori'),
    ('category:list:any', 'Melihat daftar kategori seluruh pengguna'),
    ('category:read:any', 'Membaca kategori milik pengguna lain'),
    ('category:update:any', 'Mengubah kategori milik pengguna lain'),
    ('category:delete:any', 'Menghapus kategori milik pengguna lain')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'category:list'),
    ('admin', 'category:create'),
    ('admin', 'category:delete'),
    ('admin', 'category:list:any'),
    ('admin', 'category:read:any'),
    ('admin', 'category:update:any'),
    ('admin', 'category:delete:any'),
    ('user', 'category:list'),
    ('user', 'category:create'),
    ('user', 'category:delete')
ON CONFLICT DO NOTHING;

COMMIT;