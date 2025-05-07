-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_invitations (
  token bytea NOT NULL, 
  user_id bigint NOT NULL,
  PRIMARY KEY (token, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_invitations;
-- +goose StatementEnd
