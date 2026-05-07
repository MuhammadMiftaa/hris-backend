-- +goose Up
-- +goose StatementBegin

-- 1a. Alter leave_types: hapus kolom unit jam, ubah ke DECIMAL, tambah kolom baru
ALTER TABLE leave_types
  DROP COLUMN IF EXISTS max_duration_unit,
  DROP COLUMN IF EXISTS max_total_duration_unit;

ALTER TABLE leave_types
  ALTER COLUMN max_duration_per_request    TYPE DECIMAL(5,2),
  ALTER COLUMN max_total_duration_per_year TYPE DECIMAL(5,2);

ALTER TABLE leave_types
  ADD COLUMN IF NOT EXISTS max_per_month        DECIMAL(5,2),
  ADD COLUMN IF NOT EXISTS parent_leave_type_id INTEGER REFERENCES leave_types(id),
  ADD COLUMN IF NOT EXISTS deduct_days          DECIMAL(5,2) NOT NULL DEFAULT 1.0;

-- Kolom deduct_days: berapa hari kerja yang dipotong per pengajuan tipe ini.
-- Cuti Tahunan = 1.0, Cuti Setengah Hari = 0.5
-- Keduanya merujuk ke saldo yang sama via parent_leave_type_id.

-- 1b. Alter leave_balances: ubah tipe used_duration, tambah kolom alokasi
ALTER TABLE leave_balances
  ALTER COLUMN used_duration TYPE DECIMAL(5,2);

ALTER TABLE leave_balances
  ADD COLUMN IF NOT EXISTS allocated_duration DECIMAL(5,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS effective_date     DATE         NOT NULL DEFAULT CURRENT_DATE,
  ADD COLUMN IF NOT EXISTS notes              TEXT;

-- 1c. Alter leave_requests: ubah total_days ke DECIMAL agar bisa 0.5
ALTER TABLE leave_requests
  ALTER COLUMN total_days TYPE DECIMAL(5,2);

-- 1d. Buat tabel leave_balance_adjustments (histori koreksi saldo)
CREATE TABLE IF NOT EXISTS leave_balance_adjustments (
  id               SERIAL PRIMARY KEY,
  leave_balance_id INTEGER      NOT NULL REFERENCES leave_balances(id),
  adjusted_by      INTEGER      NOT NULL REFERENCES employees(id),
  delta            DECIMAL(5,2) NOT NULL,  -- positif = tambah, negatif = kurangi
  reason           TEXT,
  created_at       TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_leave_balance_adj ON leave_balance_adjustments(leave_balance_id);

INSERT INTO permissions (code, module, action, description) VALUES
  ('leave_type-export',           'leave_type',          'export',            'Export data jenis cuti'),
  ('leave_balance-create',        'leave_balance',       'create',            'Membuat data sisa cuti'),
  ('leave_balance-update',        'leave_balance',       'update',            'Memperbarui data sisa cuti'),
  ('leave_balance-delete',        'leave_balance',       'delete',            'Menghapus data sisa cuti');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM permissions WHERE code IN (
  'leave_type-export',
  'leave_balance-create',
  'leave_balance-update',
  'leave_balance-delete'
);

DROP TABLE IF EXISTS leave_balance_adjustments CASCADE;

ALTER TABLE leave_requests  ALTER COLUMN total_days TYPE INTEGER;
ALTER TABLE leave_balances  ALTER COLUMN used_duration TYPE INTEGER;

ALTER TABLE leave_balances
  DROP COLUMN IF EXISTS notes,
  DROP COLUMN IF EXISTS effective_date,
  DROP COLUMN IF EXISTS allocated_duration;

ALTER TABLE leave_types
  DROP COLUMN IF EXISTS deduct_days,
  DROP COLUMN IF EXISTS parent_leave_type_id,
  DROP COLUMN IF EXISTS max_per_month;

ALTER TABLE leave_types
  ALTER COLUMN max_duration_per_request    TYPE INTEGER,
  ALTER COLUMN max_total_duration_per_year TYPE INTEGER;

ALTER TABLE leave_types
  ADD COLUMN IF NOT EXISTS max_duration_unit        duration_unit_enum,
  ADD COLUMN IF NOT EXISTS max_total_duration_unit  duration_unit_enum;
-- +goose StatementEnd
