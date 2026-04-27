-- +goose Up
-- +goose StatementBegin
ALTER TABLE mutabaah_logs ALTER COLUMN attendance_log_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE mutabaah_logs SET attendance_log_id = 0 WHERE attendance_log_id IS NULL;
ALTER TABLE mutabaah_logs ALTER COLUMN attendance_log_id SET NOT NULL;
-- +goose StatementEnd
