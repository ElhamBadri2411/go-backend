-- +goose Up
-- +goose StatementBegin
-- Add new role column w some random default value
ALTER TABLE IF EXISTS users 
ADD COLUMN role_id INT REFERENCES roles(id) DEFAULT 1;

-- update role of all existing users to user
UPDATE users SET role_id = (
SELECT id FROM roles WHERE name = 'user'
);

-- drop default role value
ALTER TABLE users ALTER COLUMN role_id DROP DEFAULT;

-- ensure future roles are not nullable
ALTER TABLE users ALTER COLUMN role_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS users 
DROP COLUMN role_id;
-- +goose StatementEnd
