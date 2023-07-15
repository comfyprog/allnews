package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/comfyprog/allnews/feed"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testStorage struct {
	err             error
	getArticlesData []feed.Article
}

func (s *testStorage) Ping(ctx context.Context) error {
	return s.err
}

func (s *testStorage) GetArticles(ctx context.Context, options ...GetArticleOption) ([]feed.Article, error) {
	if s.err != nil {
		return s.getArticlesData, s.err
	}

	return s.getArticlesData, nil
}

func TestHealthcheck(t *testing.T) {
	db := &testStorage{}
	r := gin.Default()

	r.GET("/health", handleHealth(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")

	w = httptest.NewRecorder()
	db.err = errors.New("ping error")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "ping error")
}

type resourceFinder struct {
	result []string
	err    error
}

func (f *resourceFinder) GetResourcesWithTags(tags []string) ([]string, error) {
	return f.result, f.err
}

func TestGetArticles(t *testing.T) {
	db := &testStorage{getArticlesData: []feed.Article{}}
	getArticlesData := []feed.Article{
		{
			Resource:    "test",
			Url:         "example.com",
			Title:       "title1",
			Published:   time.Now(),
			Description: "desc1",
			ItemJSON:    "",
		},
	}

	finder := resourceFinder{result: []string{"resource1", "resource2"}, err: nil}

	r := gin.Default()
	r.GET("/articles", handleGetArticles(db, &finder))

	t.Run("happy path", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("with filter", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?filter=str", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("with empty filter", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?filter=", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("with limit", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?limit=1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("with negative limit", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?limit=-1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})

	t.Run("with offset", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?offset=1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("with negative offset", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?offset=-1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})

	t.Run("with start date", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?date_start=2023-07-14T13:00:00Z", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("with wrong start date", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?date_start=2023-07-14T13:00:00+07:00", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})

	t.Run("with end date", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?date_end=2023-07-14T13:00:00Z", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("with wrong end date", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?date_end=2023-07-14T13:00:00+07:00", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})

	t.Run("with multiple params", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet,
			"/articles?date_start=2023-07-14T13:00:00Z&date_end=2023-07-14T13:00:00Z&limit=100&offset=100&filter=text",
			nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("storage returns err", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles", nil)
		db.err = errors.New("storage error")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "storage error")
	})

	t.Run("nothing found", func(t *testing.T) {
		db.err = nil
		db.getArticlesData = []feed.Article{}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "articles")
	})

}

func TestGetArticlesWithTags(t *testing.T) {
	db := &testStorage{getArticlesData: []feed.Article{}}
	getArticlesData := []feed.Article{
		{
			Resource:    "test",
			Url:         "example.com",
			Title:       "title1",
			Published:   time.Now(),
			Description: "desc1",
			ItemJSON:    "",
		},
	}

	finder := resourceFinder{}

	r := gin.Default()
	r.GET("/articles", handleGetArticles(db, &finder))

	t.Run("happy path", func(t *testing.T) {
		finder.result = []string{"resource1", "resource2"}
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?tags[]=tag1:val1&tags[]=tag1:val2&tags[]=tag2:val3", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "title1")
	})

	t.Run("tag error", func(t *testing.T) {
		finder.result = []string{"resource1", "resource2"}
		finder.err = errors.New("tag error")
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?tags[]=tag1:val1&tags[]=tag1:val2&tags[]=tag2:val3", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "tag error")
	})

	t.Run("resources not found", func(t *testing.T) {
		finder.result = []string{}
		finder.err = nil
		db.err = nil
		db.getArticlesData = getArticlesData
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/articles?tags[]=tag1:val1&tags[]=tag1:val2&tags[]=tag2:val3", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "articles")
	})

}
