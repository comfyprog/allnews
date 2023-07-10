package storage

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/comfyprog/allnews/feed"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) SaveArticles(ctx context.Context, articles []feed.Article) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	insert := psql.Insert("articles").
		Columns("resource_name", "url", "title", "description", "published", "feed_item")

	for _, a := range articles {
		insert = insert.Values(a.Resource, a.Url, a.Title, a.Description, a.Published, a.ItemJSON)
	}

	insert = insert.Suffix("ON CONFLICT (url) DO NOTHING")

	query, args, err := insert.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)

	return err
}
