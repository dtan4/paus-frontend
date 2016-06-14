package controller

import (
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/dtan4/paus-frontend/util"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type SessionController struct {
	*ApplicationController
	oauthConf oauth2.Config
}

func NewSessionController(config *server.Config, etcd *store.Etcd, oauthConf oauth2.Config) *SessionController {
	return &SessionController{NewApplicationController(config, etcd), oauthConf}
}

func (self *SessionController) SignIn(c *gin.Context) {
	state, err := util.GenerateRandomString()

	if err != nil {
		errors.Fprint(os.Stderr, err)

		c.String(http.StatusBadRequest, "Failed to generate state string.", err)
		return
	}

	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()

	url := self.oauthConf.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusSeeOther, url)
}

func (self *SessionController) SignOut(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("token")
	session.Save()

	c.Redirect(http.StatusSeeOther, "/")
}
