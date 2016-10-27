package controller

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/healthcheck"
	"github.com/gin-gonic/gin"
)

type HealthcheckController struct {
	*ApplicationController
}

// NewHealthcheckController creates new HealthcheckController object
func NewHealthcheckController(config *config.Config) *HealthcheckController {
	return &HealthcheckController{NewApplicationController(config)}
}

func (self *HealthcheckController) Update(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	appName := c.Param("appName")
	path := c.PostForm("path")

	interval, err := strconv.Atoi(c.PostForm("interval"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusBadRequest, "apps.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "healthcheck interval is invalid.",
		})

		return
	}

	maxTry, err := strconv.Atoi(c.PostForm("maxTry"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.HTML(http.StatusBadRequest, "apps.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "healthcheck maxTry is invalid.",
		})

		return
	}

	hc := healthcheck.NewHealthcheck(path, interval, maxTry)

	if err := healthcheck.Update(username, appName, hc); err != nil {
		c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
			"alert":   true,
			"error":   true,
			"message": "Failed to update healthcheck.",
		})

		return
	}

	c.Redirect(http.StatusSeeOther, "/apps/"+appName)
}
