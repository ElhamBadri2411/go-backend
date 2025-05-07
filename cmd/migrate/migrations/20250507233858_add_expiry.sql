-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_invitations ADD COLUMN expiry TIMESTAMP(0) WITH TIME ZONE NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_invitations DROP COLUMN expiry;
-- +goose StatementEnd
