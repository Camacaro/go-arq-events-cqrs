package search

import (
	"context"

	"github.com/Camacaro/cqrs/models"
)

type SearchRepository interface {
	Close()
	IndexFeed(ctx context.Context, feed *models.Feed) error                // Indexar un nuevo feed
	SearchFeeds(ctx context.Context, query string) ([]*models.Feed, error) // devolvera todos los feed que esten haciendo match con la query
}

var repo SearchRepository

func SetRepository(r SearchRepository) {
	repo = r
}

func Close() {
	repo.Close()
}

func IndexFeed(ctx context.Context, feed *models.Feed) error {
	return repo.IndexFeed(ctx, feed)
}

func SearchFeeds(ctx context.Context, query string) ([]*models.Feed, error) {
	return repo.SearchFeeds(ctx, query)
}
