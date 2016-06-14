package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dtan4/paus-frontend/controller"
	"github.com/dtan4/paus-frontend/model/app"
	"github.com/dtan4/paus-frontend/model/arg"
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

	r.GET("/", rootController.Index)

	r.GET("/apps", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		if !user.Exists(etcd, username) {
			c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("User %s does not exist.", username),
			})

			return
		}

		apps, err := app.List(etcd, username)

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
			"avater_url": user.GetAvaterURL(etcd, username),
			"logged_in":  true,
			"username":   username,
		})
	})

	r.POST("/apps", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.PostForm("appName")

		err := app.Create(etcd, username, appName)

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
	})

	r.GET("/apps/:appName", func(c *gin.Context) {
		var latestURL string

		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		if !user.Exists(etcd, username) {
			c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("User %s does not exist.", username),
			})

			return
		}

		appName := c.Param("appName")

		if !app.Exists(etcd, username, appName) {
			c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("Application %s does not exist.", appName),
			})

			return
		}

		urls, err := app.URLs(etcd, config.URIScheme, config.BaseDomain, username, appName)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": "Failed to list app URLs.",
			})

			return
		}

		envs, err := env.List(etcd, username, appName)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": "Failed to list environment variables.",
			})

			return
		}

		buildArgs, err := arg.List(etcd, username, appName)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": "Failed to list build args.",
			})

			return
		}

		if len(urls) > 0 {
			latestURL = app.LatestAppURLOfUser(config.URIScheme, config.BaseDomain, username, appName)
		}

		c.HTML(http.StatusOK, "app.tmpl", gin.H{
			"error":      false,
			"app":        appName,
			"avater_url": user.GetAvaterURL(etcd, username),
			"buildArgs":  buildArgs,
			"envs":       envs,
			"latestURL":  latestURL,
			"logged_in":  true,
			"urls":       urls,
			"username":   username,
		})
	})

	r.POST("/apps/:appName/build-args", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := arg.Create(etcd, username, appName, key, value)

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
	})

	// TODO: DELETE /apps/:appName/build-args
	r.POST("/apps/:appName/build-args/delete", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")

		err := arg.Delete(etcd, username, appName, key)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to delete build arg.",
			})

			return
		}

		c.Redirect(http.StatusSeeOther, "/apps/"+appName)
	})

	r.POST("/apps/:appName/envs", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := env.Create(etcd, username, appName, key, value)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to add environment variable.",
			})

			return
		}

		c.Redirect(http.StatusSeeOther, "/apps/"+appName)
	})

	// TODO: DELETE /apps/:appName/envs
	r.POST("/apps/:appName/envs/delete", func(c *gin.Context) {
		session := sessions.Default(c)
		username := currentLoginUser(etcd, session)

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")

		fmt.Println(key)

		err := env.Delete(etcd, username, appName, key)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to detele environment variable.",
			})

			return
		}

		c.Redirect(http.StatusSeeOther, "/apps/"+appName)
	})

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
