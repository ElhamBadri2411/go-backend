-- +goose Up
-- +goose StatementBegin
ALTER TABLE posts
  ADD COLUMN tags VARCHAR(100) [];

ALTER TABLE posts
  ADD COLUMN updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW() ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE posts
  DROP COLUMN tags;

ALTER TABLE posts
  DROP COLUMN updated_at;
-- +goose StatementEnd
