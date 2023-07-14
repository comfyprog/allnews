package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (s *PostgresStorage) Ping() error {
	return s.db.Ping()
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

type articleSearchParams struct {
	dateStart string
	dateEnd   string
	filter    string
	limit     uint64
	offset    uint64
}

func newArticleSearchParams() (*articleSearchParams, error) {
	now := time.Now()
	year, month, day := now.Date()
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return &articleSearchParams{}, err
	}

	start := time.Date(year, month, day, 0, 0, 0, 0, loc)
	end := time.Date(year, month, day, 23, 59, 59, 0, loc)

	return &articleSearchParams{
			dateStart: start.Format(time.RFC3339),
			dateEnd:   end.Format(time.RFC3339),
			filter:    "",
			limit:     50,
			offset:    0,
		},
		nil
}

type GetArticleOption func(*articleSearchParams)

func WithDateStart(dateStart string) GetArticleOption {
	return func(p *articleSearchParams) {
		p.dateStart = dateStart
	}
}

func WithDateEnd(dateEnd string) GetArticleOption {
	return func(p *articleSearchParams) {
		p.dateEnd = dateEnd
	}
}

func WithFilter(filter string) GetArticleOption {
	return func(p *articleSearchParams) {
		p.filter = filter
	}
}

func WithLimit(limit uint64) GetArticleOption {
	return func(p *articleSearchParams) {
		p.limit = limit
	}
}

func WithOffset(offset uint64) GetArticleOption {
	return func(p *articleSearchParams) {
		p.offset = offset
	}
}

func (s *PostgresStorage) GetArticles(ctx context.Context, options ...GetArticleOption) ([]feed.Article, error) {
	searchParams, err := newArticleSearchParams()
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
		Limit(searchParams.limit).Offset(searchParams.offset)

	builder = builder.Where(squirrel.GtOrEq{"published": searchParams.dateStart})
	builder = builder.Where(squirrel.LtOrEq{"published": searchParams.dateEnd})

	if searchParams.filter != "" {
		builder = builder.Where("title LIKE ?", fmt.Sprintf("%%%s%%", searchParams.filter))
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
