package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/comfyprog/allnews/feed"
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
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
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
