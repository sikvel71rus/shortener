package repository

import (
	"context"
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) SaveURL(ctx context.Context, id, originalURL string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO shortener (short_id, original_url) VALUES ($1, $2)", id, originalURL)
	return err
}

func (r *PostgresRepo) GetURL(ctx context.Context, id string) (string, error) {
	var originalURL string
	err := r.db.QueryRowContext(ctx, "SELECT original_url FROM shortener WHERE short_id = $1", id).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return originalURL, err
}

func (r *PostgresRepo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresRepo) CheckIfURLExist(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM shortener WHERE short_id = $1)", id).Scan(&exists)
	return exists, err
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}
