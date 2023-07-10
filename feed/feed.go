package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mmcdole/gofeed"
)

type Article struct {
	Resource    string
	Url         string
	Title       string
	Published   time.Time
	Description string
	ItemJSON    string
}

func GetFeed(url string, timeout time.Duration) (*gofeed.Feed, error) {
	parser := gofeed.NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return parser.ParseURLWithContext(url, ctx)
}

func ExtractArticles(feed *gofeed.Feed, resource string) ([]Article, error) {
	articles := make([]Article, 0, len(feed.Items))

	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}
		itemData, err := json.Marshal(item)
		if err != nil {
			return articles, err
		}
		articles = append(articles, Article{
			Resource:    resource,
			Url:         item.Link,
			Title:       item.Title,
			Description: item.Description,
			Published:   *item.PublishedParsed,
			ItemJSON:    string(itemData),
		})
	}

	return articles, nil
}
