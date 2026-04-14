-- +goose Up
-- +goose StatementBegin
-- -----------------------------------------------------------------------------
-- 0. BRANCHES
-- -----------------------------------------------------------------------------
CREATE TABLE branches (
  id              SERIAL PRIMARY KEY,
  code            VARCHAR(20)    NOT NULL UNIQUE,
  name            VARCHAR(100)   NOT NULL,
  address         TEXT,
  latitude        DECIMAL(10,8),
  longitude       DECIMAL(11,8),
  radius_meters   INTEGER        NOT NULL DEFAULT 100,
  allow_wfh       BOOLEAN        NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMP      NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMP,
  deleted_at      TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 1. ROLES
-- -----------------------------------------------------------------------------
CREATE TABLE roles (
  id          SERIAL PRIMARY KEY,
  name        VARCHAR(100) NOT NULL,
  description TEXT,
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP,
  deleted_at  TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 2. PERMISSIONS
-- -----------------------------------------------------------------------------
CREATE TABLE permissions (
  code        VARCHAR(100) PRIMARY KEY,
  module      VARCHAR(100) NOT NULL,
  action      VARCHAR(100) NOT NULL,
  description TEXT,
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP,
  deleted_at  TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 3. DEPARTMENTS
-- -----------------------------------------------------------------------------
CREATE TABLE departments (
  id          SERIAL PRIMARY KEY,
  code        VARCHAR(20)  NOT NULL UNIQUE,
  name        VARCHAR(100) NOT NULL,
  branch_id   INTEGER      REFERENCES branches(id),  -- NULL = berlaku semua cabang
  description TEXT,
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP,
  deleted_at  TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 4. JOB POSITIONS
-- -----------------------------------------------------------------------------
CREATE TABLE job_positions (
  id            SERIAL PRIMARY KEY,
  title         VARCHAR(100),
  department_id INTEGER NOT NULL REFERENCES departments(id),
  created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP,
  deleted_at    TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 5. EMPLOYEES
-- -----------------------------------------------------------------------------
CREATE TABLE employees (
  id               SERIAL PRIMARY KEY,
  employee_number  VARCHAR(20)          NOT NULL UNIQUE,
  full_name        VARCHAR(150)         NOT NULL,
  nik              VARCHAR(16),
  npwp             VARCHAR(20),
  kk_number        VARCHAR(16),
  birth_date       DATE                 NOT NULL,
  birth_place      VARCHAR(100),
  gender           gender_enum,
  religion         VARCHAR(50),
  marital_status   marital_status_enum,
  blood_type       VARCHAR(5),
  nationality      VARCHAR(50),
  height           NUMERIC(5,2),
  weight           NUMERIC(5,2),
  photo_url        TEXT,
  is_trainer       BOOLEAN              NOT NULL DEFAULT FALSE,
  branch_id        INTEGER              REFERENCES branches(id),
  department_id    INTEGER              REFERENCES departments(id),
  job_positions_id INTEGER              REFERENCES job_positions(id),
  created_at       TIMESTAMP            NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP,
  deleted_at       TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 6. ACCOUNTS
-- -----------------------------------------------------------------------------
CREATE TABLE accounts (
  id              SERIAL PRIMARY KEY,
  employee_id     INTEGER               NOT NULL REFERENCES employees(id),
  role_id         INTEGER               NOT NULL REFERENCES roles(id),
  email           VARCHAR(150)          NOT NULL UNIQUE,
  password        TEXT                  NOT NULL,
  last_login_at   TIMESTAMP,
  is_active       BOOLEAN               NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMP             NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMP,
  deleted_at      TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 7. ROLE PERMISSIONS
-- -----------------------------------------------------------------------------
CREATE TABLE role_permissions (
  id              SERIAL PRIMARY KEY,
  role_id         INTEGER NOT NULL REFERENCES roles(id),
  permission_code VARCHAR(100) NOT NULL REFERENCES permissions(code),
  created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (role_id, permission_code)
);

-- -----------------------------------------------------------------------------
-- 8. EMPLOYEE CONTACTS
-- -----------------------------------------------------------------------------
CREATE TABLE employee_contacts (
  id           SERIAL PRIMARY KEY,
  employee_id  INTEGER      NOT NULL REFERENCES employees(id),
  phone        VARCHAR(20),
  email        VARCHAR(150),
  address_line TEXT,
  city         VARCHAR(50),
  province     VARCHAR(50),
  postal_code  VARCHAR(10),
  is_primary   BOOLEAN      NOT NULL DEFAULT FALSE,
  created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMP,
  deleted_at   TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 9. EMPLOYMENT CONTRACTS
-- -----------------------------------------------------------------------------
CREATE TABLE employment_contracts (
  id            SERIAL PRIMARY KEY,
  employee_id   INTEGER            NOT NULL REFERENCES employees(id),
  contract_type contract_type_enum NOT NULL,
  start_date    DATE,
  end_date      DATE,
  salary        NUMERIC(12,2),
  created_at    TIMESTAMP          NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP,
  deleted_at    TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 10. SHIFT TEMPLATES
-- -----------------------------------------------------------------------------
CREATE TABLE shift_templates (
  id          SERIAL PRIMARY KEY,
  name        VARCHAR(100) NOT NULL,
  is_flexible BOOLEAN      NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP,
  deleted_at  TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 11. SHIFT TEMPLATE DETAILS
-- -----------------------------------------------------------------------------
CREATE TABLE shift_template_details (
  id                SERIAL PRIMARY KEY,
  shift_template_id INTEGER          NOT NULL REFERENCES shift_templates(id),
  day_of_week       day_of_week_enum NOT NULL,
  is_working_day    BOOLEAN          NOT NULL DEFAULT TRUE,
  clock_in_start    TIME,
  clock_in_end      TIME,
  break_dhuhr_start TIME,
  break_dhuhr_end   TIME,
  break_asr_start   TIME,
  break_asr_end     TIME,
  clock_out_start   TIME,
  clock_out_end     TIME,
  created_at        TIMESTAMP        NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP,
  deleted_at        TIMESTAMP,
  UNIQUE (shift_template_id, day_of_week)
);

-- -----------------------------------------------------------------------------
-- 12. EMPLOYEE SCHEDULES
-- -----------------------------------------------------------------------------
CREATE TABLE employee_schedules (
  id                SERIAL PRIMARY KEY,
  employee_id       INTEGER   NOT NULL REFERENCES employees(id),
  shift_template_id INTEGER   NOT NULL REFERENCES shift_templates(id),
  effective_date    DATE      NOT NULL,
  end_date          DATE,  -- NULL = berlaku seterusnya
  is_active         BOOLEAN   NOT NULL DEFAULT TRUE,
  created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP,
  deleted_at        TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 13. LEAVE TYPES
-- -----------------------------------------------------------------------------
CREATE TABLE leave_types (
  id                          SERIAL PRIMARY KEY,
  name                        VARCHAR(100)       NOT NULL,
  category                    leave_category_enum NOT NULL,
  requires_document           BOOLEAN            NOT NULL DEFAULT FALSE,
  requires_document_type      VARCHAR(100),
  max_duration_per_request    INTEGER,
  max_duration_unit           duration_unit_enum,
  max_occurrences_per_year    INTEGER,
  max_total_duration_per_year INTEGER,
  max_total_duration_unit     duration_unit_enum,
  created_at                  TIMESTAMP          NOT NULL DEFAULT NOW(),
  updated_at                  TIMESTAMP,
  deleted_at                  TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 14. LEAVE BALANCES
-- -----------------------------------------------------------------------------
CREATE TABLE leave_balances (
  id               SERIAL PRIMARY KEY,
  employee_id      INTEGER   NOT NULL REFERENCES employees(id),
  leave_type_id    INTEGER   NOT NULL REFERENCES leave_types(id),
  year             INTEGER   NOT NULL,
  used_occurrences INTEGER   NOT NULL DEFAULT 0,
  used_duration    INTEGER   NOT NULL DEFAULT 0,
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP,
  deleted_at       TIMESTAMP,
  UNIQUE (employee_id, leave_type_id, year)
);

-- -----------------------------------------------------------------------------
-- 15. PERMISSION REQUESTS
-- Dibuat sebelum attendance_logs karena attendance_logs mereferensikan tabel ini
-- -----------------------------------------------------------------------------
CREATE TABLE permission_requests (
  id              SERIAL PRIMARY KEY,
  employee_id     INTEGER              NOT NULL REFERENCES employees(id),
  permission_type permission_type_enum NOT NULL,
  date            DATE                 NOT NULL,
  leave_time      TIME,
  return_time     TIME,
  reason          TEXT                 NOT NULL,
  document_url    TEXT,
  status          request_status_enum  NOT NULL DEFAULT 'pending',
  approved_by     INTEGER              REFERENCES employees(id),
  approver_notes  TEXT,
  created_at      TIMESTAMP            NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMP,
  deleted_at      TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 16. LEAVE REQUESTS
-- Dibuat sebelum attendance_logs karena attendance_logs mereferensikan tabel ini
-- -----------------------------------------------------------------------------
CREATE TABLE leave_requests (
  id            SERIAL PRIMARY KEY,
  employee_id   INTEGER                   NOT NULL REFERENCES employees(id),
  leave_type_id INTEGER                   NOT NULL REFERENCES leave_types(id),
  start_date    DATE                      NOT NULL,
  end_date      DATE                      NOT NULL,
  total_days    INTEGER                   NOT NULL,
  total_hours   INTEGER,
  reason        TEXT,
  document_url  TEXT,
  status        leave_request_status_enum NOT NULL DEFAULT 'pending',
  created_at    TIMESTAMP                 NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP,
  deleted_at    TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 17. LEAVE REQUEST APPROVALS
-- -----------------------------------------------------------------------------
CREATE TABLE leave_request_approvals (
  id               SERIAL PRIMARY KEY,
  leave_request_id INTEGER              NOT NULL REFERENCES leave_requests(id),
  approver_id      INTEGER              NOT NULL REFERENCES employees(id),
  level            INTEGER              NOT NULL,  -- 1 = Leader Dept, 2 = Leader HRGA
  status           approval_status_enum NOT NULL DEFAULT 'pending',
  notes            TEXT,
  decided_at       TIMESTAMP,
  created_at       TIMESTAMP            NOT NULL DEFAULT NOW()
);

-- -----------------------------------------------------------------------------
-- 18. BUSINESS TRIP REQUESTS
-- Dibuat sebelum attendance_logs karena attendance_logs mereferensikan tabel ini
-- -----------------------------------------------------------------------------
CREATE TABLE business_trip_requests (
  id             SERIAL PRIMARY KEY,
  employee_id    INTEGER             NOT NULL REFERENCES employees(id),
  destination    VARCHAR(255)        NOT NULL,
  start_date     DATE                NOT NULL,
  end_date       DATE                NOT NULL,
  total_days     INTEGER             NOT NULL,
  purpose        TEXT                NOT NULL,
  document_url   TEXT,
  status         request_status_enum NOT NULL DEFAULT 'pending',
  approved_by    INTEGER             REFERENCES employees(id),
  approver_notes TEXT,
  created_at     TIMESTAMP           NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMP,
  deleted_at     TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 19. ATTENDANCE LOGS
-- Dibuat setelah permission_requests, leave_requests, business_trip_requests
-- -----------------------------------------------------------------------------
CREATE TABLE attendance_logs (
  id                       SERIAL PRIMARY KEY,
  employee_id              INTEGER                NOT NULL REFERENCES employees(id),
  schedule_id              INTEGER                REFERENCES employee_schedules(id),
  attendance_date          DATE                   NOT NULL,
  clock_in_at              TIMESTAMP,
  clock_out_at             TIMESTAMP,
  clock_in_lat             DECIMAL(10,8),
  clock_in_lng             DECIMAL(11,8),
  clock_out_lat            DECIMAL(10,8),
  clock_out_lng            DECIMAL(11,8),
  clock_in_photo_url       TEXT,
  clock_out_photo_url      TEXT,
  clock_in_method          clock_method_enum,
  clock_out_method         clock_method_enum,
  status                   attendance_status_enum NOT NULL DEFAULT 'absent',
  permission_request_id    INTEGER                REFERENCES permission_requests(id),
  leave_request_id         INTEGER                REFERENCES leave_requests(id),
  business_trip_request_id INTEGER                REFERENCES business_trip_requests(id),
  is_counted_as_full_day   BOOLEAN                NOT NULL DEFAULT FALSE,
  late_minutes             INTEGER                NOT NULL DEFAULT 0,
  early_leave_minutes      INTEGER                NOT NULL DEFAULT 0,
  late_notes               TEXT,
  early_leave_notes        TEXT,
  late_document_url        TEXT,
  overtime_minutes         INTEGER                NOT NULL DEFAULT 0,
  is_auto_generated        BOOLEAN                NOT NULL DEFAULT FALSE,
  created_at               TIMESTAMP              NOT NULL DEFAULT NOW(),
  updated_at               TIMESTAMP,
  deleted_at               TIMESTAMP,
  UNIQUE (employee_id, attendance_date)
);

-- -----------------------------------------------------------------------------
-- 20. MUTABAAH LOGS
-- -----------------------------------------------------------------------------
CREATE TABLE mutabaah_logs (
  id                SERIAL PRIMARY KEY,
  employee_id       INTEGER   NOT NULL REFERENCES employees(id),
  attendance_log_id INTEGER   NOT NULL REFERENCES attendance_logs(id),
  log_date          DATE      NOT NULL,
  target_pages      INTEGER   NOT NULL,
  is_submitted      BOOLEAN   NOT NULL DEFAULT FALSE,
  submitted_at      TIMESTAMP,
  is_auto_generated BOOLEAN   NOT NULL DEFAULT FALSE,
  created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP,
  deleted_at        TIMESTAMP,
  UNIQUE (employee_id, log_date)
);

-- -----------------------------------------------------------------------------
-- 21. DAILY REPORTS
-- -----------------------------------------------------------------------------
CREATE TABLE daily_reports (
  id                SERIAL PRIMARY KEY,
  employee_id       INTEGER   NOT NULL REFERENCES employees(id),
  attendance_log_id INTEGER   NOT NULL REFERENCES attendance_logs(id),
  report_date       DATE      NOT NULL,
  activities        TEXT,
  is_submitted      BOOLEAN   NOT NULL DEFAULT FALSE,
  submitted_at      TIMESTAMP,
  is_auto_generated BOOLEAN   NOT NULL DEFAULT FALSE,
  created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP,
  deleted_at        TIMESTAMP,
  UNIQUE (employee_id, report_date)
);

-- -----------------------------------------------------------------------------
-- 22. ATTENDANCE OVERRIDES
-- -----------------------------------------------------------------------------
CREATE TABLE attendance_overrides (
  id                 SERIAL PRIMARY KEY,
  attendance_log_id  INTEGER             NOT NULL REFERENCES attendance_logs(id),
  requested_by       INTEGER             NOT NULL REFERENCES employees(id),
  approved_by        INTEGER             REFERENCES employees(id),
  override_type      override_type_enum  NOT NULL,
  original_clock_in  TIMESTAMP,
  original_clock_out TIMESTAMP,
  corrected_clock_in TIMESTAMP,
  corrected_clock_out TIMESTAMP,
  reason             TEXT                NOT NULL,
  status             request_status_enum NOT NULL DEFAULT 'pending',
  created_at         TIMESTAMP           NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP,
  deleted_at         TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 23. OVERTIME REQUESTS
-- attendance_log_id nullable karena lembur terencana diajukan sebelum hari H
-- -----------------------------------------------------------------------------
CREATE TABLE overtime_requests (
  id                 SERIAL PRIMARY KEY,
  employee_id        INTEGER                 NOT NULL REFERENCES employees(id),
  attendance_log_id  INTEGER                 REFERENCES attendance_logs(id),  -- nullable
  overtime_date      DATE                    NOT NULL,
  planned_start      TIMESTAMP,
  planned_end        TIMESTAMP,
  actual_start       TIMESTAMP,
  actual_end         TIMESTAMP,
  planned_minutes    INTEGER                 NOT NULL,
  actual_minutes     INTEGER,
  reason             TEXT                    NOT NULL,
  work_location_type work_location_type_enum,
  status             request_status_enum     NOT NULL DEFAULT 'pending',
  approved_by        INTEGER                 REFERENCES employees(id),
  approver_notes     TEXT,
  created_at         TIMESTAMP               NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP,
  deleted_at         TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 24. HOLIDAYS
-- -----------------------------------------------------------------------------
CREATE TABLE holidays (
  id          SERIAL PRIMARY KEY,
  name        VARCHAR(100)      NOT NULL,
  year        INTEGER           NOT NULL,
  date        DATE              NOT NULL,
  type        holiday_type_enum NOT NULL,
  branch_id   INTEGER           REFERENCES branches(id),  -- NULL = semua cabang
  description TEXT,
  created_at  TIMESTAMP         NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP,
  deleted_at  TIMESTAMP
);

-- Partial unique index untuk libur nasional (branch_id IS NULL):
-- mencegah duplikasi tanggal libur nasional yang sama
CREATE UNIQUE INDEX idx_holidays_national_unique
  ON holidays(date)
  WHERE branch_id IS NULL AND deleted_at IS NULL;

-- Partial unique index untuk libur perusahaan (branch_id IS NOT NULL):
-- mencegah duplikasi tanggal libur di cabang yang sama
CREATE UNIQUE INDEX idx_holidays_company_unique
  ON holidays(date, branch_id)
  WHERE branch_id IS NOT NULL AND deleted_at IS NULL;

-- -----------------------------------------------------------------------------
-- 25. AUDIT LOGS
-- employee_id nullable untuk aksi yang dilakukan sistem otomatis
-- -----------------------------------------------------------------------------
CREATE TABLE audit_logs (
  id          SERIAL PRIMARY KEY,
  employee_id INTEGER           REFERENCES employees(id),  -- NULL = system action
  table_name  VARCHAR(100)      NOT NULL,
  record_id   INTEGER           NOT NULL,
  action      audit_action_enum NOT NULL,
  old_values  JSONB,
  new_values  JSONB,
  ip_address  VARCHAR(45),
  user_agent  TEXT,
  created_at  TIMESTAMP         NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS holidays CASCADE;
DROP TABLE IF EXISTS overtime_requests CASCADE;
DROP TABLE IF EXISTS attendance_overrides CASCADE;
DROP TABLE IF EXISTS daily_reports CASCADE;
DROP TABLE IF EXISTS mutabaah_logs CASCADE;
DROP TABLE IF EXISTS attendance_logs CASCADE;
DROP TABLE IF EXISTS business_trip_requests CASCADE;
DROP TABLE IF EXISTS leave_request_approvals CASCADE;
DROP TABLE IF EXISTS leave_requests CASCADE;
DROP TABLE IF EXISTS permission_requests CASCADE;
DROP TABLE IF EXISTS leave_balances CASCADE;
DROP TABLE IF EXISTS leave_types CASCADE;
DROP TABLE IF EXISTS employee_schedules CASCADE;
DROP TABLE IF EXISTS shift_template_details CASCADE;
DROP TABLE IF EXISTS shift_templates CASCADE;
DROP TABLE IF EXISTS employment_contracts CASCADE;
DROP TABLE IF EXISTS employee_contacts CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS employees CASCADE;
DROP TABLE IF EXISTS job_positions CASCADE;
DROP TABLE IF EXISTS departments CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS branches CASCADE;
-- +goose StatementEnd
