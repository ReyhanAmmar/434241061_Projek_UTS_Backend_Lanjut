ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS type VARCHAR(10) NOT NULL DEFAULT 'expense';

ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_type_check;
ALTER TABLE categories
    ADD CONSTRAINT categories_type_check CHECK (type IN ('income', 'expense'));

CREATE TABLE IF NOT EXISTS default_categories (
    type        VARCHAR(10)  NOT NULL
        CHECK (type IN ('income', 'expense')),
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    PRIMARY KEY (type, name)
);

INSERT INTO default_categories (type, name, description) VALUES
    ('income',  'Gaji',          'Gaji bulanan'),
    ('income',  'Bonus',         'Bonus atau tunjangan'),
    ('income',  'Hadiah',        'Uang hadiah atau pemberian'),
    ('income',  'Lainnya',       'Pemasukan lain-lain'),
    ('expense', 'Makan',         'Makan dan minum'),
    ('expense', 'Belanja',       'Belanja kebutuhan'),
    ('expense', 'Bensin',        'Isi bensin kendaraan'),
    ('expense', 'Transportasi',  'Ongkos transportasi'),
    ('expense', 'Tagihan',       'Listrik, air, internet, pulsa'),
    ('expense', 'Hiburan',       'Hiburan dan rekreasi'),
    ('expense', 'Lainnya',       'Pengeluaran lain-lain')
ON CONFLICT DO NOTHING;

ALTER TABLE categories ALTER COLUMN type DROP DEFAULT;

DROP INDEX IF EXISTS categories_owner_name_lower_key;

CREATE UNIQUE INDEX IF NOT EXISTS categories_owner_type_name_lower_key
    ON categories (owner_id, type, LOWER(name));

INSERT INTO categories (name, description, type, owner_id)
SELECT d.name, d.description, d.type, u.id
FROM users u
CROSS JOIN default_categories d
ON CONFLICT DO NOTHING;

ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_id_type_key;
ALTER TABLE categories
    ADD CONSTRAINT categories_id_type_key UNIQUE (id, type);

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_category_type_fk;
ALTER TABLE transactions
    ADD CONSTRAINT transactions_category_type_fk
    FOREIGN KEY (category_id, type)
    REFERENCES categories (id, type)
    ON DELETE RESTRICT;
