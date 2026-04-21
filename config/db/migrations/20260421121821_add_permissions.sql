-- +goose Up
-- +goose StatementBegin
INSERT INTO permissions (code, module, action, description) VALUES
  ('employee-delete',                 'employee',               'delete',            'Menghapus data pegawai'),
  ('attendance-update',               'attendance',             'update',            'Memperbarui data jadwal shift'),
  ('leave-update',                    'leave',                  'update',            'Memperbarui data jenis cuti'),
  ('request-update',                  'request',                'update',            'Memperbarui data pengajuan'),
  ('request-delete',                  'request',                'delete',            'Menghapus data pengajuan'),
  ('daily_report-delete',             'daily_report',           'delete',            'Menghapus data laporan harian'),
  ('mutabaah-update',                 'mutabaah',               'update',            'Memperbarui data mutabaah');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM permissions WHERE code IN (
  'employee-delete',
  'attendance-update',
  'leave-update',
  'request-update',
  'request-delete',
  'daily_report-delete',
  'mutabaah-update'
);
-- +goose StatementEnd
