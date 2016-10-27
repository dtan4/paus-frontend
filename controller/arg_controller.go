package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/arg"
	"github.com/gin-gonic/gin"
)

type ArgController struct {
	*ApplicationController
}

// NewArgController creates new ArgController
func NewArgController(config *config.Config) *ArgController {
	return &ArgController{NewApplicationController(config)}
}

func (self *ArgController) Delete(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.Param("appName")
	key := c.PostForm("key")

	err := arg.Delete(username, appName, key)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to delete build arg.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}

func (self *ArgController) New(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.Param("appName")
	key := c.PostForm("key")
	value := c.PostForm("value")

	err := arg.Create(username, appName, key, value)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to add build arg.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}
