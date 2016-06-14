package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/model/app"
	"github.com/dtan4/paus-frontend/model/user"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type AppController struct {
	*ApplicationController
}

func NewAppController(config *server.Config, etcd *store.Etcd) *AppController {
	return &AppController{NewApplicationController(config, etcd)}
}

func (self *AppController) Index(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	if !user.Exists(self.etcd, username) {
		c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
			"error":   true,
			"message": fmt.Sprintf("User %s does not exist.", username),
		})

		return
	}

	apps, err := app.List(self.etcd, username)

	if err != nil {
		errors.Fprint(os.Stderr, err)

		c.HTML(http.StatusInternalServerError, "apps.tmpl", gin.H{
			"error":   true,
			"message": "Failed to list apps.",
		})

		return
	}

	c.HTML(http.StatusOK, "apps.tmpl", gin.H{
		"error":      false,
		"apps":       apps,
		"avater_url": user.GetAvaterURL(self.etcd, username),
		"logged_in":  true,
		"username":   username,
	})
}

func (self *AppController) New(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.PostForm("appName")

	err := app.Create(self.etcd, username, appName)

	if err != nil {
		errors.Fprint(os.Stderr, err)

		c.HTML(http.StatusInternalServerError, "users.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to create app.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}
