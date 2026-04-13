-- +goose Up
-- +goose StatementBegin
CREATE TYPE gender_enum AS ENUM ('male', 'female', 'other');
CREATE TYPE marital_status_enum AS ENUM ('single', 'married', 'widowed', 'divorced');
CREATE TYPE contract_type_enum AS ENUM ('pkwt', 'pkwtt', 'probation', 'intern', 'part_time', 'freelance');
CREATE TYPE day_of_week_enum AS ENUM ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday');
CREATE TYPE leave_category_enum AS ENUM ('annual', 'sick', 'special', 'other');
CREATE TYPE duration_unit_enum AS ENUM ('days', 'hours');
CREATE TYPE permission_type_enum AS ENUM ('out_of_office', 'late_arrival', 'early_leave');
CREATE TYPE request_status_enum AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE leave_request_status_enum AS ENUM ('pending', 'approved_leader', 'approved_hr', 'rejected');
CREATE TYPE approval_status_enum AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE attendance_status_enum AS ENUM ('present', 'late', 'absent', 'half_day', 'leave', 'business_trip', 'holiday');
CREATE TYPE clock_method_enum AS ENUM ('gps', 'manual');
CREATE TYPE override_type_enum AS ENUM ('clock_in', 'clock_out', 'full_day');
CREATE TYPE work_location_type_enum AS ENUM ('office', 'home', 'outside');
CREATE TYPE holiday_type_enum AS ENUM ('national', 'joint', 'observance', 'company');
CREATE TYPE audit_action_enum AS ENUM ('create', 'update', 'delete');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS audit_action_enum;
DROP TYPE IF EXISTS holiday_type_enum;
DROP TYPE IF EXISTS work_location_type_enum;
DROP TYPE IF EXISTS override_type_enum;
DROP TYPE IF EXISTS clock_method_enum;
DROP TYPE IF EXISTS attendance_status_enum;
DROP TYPE IF EXISTS approval_status_enum;
DROP TYPE IF EXISTS leave_request_status_enum;
DROP TYPE IF EXISTS request_status_enum;
DROP TYPE IF EXISTS permission_type_enum;
DROP TYPE IF EXISTS duration_unit_enum;
DROP TYPE IF EXISTS leave_category_enum;
DROP TYPE IF EXISTS day_of_week_enum;
DROP TYPE IF EXISTS contract_type_enum;
DROP TYPE IF EXISTS marital_status_enum;
DROP TYPE IF EXISTS gender_enum;
-- +goose StatementEnd
