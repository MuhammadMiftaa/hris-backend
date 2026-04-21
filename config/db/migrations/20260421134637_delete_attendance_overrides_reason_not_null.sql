-- +goose Up
-- +goose StatementBegin
ALTER TABLE attendance_overrides ALTER COLUMN reason DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE TABLE attendance_overrides SET reason = ' ' WHERE reason = '' OR reason IS NULL;
ALTER TABLE attendance_overrides ALTER COLUMN reason SET NOT NULL;
-- +goose StatementEnd
