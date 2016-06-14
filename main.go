package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/controller"
	"github.com/dtan4/paus-frontend/model/env"
	"github.com/dtan4/paus-frontend/model/user"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/dtan4/paus-frontend/util"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

const (
	AppName = "paus"
)

func initialize(config *server.Config, etcd *store.Etcd) error {
	if !etcd.HasKey("/paus") {
		if err := etcd.Mkdir("/paus"); err != nil {
			return errors.Wrap(err, "Failed to create root directory.")
		}
	}

	if !etcd.HasKey("/paus/users") {
		if err := etcd.Mkdir("/paus/users"); err != nil {
			return errors.Wrap(err, "Failed to create users directory.")
		}
	}

	if err := etcd.Set("/paus/uri-scheme", config.URIScheme); err != nil {
		return errors.Wrap(err, "Failed to set URI scheme in etcd.")
	}

	return nil
}

func currentLoginUser(etcd *store.Etcd, session sessions.Session) string {
	token := session.Get("token")

	if token == nil {
		return ""
	}

	return user.GetLoginUser(etcd, token.(string))
}

func main() {
	printVersion()

	config, err := server.LoadConfig()

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	etcd, err := store.NewEtcd(config.EtcdEndpoint)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = initialize(config, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	oauthConf := oauth2.Config{
		ClientID:     config.GitHubClientID,
		ClientSecret: config.GitHubClientSecret,
		Scopes:       []string{"user", "read:public_key"},
		Endpoint:     githuboauth.Endpoint,
	}

	store := sessions.NewCookieStore([]byte(config.SecretKeyBase))

	r := gin.Default()
	r.Use(sessions.Sessions(AppName, store))
	r.Static("/assets", "assets")
	r.LoadHTMLGlob("templates/*")

	rootController := controller.NewRootController(config, etcd)
	appController := controller.NewAppController(config, etcd)
	argController := controller.NewArgController(config, etcd)
	envController := controller.NewEnvController(config, etcd)

	r.GET("/", rootController.Index)
	r.GET("/apps", appController.Index)
	r.POST("/apps", appController.New)
	r.GET("/apps/:appName", appController.Get)
	r.POST("/apps/:appName/build-args", argController.New)

	// TODO: DELETE /apps/:appName/build-args
	r.POST("/apps/:appName/build-args/delete", argController.Delete)

	r.POST("/apps/:appName/envs", envController.New)

	// TODO: DELETE /apps/:appName/envs
	r.POST("/apps/:appName/envs/delete", envController.Delete)

	r.POST("/apps/:appName/envs/upload", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")

		dotenvFile, _, err := c.Request.FormFile("dotenv")

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to upload dotenv.",
			})

			return
		}

		if err = env.LoadDotenv(etcd, username, appName, dotenvFile); err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to load dotenv.",
			})

			return
		}

		c.Redirect(http.StatusSeeOther, "/apps/"+appName)
	})

	r.GET("/update-keys", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		token := session.Get("token").(string)
		oauthClient := oauthConf.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token})
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
			if !config.SkipKeyUpload {
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
	})

	r.GET("/signin", func(c *gin.Context) {
		state, err := util.GenerateRandomString()

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.String(http.StatusBadRequest, "Failed to generate state string.", err)
			return
		}

		session := sessions.Default(c)
		session.Set("state", state)
		session.Save()

		url := oauthConf.AuthCodeURL(state, oauth2.AccessTypeOnline)
		c.Redirect(http.StatusSeeOther, url)
	})

	r.GET("/signout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Delete("token")
		session.Save()

		c.Redirect(http.StatusSeeOther, "/")
	})

	r.GET("/oauth/callback", func(c *gin.Context) {
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
		token, err := oauthConf.Exchange(oauth2.NoContext, code)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.String(http.StatusBadRequest, "Failed to generate OAuth access token.")
			return
		}

		if !token.Valid() {
			c.String(http.StatusBadRequest, "OAuth access token is invalid.")
			return
		}

		oauthClient := oauthConf.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token.AccessToken})
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
			if !config.SkipKeyUpload {
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

		if !user.Exists(etcd, *u.Login) {
			if err := user.Create(etcd, u); err != nil {
				errors.Fprint(os.Stderr, err)

				c.String(http.StatusBadRequest, "Failed to create user.")
				return
			}
		}

		if err := user.RegisterAccessToken(etcd, *u.Login, token.AccessToken); err != nil {
			errors.Fprint(os.Stderr, err)

			c.String(http.StatusBadRequest, "Failed to register access token.")
			return
		}

		session.Set("token", token.AccessToken)
		session.Save()

		c.Redirect(http.StatusSeeOther, "/")
	})

	r.Run()
}
