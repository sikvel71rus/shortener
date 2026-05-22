package repository

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/sikvel71rus/shortener.git/internal/model"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS
var ErrNotFound = errors.New("url not found")

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

func (r *PostgresRepo) SaveURL(ctx context.Context, id string, originalURL string, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var savedID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO shortener (short_id, original_url)
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING short_id
	`, id, originalURL).Scan(&savedID)

	if errors.Is(err, sql.ErrNoRows) {
		existingID, getErr := r.getShortIDByOriginalURLQuerier(ctx, tx, originalURL)
		if getErr != nil {
			return getErr
		}
		if err := bindUserURL(ctx, tx, userID, existingID); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return ErrConflict
	}

	if err != nil {
		return err
	}

	if err := bindUserURL(ctx, tx, userID, savedID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepo) GetURL(ctx context.Context, id string) (string, error) {
	var originalURL string
	var isDeleted bool
	err := r.db.QueryRowContext(ctx, "SELECT original_url, is_deleted FROM shortener WHERE short_id = $1", id).Scan(&originalURL, &isDeleted)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if isDeleted {
		return "", ErrDeleted
	}

	return originalURL, nil
}

func (r *PostgresRepo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

func (r *PostgresRepo) SaveBatch(ctx context.Context, records []model.BatchRecord, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, rec := range records {
		var savedID string
		err := tx.QueryRowContext(ctx, `
			INSERT INTO shortener (short_id, original_url)
			VALUES ($1, $2)
			ON CONFLICT (original_url) DO NOTHING
			RETURNING short_id
		`, rec.ShortID, rec.OriginalURL).Scan(&savedID)

		if errors.Is(err, sql.ErrNoRows) {
			existingID, getErr := r.getShortIDByOriginalURLQuerier(ctx, tx, rec.OriginalURL)
			if getErr != nil {
				return getErr
			}
			if err := bindUserURL(ctx, tx, userID, existingID); err != nil {
				return err
			}
			continue
		}

		if err != nil {
			return err
		}

		if err := bindUserURL(ctx, tx, userID, savedID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepo) GetShortIDByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	return r.getShortIDByOriginalURLQuerier(ctx, r.db, originalURL)
}

func (r *PostgresRepo) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.short_id, s.original_url
		FROM user_urls u
		JOIN shortener s ON s.short_id = u.short_id
		WHERE u.user_id = $1
		  AND s.is_deleted = FALSE
		ORDER BY s.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UserURL
	for rows.Next() {
		var item model.UserURL
		if err := rows.Scan(&item.ShortURL, &item.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrNoUserURLs
	}

	return result, nil
}

func (r *PostgresRepo) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE shortener
		SET is_deleted = TRUE
		WHERE short_id IN (
			SELECT short_id
			FROM user_urls
			WHERE user_id = $1 AND short_id = ANY($2)
		)
	`, userID, shortIDs)
	return err
}

func (r *PostgresRepo) CountURLs(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shortener").Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PostgresRepo) getShortIDByOriginalURLQuerier(ctx context.Context, querier queryRower, originalURL string) (string, error) {
	var id string
	err := querier.QueryRowContext(ctx,
		"SELECT short_id FROM shortener WHERE original_url = $1",
		originalURL).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return id, err
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func bindUserURL(ctx context.Context, execer execContext, userID, shortID string) error {
	if userID == "" {
		return fmt.Errorf("empty user id")
	}

	_, err := execer.ExecContext(ctx,
		`INSERT INTO user_urls (user_id, short_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID,
		shortID,
	)
	return err
}

type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
