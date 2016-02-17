package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"alert": false,
		})
	})

	r.POST("/submit", func(c *gin.Context) {
		username := c.PostForm("username")
		pubKey := c.PostForm("pubKey")

		c.HTML(http.StatusCreated, "index.tmpl", gin.H{
			"alert":   true,
			"message": strings.Join([]string{username, ": ", pubKey}, " "),
		})
	})

	r.Run()
}
