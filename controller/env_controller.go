package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/env"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type EnvController struct {
	*ApplicationController
}

func NewEnvController(config *config.Config, etcd *store.Etcd) *EnvController {
	return &EnvController{NewApplicationController(config, etcd)}
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

	err := env.Delete(self.etcd, username, appName, key)

	if err != nil {
		errors.Fprint(os.Stderr, err)

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

	err := env.Create(self.etcd, username, appName, key, value)

	if err != nil {
		errors.Fprint(os.Stderr, err)

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
		errors.Fprint(os.Stderr, err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to upload dotenv.",
		})

		return
	}

	if err = env.LoadDotenv(self.etcd, username, appName, dotenvFile); err != nil {
		errors.Fprint(os.Stderr, err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to load dotenv.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}
