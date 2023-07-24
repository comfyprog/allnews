package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/comfyprog/allnews/feed"
	"github.com/gin-gonic/gin"
)

//go:embed public
var public embed.FS
var frontendFs fs.FS
var staticFs fs.FS

func init() {
	var err error
	frontendFs, err = fs.Sub(public, "public")

	if err != nil {
		log.Fatal(err)
	}

	staticFs, err = fs.Sub(frontendFs, "static")

	if err != nil {
		log.Fatal(err)
	}
}

func handleIndexPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Url":   c.Request.URL.Path,
			"Title": "Main page",
		})
	}
}

func handleAboutPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{
			"Url":   c.Request.URL.Path,
			"Title": "About",
		})
	}
}

func handleSearchPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "search.html", gin.H{
			"Url":   c.Request.URL.Path,
			"Title": "Search",
		})
	}
}

type StatsGetter interface {
	GetArticleStats(context.Context) ([]feed.ArticleStats, error)
}

func handleStatsPage(db StatsGetter, tags map[string][]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := db.GetArticleStats(c.Request.Context())
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Url":   c.Request.URL.Path,
				"Title": "Stats",
				"Error": fmt.Sprintf("Error happened: %v", err),
			})
			return
		}

		c.HTML(http.StatusOK, "stats.html", gin.H{
			"Url":       c.Request.URL.Path,
			"Title":     "Stats",
			"Resources": stats,
			"Tags":      tags,
		})
	}
}
