-- +goose Up
-- +goose StatementBegin
INSERT INTO permissions (code, module, action, description) VALUES
  ('employee-export',               'employee',                   'export', 'Export data pegawai'),
  ('template_shift-export',         'template_shift',             'export', 'Export data template shift'),
  ('schedule-export',               'schedule',                   'export', 'Export data jadwal'),
  ('holiday-export',                'holiday',                    'export', 'Export data hari libur'),
  ('attendance-export',             'attendance',                 'export', 'Export data absensi'),
  ('attendance_adjustment-export',  'attendance_adjustment',      'export', 'Export data adjustment absensi'),
  ('leave-export',                  'leave',                      'export', 'Export data cuti'),
  ('leave_balance-export',          'leave_balance',              'export', 'Export data saldo cuti'),
  ('request-export',                'request',                    'export', 'Export data permintaan'),
  ('overtime-export',               'overtime',                   'export', 'Export data lembur'),
  ('business_trip-export',          'business_trip',              'export', 'Export data perjalanan dinas'),
  ('daily_report-export',           'daily_report',               'export', 'Export data laporan harian'),
  ('mutabaah-export',               'mutabaah',                   'export', 'Export data mutabaah');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM permissions WHERE code IN (
  'employee-export',
  'template_shift-export',
  'schedule-export',
  'holiday-export',
  'attendance-export',
  'attendance_adjustment-export',
  'leave-export',
  'leave_balance-export',
  'request-export',
  'overtime-export',
  'business_trip-export',
  'daily_report-export',
  'mutabaah-export'
);
-- +goose StatementEnd
