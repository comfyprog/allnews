package feed

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/comfyprog/allnews/config"
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

type ArticleSaver interface {
	SaveArticles(context.Context, []Article) error
}

func processFeed(feedConfig config.SourceConfig, storage ArticleSaver) {
	log.Printf("Getting %s", feedConfig.FeedUrl)
	feed, err := GetFeed(feedConfig.FeedUrl, feedConfig.Timeout)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	articles, err := ExtractArticles(feed, feedConfig.Name)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	err = storage.SaveArticles(context.Background(), articles)
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func ProcessFeeds(ctx context.Context, feedGroups map[string][]config.SourceConfig, storage ArticleSaver) {
	for groupName := range feedGroups {
		log.Printf("Processing feed group `%s`", groupName)
		for _, feedConfig := range feedGroups[groupName] {
			processFeed(feedConfig, storage)
		}

	}
}
