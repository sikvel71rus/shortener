-- +goose Up
CREATE TABLE IF NOT EXISTS shortener (
                                         id SERIAL PRIMARY KEY,
                                         short_id VARCHAR(255) UNIQUE NOT NULL,
    original_url TEXT NOT NULL
    );

-- +goose Down
DROP TABLE IF EXISTS shortener;