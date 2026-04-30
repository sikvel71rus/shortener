-- +goose Up
ALTER TABLE shortener
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE shortener
    DROP COLUMN IF EXISTS is_deleted;
