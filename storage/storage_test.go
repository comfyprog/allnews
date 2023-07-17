package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/comfyprog/allnews/feed"
	"github.com/comfyprog/allnews/server"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type clearDbFunc func() error

var (
	connStr string
	db      *sql.DB
	clearDb clearDbFunc
)

func prepareDb() (string, *sql.DB, clearDbFunc, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		// WaitingFor:   wait.ForListeningPort("5432/tcp"),
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		),
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
		},
	}

	dbContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)

	if err != nil {
		return "", nil, nil, err
	}

	host, err := dbContainer.Host(context.Background())
	if err != nil {
		return "", nil, nil, err
	}

	port, err := dbContainer.MappedPort(context.Background(), "5432")
	if err != nil {
		return "", nil, nil, err
	}

	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", nil, nil, err
	}

	err = db.Ping()
	if err != nil {
		return "", nil, nil, err
	}

	clearDbFunc := func() error {
		_, err := db.Exec("DELETE FROM articles;")
		return err
	}

	return connStr, db, clearDbFunc, nil
}

func TestMain(m *testing.M) {
	var err error
	connStr, db, clearDb, err = prepareDb()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code := m.Run()

	os.Exit(code)
}

func TestMigrate(t *testing.T) {
	Migrate(connStr)

	rows, err := db.Query("select id, resource_name, url, title, description, published, feed_item from articles")
	assert.Nil(t, err)

	if rows.Next() {
		t.Error("expected no rows")
	}

	err = rows.Err()
	assert.Nil(t, err)

	rows.Close()
}

func TestSaveArticles(t *testing.T) {
	err := clearDb()
	assert.Nil(t, err)

	articles := []feed.Article{
		{Resource: "resource1", Url: "example.com", Title: "title1", Published: time.Now(), Description: "description1", ItemJSON: `{"item": 1}`},
		{Resource: "resource2", Url: "google.com", Title: "title2", Published: time.Now().Add(time.Hour), Description: "description2", ItemJSON: `{"item": 2}`},
	}

	storage, err := NewPostgresStorage(connStr)
	assert.Nil(t, err)

	err = storage.SaveArticles(context.Background(), articles)
	assert.Nil(t, err)

	var count int
	err = db.QueryRow("select count(*) from articles;").Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
}

func TestSaveArticlesWithConflictingFields(t *testing.T) {
	err := clearDb()
	assert.Nil(t, err)

	articles := []feed.Article{
		{Resource: "resource1", Url: "example.com", Title: "title1", Published: time.Now(), Description: "description1", ItemJSON: `{"item": 1}`},
		{Resource: "resource2", Url: "google.com", Title: "title2", Published: time.Now().Add(time.Hour), Description: "description2", ItemJSON: `{"item": 2}`},
		{Resource: "resource1", Url: "example.com", Title: "title1", Published: time.Now(), Description: "description1", ItemJSON: `{"item": 1}`},
	}

	storage, err := NewPostgresStorage(connStr)
	assert.Nil(t, err)

	err = storage.SaveArticles(context.Background(), articles)
	assert.Nil(t, err)

	var count int
	err = db.QueryRow("select count(*) from articles;").Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
}

func TestGetArticles(t *testing.T) {
	storage, err := NewPostgresStorage(connStr)
	assert.Nil(t, err)

	err = clearDb()
	assert.Nil(t, err)

	articles := []feed.Article{
		{Resource: "resource1", Url: "google..com", Title: "title1", Published: time.Now(), Description: "description1", ItemJSON: "{}"},
		{Resource: "resource2", Url: "yahoo.com", Title: "title2", Published: time.Now().Add(time.Hour), Description: "description2", ItemJSON: "{}"},
		{Resource: "resource3", Url: "bing.com", Title: "title3", Published: time.Now().Add(time.Hour * -25), Description: "description3", ItemJSON: "{}"},
	}

	ctx := context.Background()

	err = storage.SaveArticles(ctx, articles)
	assert.Nil(t, err)

	t.Run("with default params", func(t *testing.T) {
		retrived, err := storage.GetArticles(ctx)
		assert.Nil(t, err)
		assert.Len(t, retrived, 3)
		assert.Equal(t, "title2", retrived[0].Title)
		assert.Equal(t, "title1", retrived[1].Title)
		assert.Equal(t, "title3", retrived[2].Title)
	})

	t.Run("with filter", func(t *testing.T) {
		retrived, err := storage.GetArticles(ctx, server.WithFilter("title2"))
		assert.Nil(t, err)
		assert.Len(t, retrived, 1)
		assert.Equal(t, "title2", retrived[0].Title)
	})

	t.Run("with date start", func(t *testing.T) {
		retrived, err := storage.GetArticles(ctx, server.WithDateStart(time.Now().Add(time.Hour*-26)))
		assert.Nil(t, err)
		assert.Len(t, retrived, 3)
		assert.Equal(t, "title2", retrived[0].Title)
		assert.Equal(t, "title1", retrived[1].Title)
		assert.Equal(t, "title3", retrived[2].Title)
	})

	t.Run("with date end", func(t *testing.T) {
		retrived, err := storage.GetArticles(ctx,
			server.WithDateStart(time.Now().Add(time.Hour*-26)),
			server.WithDateEnd(time.Now().Add(time.Hour*-24)))
		assert.Nil(t, err)
		assert.Len(t, retrived, 1)
		assert.Equal(t, "title3", retrived[0].Title)
	})

	t.Run("with limit", func(t *testing.T) {
		retrived, err := storage.GetArticles(ctx, server.WithLimit(1))
		assert.Nil(t, err)
		assert.Len(t, retrived, 1)
		assert.Equal(t, "title2", retrived[0].Title)
	})

	t.Run("with offset", func(t *testing.T) {
		retrived, err := storage.GetArticles(ctx, server.WithLimit(1), server.WithOffset(1))
		assert.Nil(t, err)
		assert.Len(t, retrived, 1)
		assert.Equal(t, "title1", retrived[0].Title)
	})

	t.Run("with resource names", func(t *testing.T) {
		retrived, err := storage.GetArticles(ctx, server.WithResourceNames([]string{"resource2", "resource1"}))
		assert.Nil(t, err)
		assert.Len(t, retrived, 2)
		assert.Equal(t, "title2", retrived[0].Title)
		assert.Equal(t, "title1", retrived[1].Title)
	})
}

func TestGetArticleStats(t *testing.T) {
	storage, err := NewPostgresStorage(connStr)
	assert.Nil(t, err)

	err = clearDb()
	assert.Nil(t, err)

	now := time.Now()
	articles := []feed.Article{
		{Resource: "resource1", Url: "google..com", Title: "title1", Published: now, Description: "description1", ItemJSON: "{}"},
		{Resource: "resource2", Url: "yahoo.com", Title: "title2", Published: now.Add(time.Hour), Description: "description2", ItemJSON: "{}"},
		{Resource: "resource2", Url: "bing.com", Title: "title3", Published: now.Add(time.Hour * -25), Description: "description3", ItemJSON: "{}"},
	}

	ctx := context.Background()

	err = storage.SaveArticles(ctx, articles)
	assert.Nil(t, err)

	stats, err := storage.GetArticleStats(ctx)
	assert.Nil(t, err)

	utc, err := time.LoadLocation("")
	assert.Nil(t, err)

	assert.Len(t, stats, 2)
	assert.Equal(t, "resource1", stats[0].Resource)
	assert.Equal(t, 1, stats[0].TotalArticles)
	assert.Less(t, now.In(utc).Sub(stats[0].FirstDate), time.Millisecond)
	assert.Less(t, now.In(utc).Sub(stats[0].LastDate), time.Millisecond)

	assert.Equal(t, "resource2", stats[1].Resource)
	assert.Equal(t, 2, stats[1].TotalArticles)
	assert.Less(t, now.In(utc).Add(time.Hour*-25).Sub(stats[1].FirstDate), time.Millisecond)
	assert.Less(t, now.In(utc).Add(time.Hour).Sub(stats[1].LastDate), time.Millisecond)
}
