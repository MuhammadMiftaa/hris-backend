-- +goose Up
-- +goose StatementBegin
INSERT INTO employees (
  id, employee_number, full_name, birth_date, birth_place, gender, religion, marital_status, nationality, is_trainer, branch_id
) VALUES (
  1, 'EMP-001', 'Muhammad Miftakul Salam', '2003-12-27', 'Surabaya', 'male', 'Islam', 'single', 'WNI', false, 1
);

INSERT INTO accounts (id, employee_id, role_id, email, password)
VALUES (
  1, 1, 1, 'mifta@wafa.id', '$2a$10$3bhH2nExhbXZtNHljOGn5OXHJ.yJiJxA0jTvKDcNbZusXgsUO/LCS'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM accounts WHERE id = 1;
DELETE FROM employees WHERE id = 1;
-- +goose StatementEnd
