package storage_test

import (
	"context"
	"fmt"
	"testing"

	"database/sql"

	"github.com/comfyprog/allnews/storage"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func prepareDb() (string, *sql.DB, error) {
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
		return "", nil, err
	}

	host, err := dbContainer.Host(context.Background())
	if err != nil {
		return "", nil, err
	}

	port, err := dbContainer.MappedPort(context.Background(), "5432")
	if err != nil {
		return "", nil, err
	}

	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", nil, err
	}

	err = db.Ping()
	if err != nil {
		return "", nil, err
	}

	return connStr, db, nil
}

func TestMigrate(t *testing.T) {
	connStr, db, err := prepareDb()

	if err != nil {
		t.Fatal(err)
	}

	storage.Migrate(connStr)

	rows, err := db.Query("select id, resource_name, url, title, description, published, feed_item from articles")
	if err != nil {
		t.Error(err)
	}

	if rows.Next() {
		t.Error("expected no rows")
	}

	err = rows.Err()
	if err != nil {
		t.Fatal(err)
	}

	rows.Close()
}
