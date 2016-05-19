package main

import (
	"fmt"
	"net/http"
	"os"

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

func initialize(config *Config, etcd *Etcd) error {
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

func main() {
	config, err := LoadConfig()

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	etcd, err := NewEtcd(config.EtcdEndpoint)

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

	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"alert":      false,
				"error":      false,
				"logged_in":  false,
				"message":    "",
				"baseDomain": config.BaseDomain,
			})

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"alert":      false,
				"error":      false,
				"logged_in":  false,
				"message":    "",
				"baseDomain": config.BaseDomain,
			})

			return
		}

		if !UserExists(etcd, username) {
			c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("User %s does not exist.", username),
			})

			return
		}

		if err != nil {
			c.String(http.StatusNotFound, "User not found.")
			return
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"alert":      false,
			"avater_url": GetAvaterURL(etcd, username),
			"error":      false,
			"logged_in":  true,
			"message":    "",
			"username":   username,
			"baseDomain": config.BaseDomain,
		})
	})

	r.GET("/apps", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		if !UserExists(etcd, username) {
			c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("User %s does not exist.", username),
			})

			return
		}

		apps, err := Apps(etcd, username)

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
			"avater_url": GetAvaterURL(etcd, username),
			"logged_in":  true,
			"username":   username,
		})
	})

	r.POST("/apps", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.PostForm("appName")

		err := CreateApp(etcd, username, appName)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "users.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to create app.",
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/apps/"+appName)
	})

	r.GET("/apps/:appName", func(c *gin.Context) {
		var latestURL string

		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		if !UserExists(etcd, username) {
			c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("User %s does not exist.", username),
			})

			return
		}

		appName := c.Param("appName")

		if !AppExists(etcd, username, appName) {
			c.HTML(http.StatusNotFound, "apps.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("Application %s does not exist.", appName),
			})

			return
		}

		urls, err := AppURLs(etcd, config.URIScheme, config.BaseDomain, username, appName)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": "Failed to list app URLs.",
			})

			return
		}

		envs, err := EnvironmentVariables(etcd, username, appName)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": "Failed to list environment variables.",
			})

			return
		}

		buildArgs, err := BuildArgs(etcd, username, appName)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": "Failed to list build args.",
			})

			return
		}

		if len(urls) > 0 {
			latestURL = LatestAppURLOfUser(config.URIScheme, config.BaseDomain, username, appName)
		}

		c.HTML(http.StatusOK, "app.tmpl", gin.H{
			"error":      false,
			"app":        appName,
			"avater_url": GetAvaterURL(etcd, username),
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
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := AddBuildArg(etcd, username, appName, key, value)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to add build arg.",
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/apps/"+appName)
	})

	// TODO: DELETE /apps/:appName/build-args
	r.POST("/apps/:appName/build-args/delete", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")

		err := DeleteBuildArg(etcd, username, appName, key)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to delete build arg.",
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/apps/"+appName)
	})

	r.POST("/apps/:appName/envs", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := AddEnvironmentVariable(etcd, username, appName, key, value)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to add environment variable.",
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/apps/"+appName)
	})

	// TODO: DELETE /apps/:appName/envs
	r.POST("/apps/:appName/envs/delete", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

		if username == "" {
			c.Redirect(http.StatusFound, "/")

			return
		}

		appName := c.Param("appName")
		key := c.PostForm("key")

		fmt.Println(key)

		err := DeleteEnvironmentVariable(etcd, username, appName, key)

		if err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to detele environment variable.",
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/apps/"+appName)
	})

	r.POST("/apps/:appName/envs/upload", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/")

			return
		}

		username := GetLoginUser(etcd, token.(string))

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

		if err = LoadDotenv(etcd, username, appName, dotenvFile); err != nil {
			errors.Fprint(os.Stderr, err)

			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": "Failed to load dotenv.",
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/apps/"+appName)
	})

	r.GET("/signin", func(c *gin.Context) {
		url := oauthConf.AuthCodeURL("hoge", oauth2.AccessTypeOnline)
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/signout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Delete("token")
		session.Save()

		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/oauth/callback", func(c *gin.Context) {
		code := c.Query("code")

		// TODO: compare to stored code in session

		token, err := oauthConf.Exchange(oauth2.NoContext, code)

		if err != nil {
			c.String(http.StatusBadRequest, "Error: %s", err)
			return
		}

		if !token.Valid() {
			c.String(http.StatusBadRequest, "%v", token)
			return
		}

		oauthClient := oauthConf.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token.AccessToken})
		client := github.NewClient(oauthClient)

		user, _, err := client.Users.Get("")

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
				_, err := UploadPublicKey(*user.Login, *key.Key)

				if err != nil {
					errors.Fprint(os.Stderr, err)

					c.String(http.StatusBadRequest, "Failed to register SSH public key.")
					return
				}
			} else {
				fmt.Printf("User: %s, Key: %s\n", *user.Login, *key.Key)
			}
		}

		if !UserExists(etcd, *user.Login) {
			if err := CreateUser(etcd, user); err != nil {
				errors.Fprint(os.Stderr, err)

				c.String(http.StatusBadRequest, "Failed to create user.")
				return
			}
		}

		if err := RegisterAccessToken(etcd, *user.Login, token.AccessToken); err != nil {
			errors.Fprint(os.Stderr, err)

			c.String(http.StatusBadRequest, "Failed to register access token.")
			return
		}

		session := sessions.Default(c)
		session.Set("token", token.AccessToken)
		session.Save()

		c.Redirect(http.StatusFound, "/apps")
	})

	r.Run()
}
