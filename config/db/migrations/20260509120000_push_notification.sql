-- +goose Up
-- +goose StatementBegin

-- =============================================================================
-- PUSH NOTIFICATION TABLES
-- Issue Integration 14
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. PUSH SUBSCRIPTIONS
-- ---------------------------------------------------------------------------
CREATE TABLE push_subscriptions (
    id          SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees(id),
    endpoint    TEXT NOT NULL,
    p256dh      TEXT NOT NULL,
    auth        TEXT NOT NULL,
    user_agent  TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP,
    deleted_at  TIMESTAMP
);

CREATE INDEX idx_push_subs_employee_active
    ON push_subscriptions(employee_id)
    WHERE deleted_at IS NULL AND is_active = TRUE;

-- ---------------------------------------------------------------------------
-- 2. NOTIFICATIONS
-- ---------------------------------------------------------------------------
CREATE TABLE notifications (
    id                  SERIAL PRIMARY KEY,
    employee_id         INTEGER NOT NULL REFERENCES employees(id),
    type                VARCHAR(50) NOT NULL,
    title               VARCHAR(255) NOT NULL,
    body                TEXT NOT NULL,
    action_url          TEXT,
    action_tab          VARCHAR(50),
    is_read             BOOLEAN NOT NULL DEFAULT FALSE,
    read_at             TIMESTAMP,
    push_status         VARCHAR(20) NOT NULL DEFAULT 'pending',
    push_attempts       INTEGER NOT NULL DEFAULT 0,
    last_attempt_at     TIMESTAMP,
    related_entity_type VARCHAR(50),
    related_entity_id   INTEGER,
    send_at             TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP,
    deleted_at          TIMESTAMP
);

CREATE INDEX idx_notifications_employee_read
    ON notifications(employee_id, is_read, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_notifications_status
    ON notifications(push_status, push_attempts)
    WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS push_subscriptions;
-- +goose StatementEnd
