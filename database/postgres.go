package database

import (
	"context"
	"database/sql"

	"github.com/Camacaro/cqrs/models"
	_ "github.com/lib/pq"
)

type PostgreRepository struct {
	db *sql.DB
}

func NewPostgreRepository(url string) (*PostgreRepository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &PostgreRepository{db: db}, nil
}

func (repo *PostgreRepository) Close() {
	repo.db.Close()
}

func (repo *PostgreRepository) InsertFeed(ctx context.Context, feed *models.Feed) error {
	_, err := repo.db.ExecContext(ctx, "INSERT INTO feeds (id, title, description) VALUES ($1, $2, $3)", feed.ID, feed.Title, feed.Description)
	return err
}

func (repo *PostgreRepository) ListFeeds(ctx context.Context) ([]*models.Feed, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT id, title, description, created_at FROM feeds")
	if err != nil {
		return nil, err
	}

	// Esta lectura debe ser cerrada
	defer rows.Close()

	feeds := []*models.Feed{}

	for rows.Next() {
		feed := &models.Feed{}
		if err := rows.Scan(&feed.ID, &feed.Title, &feed.Description, &feed.CreatedAt); err != nil {
			return nil, err
		}

		feeds = append(feeds, feed)
	}

	return feeds, nil
}
