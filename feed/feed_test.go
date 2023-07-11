package feed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/comfyprog/allnews/config"
	"github.com/stretchr/testify/assert"
)

func TestGetFeed(t *testing.T) {
	tt := []struct {
		file     string
		url      string
		numItems int
	}{}

	files := []struct {
		name     string
		numItems int
	}{
		{"./testdata/rss1.xml", 2},
		{"./testdata/rss2.xml", 3},
	}

	for _, file := range files {
		data, err := os.ReadFile(file.name)
		if err != nil {
			t.Fatal(err)
		}

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("content-type", "text/xml;charset=UTF-8")
			w.Write(data)
		}))

		tt = append(tt, struct {
			file     string
			url      string
			numItems int
		}{file.name, srv.URL, file.numItems})
	}

	for _, test := range tt {
		t.Run(test.file, func(t *testing.T) {
			feed, err := GetFeed(test.url, time.Second)
			assert.Nil(t, err)
			assert.NotNil(t, feed)
			assert.Len(t, feed.Items, test.numItems)
		})
	}
}

func TestGetFeedTimeout(t *testing.T) {
	data, err := os.ReadFile("./testdata/rss1.xml")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(time.Millisecond * 100)
		w.Header().Add("content-type", "text/xml;charset=UTF-8")
		w.Write(data)
	}))

	feed, err := GetFeed(srv.URL, time.Millisecond)
	assert.NotNil(t, err)
	assert.Nil(t, feed)
}

func TestGetFeedUnparseable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/xml;charset=UTF-8")
		w.Write([]byte("null"))
	}))

	feed, err := GetFeed(srv.URL, time.Millisecond)
	assert.NotNil(t, err)
	assert.Nil(t, feed)
}

func TestExtractArticles(t *testing.T) {
	data, err := os.ReadFile("./testdata/rss1.xml")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/xml;charset=UTF-8")
		w.Write(data)
	}))

	feed, err := GetFeed(srv.URL, time.Millisecond)
	assert.Nil(t, err)
	assert.NotNil(t, feed)

	articles, err := ExtractArticles(feed, "site1")
	assert.Nil(t, err)

	assert.Len(t, articles, len(feed.Items))

	for _, a := range articles {
		t.Run(a.Url, func(t *testing.T) {
			assert.Equal(t, "site1", a.Resource)
			assert.Greater(t, len(a.ItemJSON), 0)
		})
	}

}

type testStorage struct {
	articles []Article
}

func newTestStorage() *testStorage {
	s := testStorage{}
	s.articles = make([]Article, 0)
	return &s
}

func (s *testStorage) SaveArticles(ctx context.Context, articles []Article) error {
	s.articles = append(s.articles, articles...)
	return nil
}

func TestProcessFeeds(t *testing.T) {
	data, err := os.ReadFile("./testdata/rss1.xml")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/xml;charset=UTF-8")
		w.Write(data)
	}))

	feedGroups := map[string][]config.SourceConfig{
		"testgroup": []config.SourceConfig{{
			Name:         "test",
			FeedUrl:      srv.URL,
			Timeout:      time.Second,
			UpdatePeriod: time.Second,
			Country:      "us",
		}},
	}

	storage := newTestStorage()

	ProcessFeeds(context.Background(), feedGroups, storage, false)

	assert.Len(t, storage.articles, 2)

}

func TestProcessFeedsContinuously(t *testing.T) {
	data, err := os.ReadFile("./testdata/rss1.xml")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/xml;charset=UTF-8")
		w.Write(data)
	}))

	feedGroups := map[string][]config.SourceConfig{
		"testgroup": []config.SourceConfig{{
			Name:         "test",
			FeedUrl:      srv.URL,
			Timeout:      time.Second,
			UpdatePeriod: time.Millisecond * 100,
			Country:      "us",
		}},
	}

	storage := newTestStorage()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*150)
	defer cancel()

	ProcessFeeds(ctx, feedGroups, storage, true)

	assert.Len(t, storage.articles, 4)

}
