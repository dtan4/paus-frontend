package controller

import (
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/model/arg"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type ArgController struct {
	*ApplicationController
}

func NewArgController(config *server.Config, etcd *store.Etcd) *ArgController {
	return &ArgController{NewApplicationController(config, etcd)}
}

func (self *ArgController) Index(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.Param("appName")
	key := c.PostForm("key")
	value := c.PostForm("value")

	err := arg.Create(self.etcd, username, appName, key, value)

	if err != nil {
		errors.Fprint(os.Stderr, err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to add build arg.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}
