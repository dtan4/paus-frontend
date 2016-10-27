package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/env"
	"github.com/gin-gonic/gin"
)

type EnvController struct {
	*ApplicationController
}

// NewEnvController creates new EnvController object
func NewEnvController(config *config.Config) *EnvController {
	return &EnvController{NewApplicationController(config)}
}

func (self *EnvController) Delete(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.Param("appName")
	key := c.PostForm("key")

	fmt.Println(key)

	err := env.Delete(username, appName, key)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to detele environment variable.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}

func (self *EnvController) New(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.Param("appName")
	key := c.PostForm("key")
	value := c.PostForm("value")

	err := env.Create(username, appName, key, value)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to add environment variable.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}

func (self *EnvController) Upload(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.Param("appName")

	dotenvFile, _, err := c.Request.FormFile("dotenv")

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to upload dotenv.",
		})

		return
	}

	if err = env.LoadDotenv(username, appName, dotenvFile); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to load dotenv.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}
