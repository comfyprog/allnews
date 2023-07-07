package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mmcdole/gofeed"
)

type article struct {
	resource    string
	url         string
	title       string
	published   *time.Time
	description string
	item        []byte
}

func getFeed(url string, timeout time.Duration) (*gofeed.Feed, error) {
	parser := gofeed.NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return parser.ParseURLWithContext(url, ctx)
}

func extractArticles(feed *gofeed.Feed, resource string) ([]article, error) {
	articles := make([]article, 0, len(feed.Items))

	for _, item := range feed.Items {
		itemData, err := json.Marshal(item)
		if err != nil {
			return articles, err
		}
		articles = append(articles, article{
			resource:    resource,
			url:         item.Link,
			title:       item.Title,
			description: item.Description,
			published:   item.PublishedParsed,
			item:        itemData,
		})
	}

	return articles, nil
}
