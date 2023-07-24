package server

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/comfyprog/allnews/config"
	"github.com/comfyprog/allnews/feed"
	"github.com/gin-gonic/gin"
)

type DbPinger interface {
	Ping(context.Context) error
}

func handleHealth(db DbPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}

type ArticleSearchParams struct {
	DateStart time.Time `form:"date_start" time_format:"2006-01-02T15:04:05Z07:00"`
	DateEnd   time.Time `form:"date_end" time_format:"2006-01-02T15:04:05Z07:00"`
	Filter    string    `form:"filter"`
	Limit     uint64    `form:"limit" binding:"gte=0"`
	Offset    uint64    `form:"offset" binding:"gte=0"`
	Tags      []string  `form:"tags[]"`
	Resources []string  `form:"-"`
}

func NewArticleSearchParams() (*ArticleSearchParams, error) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return &ArticleSearchParams{}, err
	}
	start := time.Date(1970, 1, 1, 0, 0, 1, 0, loc)
	now := time.Now()
	year, month, day := now.Date()
	end := time.Date(year, month, day, 23, 59, 59, 0, loc)

	return &ArticleSearchParams{
			DateStart: start,
			DateEnd:   end,
			Filter:    "",
			Limit:     50,
			Offset:    0,
		},
		nil
}

type GetArticleOption func(*ArticleSearchParams)

func WithDateStart(dateStart time.Time) GetArticleOption {
	return func(p *ArticleSearchParams) {
		p.DateStart = dateStart
	}
}

func WithDateEnd(dateEnd time.Time) GetArticleOption {
	return func(p *ArticleSearchParams) {
		p.DateEnd = dateEnd
	}
}

func WithFilter(filter string) GetArticleOption {
	return func(p *ArticleSearchParams) {
		p.Filter = filter
	}
}

func WithLimit(limit uint64) GetArticleOption {
	return func(p *ArticleSearchParams) {
		p.Limit = limit
	}
}

func WithOffset(offset uint64) GetArticleOption {
	return func(p *ArticleSearchParams) {
		p.Offset = offset
	}
}

func WithResourceNames(names []string) GetArticleOption {
	return func(p *ArticleSearchParams) {
		p.Resources = names
	}
}

type ArticleGetter interface {
	GetArticles(context.Context, ...GetArticleOption) ([]feed.Article, error)
}

type TaggedResourcesGetter interface {
	GetResourcesWithTags([]string) ([]string, error)
}

func handleGetArticles(db ArticleGetter, config TaggedResourcesGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params ArticleSearchParams
		if err := c.ShouldBind(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		options := []GetArticleOption{}
		if !params.DateStart.IsZero() {
			options = append(options, WithDateStart(params.DateStart))
		}
		if !params.DateEnd.IsZero() {
			options = append(options, WithDateEnd(params.DateEnd))
		}
		if params.Limit != 0 {
			options = append(options, WithLimit(params.Limit))
		}
		if params.Offset != 0 {
			options = append(options, WithOffset(params.Offset))
		}
		if params.Filter != "" {
			options = append(options, WithFilter(params.Filter))
		}

		if len(params.Tags) > 0 {
			resourceNames, err := config.GetResourcesWithTags(params.Tags)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if len(resourceNames) == 0 {
				c.JSON(http.StatusNotFound, gin.H{"articles": []feed.Article{}})
				return
			}

			options = append(options, WithResourceNames(resourceNames))
		}

		articles, err := db.GetArticles(c.Request.Context(), options...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(articles) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"articles": []feed.Article{}})
			return
		}

		c.JSON(http.StatusOK, gin.H{"articles": articles})
	}
}

func handleGetTags(tags map[string][]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tags": tags})
	}
}

type ServerStorage interface {
	DbPinger
	ArticleGetter
	StatsGetter
}

func Serve(db ServerStorage, config config.Config) error {
	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.StaticFS("/static", http.FS(staticFs))

	tmpl := template.Must(template.ParseFS(frontendFs, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/stats", handleStatsPage(db, config.GetAllTags()))
	r.GET("/search", handleSearchPage())
	r.GET("/about", handleAboutPage())
	r.GET("/health", handleHealth(db))

	api := r.Group("/api/v1")
	api.GET("/articles", handleGetArticles(db, config))
	api.GET("/tags", handleGetTags(config.GetAllTags()))

	r.GET("/", handleIndexPage())
	log.Printf("Listening on %s", config.ListenAddr)

	return r.Run(config.ListenAddr)
}
