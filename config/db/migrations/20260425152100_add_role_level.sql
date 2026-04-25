-- +goose Up
-- +goose StatementBegin
CREATE TYPE role_level_enum AS ENUM ('superadmin', 'admin', 'manager', 'staff');
ALTER TABLE roles ADD COLUMN level role_level_enum NOT NULL DEFAULT 'staff';

-- Update existing seed role
UPDATE roles SET level = 'superadmin' WHERE name = 'Super Admin' AND deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles DROP COLUMN IF EXISTS level;
DROP TYPE IF EXISTS role_level_enum;
-- +goose StatementEnd
