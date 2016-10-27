package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/user"
	"github.com/dtan4/paus-frontend/store"
	"github.com/dtan4/paus-frontend/util"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type SessionController struct {
	*ApplicationController
	oauthConf oauth2.Config
}

func NewSessionController(config *config.Config, etcd *store.Etcd, oauthConf oauth2.Config) *SessionController {
	return &SessionController{NewApplicationController(config, etcd), oauthConf}
}

func (self *SessionController) Callback(c *gin.Context) {
	session := sessions.Default(c)

	if session.Get("state") == nil {
		c.String(http.StatusBadRequest, "State string is not stored in session.")
		return
	}

	storedState := session.Get("state").(string)
	state := c.Query("state")

	if state != storedState {
		c.String(http.StatusUnauthorized, "State string does not match.")
		return
	}

	session.Delete("state")

	code := c.Query("code")
	token, err := self.oauthConf.Exchange(oauth2.NoContext, code)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.String(http.StatusBadRequest, "Failed to generate OAuth access token.")
		return
	}

	if !token.Valid() {
		c.String(http.StatusBadRequest, "OAuth access token is invalid.")
		return
	}

	oauthClient := self.oauthConf.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token.AccessToken})
	client := github.NewClient(oauthClient)

	u, _, err := client.Users.Get("")

	if err != nil {
		c.String(http.StatusBadRequest, "Failed to retrive GitHub user profile.")
	}

	keys, _, err := client.Users.ListKeys("", &github.ListOptions{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.String(http.StatusBadRequest, "Failed to retrive SSH public keys from GitHub.")
		return
	}

	for _, key := range keys {
		if !self.config.SkipKeyUpload {
			_, err := user.UploadPublicKey(*u.Login, *key.Key)

			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v\n", err)

				c.String(http.StatusBadRequest, "Failed to register SSH public key.")
				return
			}
		} else {
			fmt.Printf("User: %s, Key: %s\n", *u.Login, *key.Key)
		}
	}

	if !user.Exists(*u.Login) {
		if err := user.Create(u); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)

			c.String(http.StatusBadRequest, "Failed to create user.")
			return
		}
	}

	if err := user.RegisterAccessToken(self.etcd, *u.Login, token.AccessToken); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.String(http.StatusBadRequest, "Failed to register access token.")
		return
	}

	session.Set("token", token.AccessToken)
	session.Save()

	c.Redirect(http.StatusSeeOther, "/")
}

func (self *SessionController) SignIn(c *gin.Context) {
	state, err := util.GenerateRandomString()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)

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
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		c.String(http.StatusBadRequest, "Failed to retrive SSH public keys from GitHub.")
		return
	}

	for _, key := range keys {
		if !self.config.SkipKeyUpload {
			_, err := user.UploadPublicKey(*u.Login, *key.Key)

			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v\n", err)

				c.String(http.StatusBadRequest, "Failed to register SSH public key.")
				return
			}
		} else {
			fmt.Printf("User: %s, Key: %s\n", *u.Login, *key.Key)
		}
	}

	c.Redirect(http.StatusSeeOther, "/")
}
