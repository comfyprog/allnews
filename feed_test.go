package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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
			feed, err := getFeed(test.url, time.Second)
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

	feed, err := getFeed(srv.URL, time.Millisecond)
	assert.NotNil(t, err)
	assert.Nil(t, feed)
}

func TestGetFeedUnparseable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/xml;charset=UTF-8")
		w.Write([]byte("null"))
	}))

	feed, err := getFeed(srv.URL, time.Millisecond)
	assert.NotNil(t, err)
	assert.Nil(t, feed)
}
