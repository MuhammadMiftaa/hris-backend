-- +goose Up
-- +goose StatementBegin
ALTER TABLE leave_request_approvals ALTER COLUMN approver_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE leave_request_approvals ALTER COLUMN approver_id SET NOT NULL;
-- +goose StatementEnd
