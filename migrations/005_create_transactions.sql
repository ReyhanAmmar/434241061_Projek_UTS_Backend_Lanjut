CREATE TABLE IF NOT EXISTS transactions (
    id               SERIAL PRIMARY KEY,
    type             VARCHAR(10)  NOT NULL,
    amount           BIGINT       NOT NULL,
    note             VARCHAR(255) NOT NULL DEFAULT '',
    transaction_date TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    category_id      INTEGER      NOT NULL
        REFERENCES categories(id) ON DELETE RESTRICT,
    owner_id         INTEGER      NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT transactions_type_check
        CHECK (type IN ('income', 'expense')),
    CONSTRAINT transactions_amount_positive
        CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS transactions_owner_date_idx
    ON transactions (owner_id, transaction_date DESC, id DESC);

CREATE INDEX IF NOT EXISTS transactions_category_idx
    ON transactions (category_id);

INSERT INTO permissions (name, description) VALUES
    ('transaction:list',       'Melihat daftar transaksi sesuai hak akses'),
    ('transaction:create',     'Membuat transaksi milik sendiri'),
    ('transaction:delete',     'Mengakses fitur penghapusan transaksi'),
    ('transaction:list:any',   'Melihat daftar transaksi seluruh pengguna'),
    ('transaction:read:any',   'Membaca transaksi milik pengguna lain'),
    ('transaction:update:any', 'Mengubah transaksi milik pengguna lain'),
    ('transaction:delete:any', 'Menghapus transaksi milik pengguna lain')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'transaction:list'),
    ('admin', 'transaction:create'),
    ('admin', 'transaction:delete'),
    ('admin', 'transaction:list:any'),
    ('admin', 'transaction:read:any'),
    ('admin', 'transaction:update:any'),
    ('admin', 'transaction:delete:any'),
    ('user',  'transaction:list'),
    ('user',  'transaction:create'),
    ('user',  'transaction:delete')
ON CONFLICT DO NOTHING;