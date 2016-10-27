package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/user"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/gin"
)

type RootController struct {
	*ApplicationController
}

func NewRootController(config *config.Config, etcd *store.Etcd) *RootController {
	return &RootController{NewApplicationController(config, etcd)}
}

func (self *RootController) Index(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"alert":      false,
			"error":      false,
			"logged_in":  false,
			"message":    "",
			"baseDomain": self.config.BaseDomain,
		})

		return
	}

	if !user.Exists(username) {
		c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
			"error":   true,
			"message": fmt.Sprintf("User %s does not exist.", username),
		})

		return
	}

	avaterURL, err := user.GetAvaterURL(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "apps.tmpl", gin.H{
			"error":   true,
			"message": "Failed to get avater URL.",
		})

		return
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"alert":      false,
		"avater_url": avaterURL,
		"error":      false,
		"logged_in":  true,
		"message":    "",
		"username":   username,
		"baseDomain": self.config.BaseDomain,
	})
}
