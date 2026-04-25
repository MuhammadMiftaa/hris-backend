-- +goose Up
-- +goose StatementBegin

-- 1. Ubah status column overtime_requests menggunakan leave_request_status_enum
--    Karena PostgreSQL tidak bisa ALTER TYPE pada enum in-place,
--    kita perlu: drop default → alter column type → set default baru
ALTER TABLE overtime_requests ALTER COLUMN status DROP DEFAULT;
ALTER TABLE overtime_requests 
  ALTER COLUMN status TYPE leave_request_status_enum 
  USING (CASE status::TEXT
    WHEN 'pending' THEN 'pending'::leave_request_status_enum
    WHEN 'approved' THEN 'approved_hr'::leave_request_status_enum
    WHEN 'rejected' THEN 'rejected'::leave_request_status_enum
    ELSE 'pending'::leave_request_status_enum
  END);
ALTER TABLE overtime_requests ALTER COLUMN status SET DEFAULT 'pending'::leave_request_status_enum;

-- 2. Hapus kolom approved_by dan approver_notes dari overtime_requests (dipindah ke approval table)
ALTER TABLE overtime_requests DROP COLUMN IF EXISTS approved_by;
ALTER TABLE overtime_requests DROP COLUMN IF EXISTS approver_notes;

-- 3. Buat tabel overtime_request_approvals (mirip leave_request_approvals)
CREATE TABLE overtime_request_approvals (
  id                  SERIAL PRIMARY KEY,
  overtime_request_id INTEGER              NOT NULL REFERENCES overtime_requests(id),
  approver_id         INTEGER              REFERENCES employees(id),
  level               INTEGER              NOT NULL,  -- 1 = Leader, 2 = HR
  status              approval_status_enum NOT NULL DEFAULT 'pending',
  notes               TEXT,
  decided_at          TIMESTAMP,
  created_at          TIMESTAMP            NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS overtime_request_approvals CASCADE;

ALTER TABLE overtime_requests ALTER COLUMN status DROP DEFAULT;
ALTER TABLE overtime_requests 
  ALTER COLUMN status TYPE request_status_enum 
  USING (CASE status::TEXT
    WHEN 'pending' THEN 'pending'::request_status_enum
    WHEN 'approved_leader' THEN 'approved'::request_status_enum
    WHEN 'approved_hr' THEN 'approved'::request_status_enum
    WHEN 'rejected' THEN 'rejected'::request_status_enum
    ELSE 'pending'::request_status_enum
  END);
ALTER TABLE overtime_requests ALTER COLUMN status SET DEFAULT 'pending'::request_status_enum;

ALTER TABLE overtime_requests ADD COLUMN approved_by INTEGER REFERENCES employees(id);
ALTER TABLE overtime_requests ADD COLUMN approver_notes TEXT;
-- +goose StatementEnd
