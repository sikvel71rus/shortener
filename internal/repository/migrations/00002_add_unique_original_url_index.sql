-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON shortener (original_url);

-- +goose Down
DROP INDEX IF EXISTS original_url_idx;