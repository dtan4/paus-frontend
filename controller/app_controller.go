package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/app"
	"github.com/dtan4/paus-frontend/model/arg"
	"github.com/dtan4/paus-frontend/model/env"
	"github.com/dtan4/paus-frontend/model/user"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/gin"
)

type AppController struct {
	*ApplicationController
}

func NewAppController(config *config.Config, etcd *store.Etcd) *AppController {
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
		fmt.Fprintf(os.Stderr, "%+v\n", err)

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

func (self *AppController) Get(c *gin.Context) {
	var latestURL string

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

	appName := c.Param("appName")

	if !app.Exists(self.etcd, username, appName) {
		c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
			"error":   true,
			"message": fmt.Sprintf("Application %s does not exist.", appName),
		})

		return
	}

	urls, err := app.URLs(self.etcd, self.config.URIScheme, self.config.BaseDomain, username, appName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"error":   true,
			"message": "Failed to list app URLs.",
		})

		return
	}

	envs, err := env.List(self.etcd, username, appName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"error":   true,
			"message": "Failed to list environment variables.",
		})

		return
	}

	buildArgs, err := arg.List(self.etcd, username, appName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"error":   true,
			"message": "Failed to list build args.",
		})

		return
	}

	if len(urls) > 0 {
		latestURL = app.LatestAppURLOfUser(self.config.URIScheme, self.config.BaseDomain, username, appName)
	}

	c.HTML(http.StatusOK, "app.tmpl", gin.H{
		"error":      false,
		"app":        appName,
		"avater_url": user.GetAvaterURL(self.etcd, username),
		"buildArgs":  buildArgs,
		"envs":       envs,
		"latestURL":  latestURL,
		"logged_in":  true,
		"urls":       urls,
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
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "users.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to create app.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}
