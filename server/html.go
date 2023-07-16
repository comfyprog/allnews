package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

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

type errorContext struct {
	Error string
}

func handleIndexPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	}
}
