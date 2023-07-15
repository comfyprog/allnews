package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/comfyprog/allnews/feed"
	"github.com/comfyprog/allnews/server"
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

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
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

func (s *PostgresStorage) GetArticles(ctx context.Context, options ...server.GetArticleOption) ([]feed.Article, error) {
	searchParams, err := server.NewArticleSearchParams()
	if err != nil {
		return []feed.Article{}, err
	}
	for _, f := range options {
		f(searchParams)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := psql.Select("resource_name", "url", "title", "description", "published").
		From("articles").
		OrderBy("published DESC").
		Limit(searchParams.Limit).Offset(searchParams.Offset)

	builder = builder.Where(squirrel.GtOrEq{"published": searchParams.DateStart})
	builder = builder.Where(squirrel.LtOrEq{"published": searchParams.DateEnd})

	if searchParams.Filter != "" {
		builder = builder.Where("title ILIKE ?", fmt.Sprintf("%%%s%%", searchParams.Filter))
	}

	if len(searchParams.Resources) > 0 {
		builder = builder.Where(map[string]interface{}{"resource_name": searchParams.Resources})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return []feed.Article{}, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []feed.Article{}, err
	}

	defer rows.Close()

	result := make([]feed.Article, 0)
	for rows.Next() {
		var a feed.Article
		err := rows.Scan(&a.Resource, &a.Url, &a.Title, &a.Description, &a.Published)
		if err != nil {
			return result, err
		}
		result = append(result, a)
	}

	err = rows.Err()
	if err != nil {
		return result, err
	}

	return result, nil
}
