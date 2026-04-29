-- +goose Up
CREATE TABLE IF NOT EXISTS user_urls (
    user_id VARCHAR(255) NOT NULL,
    short_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (user_id, short_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_urls;
