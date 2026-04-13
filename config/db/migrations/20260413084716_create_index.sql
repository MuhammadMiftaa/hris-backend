-- +goose Up
-- +goose StatementBegin
-- employees
CREATE INDEX idx_employees_branch_id      ON employees(branch_id);
CREATE INDEX idx_employees_department_id  ON employees(department_id);
CREATE INDEX idx_employees_is_trainer     ON employees(is_trainer) WHERE deleted_at IS NULL;

-- employee_schedules
CREATE INDEX idx_schedules_employee_id    ON employee_schedules(employee_id);
CREATE INDEX idx_schedules_effective_date ON employee_schedules(effective_date);

-- attendance_logs
CREATE INDEX idx_attendance_employee_date ON attendance_logs(employee_id, attendance_date);
CREATE INDEX idx_attendance_date          ON attendance_logs(attendance_date);
CREATE INDEX idx_attendance_status        ON attendance_logs(status) WHERE deleted_at IS NULL;

-- mutabaah_logs
CREATE INDEX idx_mutabaah_employee_date   ON mutabaah_logs(employee_id, log_date);
CREATE INDEX idx_mutabaah_log_date        ON mutabaah_logs(log_date);
CREATE INDEX idx_mutabaah_is_submitted    ON mutabaah_logs(is_submitted) WHERE deleted_at IS NULL;

-- daily_reports
CREATE INDEX idx_daily_report_employee    ON daily_reports(employee_id, report_date);
CREATE INDEX idx_daily_report_date        ON daily_reports(report_date);
CREATE INDEX idx_daily_report_submitted   ON daily_reports(is_submitted) WHERE deleted_at IS NULL;

-- leave_requests
CREATE INDEX idx_leave_requests_employee  ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status    ON leave_requests(status) WHERE deleted_at IS NULL;

-- permission_requests
CREATE INDEX idx_permission_req_employee  ON permission_requests(employee_id);
CREATE INDEX idx_permission_req_date      ON permission_requests(date);
CREATE INDEX idx_permission_req_status    ON permission_requests(status) WHERE deleted_at IS NULL;

-- overtime_requests
CREATE INDEX idx_overtime_employee        ON overtime_requests(employee_id);
CREATE INDEX idx_overtime_status          ON overtime_requests(status) WHERE deleted_at IS NULL;

-- business_trip_requests
CREATE INDEX idx_business_trip_employee   ON business_trip_requests(employee_id);
CREATE INDEX idx_business_trip_status     ON business_trip_requests(status) WHERE deleted_at IS NULL;

-- attendance_overrides
CREATE INDEX idx_overrides_log_id         ON attendance_overrides(attendance_log_id);
CREATE INDEX idx_overrides_status         ON attendance_overrides(status) WHERE deleted_at IS NULL;

-- leave_balances
CREATE INDEX idx_leave_balances_employee  ON leave_balances(employee_id, year);

-- holidays (partial unique indexes sudah dibuat di atas bersama tabel)
CREATE INDEX idx_holidays_date            ON holidays(date, year);
CREATE INDEX idx_holidays_branch          ON holidays(branch_id);

-- audit_logs
CREATE INDEX idx_audit_logs_table         ON audit_logs(table_name, record_id);
CREATE INDEX idx_audit_logs_employee      ON audit_logs(employee_id);
CREATE INDEX idx_audit_logs_created_at    ON audit_logs(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_employee;
DROP INDEX IF EXISTS idx_audit_logs_table;
DROP INDEX IF EXISTS idx_holidays_branch;
DROP INDEX IF EXISTS idx_holidays_date;
DROP INDEX IF EXISTS idx_leave_balances_employee;
DROP INDEX IF EXISTS idx_overrides_status;
DROP INDEX IF EXISTS idx_overrides_log_id;
DROP INDEX IF EXISTS idx_business_trip_status;
DROP INDEX IF EXISTS idx_business_trip_employee;
DROP INDEX IF EXISTS idx_overtime_status;
DROP INDEX IF EXISTS idx_overtime_employee;
DROP INDEX IF EXISTS idx_permission_req_status;
DROP INDEX IF EXISTS idx_permission_req_date;
DROP INDEX IF EXISTS idx_permission_req_employee;
DROP INDEX IF EXISTS idx_leave_requests_status;
DROP INDEX IF EXISTS idx_leave_requests_employee;
DROP INDEX IF EXISTS idx_daily_report_submitted;
DROP INDEX IF EXISTS idx_daily_report_date;
DROP INDEX IF EXISTS idx_daily_report_employee;
DROP INDEX IF EXISTS idx_mutabaah_is_submitted;
DROP INDEX IF EXISTS idx_mutabaah_log_date;
DROP INDEX IF EXISTS idx_mutabaah_employee_date;
DROP INDEX IF EXISTS idx_attendance_status;
DROP INDEX IF EXISTS idx_attendance_date;
DROP INDEX IF EXISTS idx_attendance_employee_date;
DROP INDEX IF EXISTS idx_schedules_effective_date;
DROP INDEX IF EXISTS idx_schedules_employee_id;
DROP INDEX IF EXISTS idx_employees_is_trainer;

DROP INDEX IF EXISTS idx_employees_department_id;
DROP INDEX IF EXISTS idx_employees_branch_id;
-- +goose StatementEnd
