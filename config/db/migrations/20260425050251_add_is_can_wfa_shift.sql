-- +goose Up
-- +goose StatementBegin
ALTER TABLE shift_templates
    ADD COLUMN IF NOT EXISTS can_wfa BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN shift_templates.can_wfa IS 'Apakah shift ini mengizinkan clock in/out di luar radius cabang (Work From Anywhere)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shift_templates
    DROP COLUMN IF EXISTS can_wfa;
-- +goose StatementEnd