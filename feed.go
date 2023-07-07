package main

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"
)

func getFeed(url string, timeout time.Duration) (*gofeed.Feed, error) {
	parser := gofeed.NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return parser.ParseURLWithContext(url, ctx)
}
