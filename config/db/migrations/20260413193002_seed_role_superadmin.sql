-- +goose Up
-- +goose StatementBegin
INSERT INTO roles (id, name, description)
VALUES (1, 'Superadmin', 'Memiliki akses penuh ke seluruh fitur dan fungsi sistem HRIS di semua cabang');

INSERT INTO role_permissions (role_id, permission_code)
SELECT 1, code FROM permissions;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions WHERE role_id = 1;
DELETE FROM roles WHERE id = 1;
-- +goose StatementEnd
