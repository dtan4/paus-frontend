package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/model/user"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/dtan4/paus-frontend/util"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
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

func (self *SessionController) UpdateKeys(c *gin.Context) {
	username := self.CurrentUser(c)

	if username == "" {
		c.Redirect(http.StatusFound, "/")

		return
	}

	session := sessions.Default(c)
	token := session.Get("token").(string)
	oauthClient := self.oauthConf.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token})
	client := github.NewClient(oauthClient)

	u, _, err := client.Users.Get("")

	if err != nil {
		c.String(http.StatusBadRequest, "Failed to retrive GitHub user profile.")
	}

	keys, _, err := client.Users.ListKeys("", &github.ListOptions{})

	if err != nil {
		errors.Fprint(os.Stderr, err)

		c.String(http.StatusBadRequest, "Failed to retrive SSH public keys from GitHub.")
		return
	}

	for _, key := range keys {
		if !self.config.SkipKeyUpload {
			_, err := user.UploadPublicKey(*u.Login, *key.Key)

			if err != nil {
				errors.Fprint(os.Stderr, err)

				c.String(http.StatusBadRequest, "Failed to register SSH public key.")
				return
			}
		} else {
			fmt.Printf("User: %s, Key: %s\n", *u.Login, *key.Key)
		}
	}

	c.Redirect(http.StatusSeeOther, "/")
}
