-- +goose Up
-- +goose StatementBegin
ALTER TABLE shift_template_details
    DROP CONSTRAINT IF EXISTS shift_template_details_shift_template_id_day_of_week_key;

CREATE UNIQUE INDEX uq_shift_template_details_active
    ON shift_template_details(shift_template_id, day_of_week)
    WHERE deleted_at IS NULL;

ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS employees_employee_number_key;

CREATE UNIQUE INDEX uq_employees_employee_number_active
    ON employees(employee_number)
    WHERE deleted_at IS NULL;

ALTER TABLE accounts
    DROP CONSTRAINT IF EXISTS accounts_email_key;

CREATE UNIQUE INDEX uq_accounts_email_active
    ON accounts(email)
    WHERE deleted_at IS NULL;

ALTER TABLE branches
    DROP CONSTRAINT IF EXISTS branches_code_key;

CREATE UNIQUE INDEX uq_branches_code_active
    ON branches(code)
    WHERE deleted_at IS NULL;

ALTER TABLE departments
    DROP CONSTRAINT IF EXISTS departments_code_key;

CREATE UNIQUE INDEX uq_departments_code_active
    ON departments(code)
    WHERE deleted_at IS NULL;

ALTER TABLE leave_balances
    DROP CONSTRAINT IF EXISTS leave_balances_employee_id_leave_type_id_year_key;

CREATE UNIQUE INDEX uq_leave_balances_active
    ON leave_balances(employee_id, leave_type_id, year)
    WHERE deleted_at IS NULL;

ALTER TABLE attendance_logs
    DROP CONSTRAINT IF EXISTS attendance_logs_employee_id_attendance_date_key;

CREATE UNIQUE INDEX uq_attendance_logs_active
    ON attendance_logs(employee_id, attendance_date)
    WHERE deleted_at IS NULL;

ALTER TABLE mutabaah_logs
    DROP CONSTRAINT IF EXISTS mutabaah_logs_employee_id_log_date_key;

CREATE UNIQUE INDEX uq_mutabaah_logs_active
    ON mutabaah_logs(employee_id, log_date)
    WHERE deleted_at IS NULL;

ALTER TABLE daily_reports
    DROP CONSTRAINT IF EXISTS daily_reports_employee_id_report_date_key;

CREATE UNIQUE INDEX uq_daily_reports_active
    ON daily_reports(employee_id, report_date)
    WHERE deleted_at IS NULL;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- SHIFT_TEMPLATE_DETAILS
DROP INDEX IF EXISTS uq_shift_template_details_active;
ALTER TABLE shift_template_details
    ADD CONSTRAINT shift_template_details_shift_template_id_day_of_week_key
    UNIQUE (shift_template_id, day_of_week);

-- EMPLOYEES
DROP INDEX IF EXISTS uq_employees_employee_number_active;
ALTER TABLE employees
    ADD CONSTRAINT employees_employee_number_key
    UNIQUE (employee_number);

-- ACCOUNTS
DROP INDEX IF EXISTS uq_accounts_email_active;
ALTER TABLE accounts
    ADD CONSTRAINT accounts_email_key
    UNIQUE (email);

-- BRANCHES
DROP INDEX IF EXISTS uq_branches_code_active;
ALTER TABLE branches
    ADD CONSTRAINT branches_code_key
    UNIQUE (code);

-- DEPARTMENTS
DROP INDEX IF EXISTS uq_departments_code_active;
ALTER TABLE departments
    ADD CONSTRAINT departments_code_key
    UNIQUE (code);

-- LEAVE_BALANCES
DROP INDEX IF EXISTS uq_leave_balances_active;
ALTER TABLE leave_balances
    ADD CONSTRAINT leave_balances_employee_id_leave_type_id_year_key
    UNIQUE (employee_id, leave_type_id, year);

-- ATTENDANCE_LOGS
DROP INDEX IF EXISTS uq_attendance_logs_active;
ALTER TABLE attendance_logs
    ADD CONSTRAINT attendance_logs_employee_id_attendance_date_key
    UNIQUE (employee_id, attendance_date);

-- MUTABAAH_LOGS
DROP INDEX IF EXISTS uq_mutabaah_logs_active;
ALTER TABLE mutabaah_logs
    ADD CONSTRAINT mutabaah_logs_employee_id_log_date_key
    UNIQUE (employee_id, log_date);

-- DAILY_REPORTS
DROP INDEX IF EXISTS uq_daily_reports_active;
ALTER TABLE daily_reports
    ADD CONSTRAINT daily_reports_employee_id_report_date_key
    UNIQUE (employee_id, report_date);

-- +goose StatementEnd