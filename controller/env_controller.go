package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/model/env"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type EnvController struct {
	*ApplicationController
}

func NewEnvController(config *server.Config, etcd *store.Etcd) *EnvController {
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

func (self *EnvController) Index(c *gin.Context) {
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
