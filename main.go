package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	config, err := LoadConfig()

	if err != nil {
		log.Fatal(err)
	}

	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	etcd, err := NewEtcd(config.EtcdEndpoint)

	if err != nil {
		log.Fatal(err)
	}

	if !etcd.HasKey("/paus") {
		if err = etcd.Mkdir("/paus"); err != nil {
			log.Fatal(err)
		}
	}

	if !etcd.HasKey("/paus/users") {
		if err = etcd.Mkdir("/paus/users"); err != nil {
			log.Fatal(err)
		}
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"alert":      false,
			"error":      false,
			"message":    "",
			"baseDomain": config.BaseDomain,
		})
	})

	r.GET("/users/:username", func(c *gin.Context) {
		username := c.Param("username")

		if !etcd.HasKey("/paus/users/" + username) {
			c.HTML(http.StatusNotFound, "user.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("User %s does not exist.", username),
			})

			return
		}

		apps, err := Apps(etcd, username)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "user.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		c.HTML(http.StatusOK, "user.tmpl", gin.H{
			"error": false,
			"user":  username,
			"apps":  apps,
		})
	})

	r.POST("/users/:username/apps", func(c *gin.Context) {
		username := c.Param("username")
		appName := c.PostForm("appName")

		err := CreateApp(etcd, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "users.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/apps/"+appName)
	})

	r.GET("/users/:username/apps/:appName", func(c *gin.Context) {
		var latestURL string

		username := c.Param("username")

		if !etcd.HasKey("/paus/users/" + username) {
			c.HTML(http.StatusNotFound, "user.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("User %s does not exist.", username),
			})

			return
		}

		appName := c.Param("appName")

		if !etcd.HasKey("/paus/users/" + username + "/" + appName) {
			c.HTML(http.StatusNotFound, "user.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("Application %s does not exist.", appName),
			})

			return
		}

		urls, err := AppURLs(etcd, config.URIScheme, config.BaseDomain, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		envs, err := EnvironmentVariables(etcd, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		buildArgs, err := BuildArgs(etcd, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		if len(urls) > 0 {
			latestURL = LatestAppURLOfUser(config.URIScheme, config.BaseDomain, username, appName)
		}

		c.HTML(http.StatusOK, "app.tmpl", gin.H{
			"error":     false,
			"user":      username,
			"app":       appName,
			"latestURL": latestURL,
			"urls":      urls,
			"buildArgs": buildArgs,
			"envs":      envs,
		})
	})

	r.POST("/users/:username/apps/:appName/build-args", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := AddBuildArg(etcd, username, appName, key, value)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/apps/"+appName)
	})

	r.POST("/users/:username/apps/:appName/envs", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := AddEnvironmentVariable(etcd, username, appName, key, value)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/apps/"+appName)
	})

	r.POST("/users/:username/apps/:appName/envs/upload", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")

		dotenvFile, _, err := c.Request.FormFile("dotenv")

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		if err = LoadDotenv(etcd, username, appName, dotenvFile); err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": fmt.Sprintf("Error: %s", err.Error()),
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/apps/"+appName)
	})

	r.POST("/submit", func(c *gin.Context) {
		username := c.PostForm("username")
		pubKey := c.PostForm("pubKey")

		if UserExists(etcd, username) {
			c.HTML(http.StatusConflict, "index.tmpl", gin.H{
				"alert":      true,
				"error":      true,
				"message":    fmt.Sprintf("User %s already exists.", username),
				"baseDomain": config.BaseDomain,
			})

			return
		}

		err := CreateUser(etcd, username)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.tmpl", gin.H{
				"alert":      true,
				"error":      true,
				"message":    fmt.Sprintf("Error: %s", err.Error()),
				"baseDomain": config.BaseDomain,
			})

			return
		}

		out, err := UploadPublicKey(username, pubKey)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.tmpl", gin.H{
				"alert":      true,
				"error":      true,
				"message":    fmt.Sprintf("Error: %s", err.Error()),
				"baseDomain": config.BaseDomain,
			})

			return
		}

		c.HTML(http.StatusCreated, "index.tmpl", gin.H{
			"alert":      true,
			"error":      false,
			"message":    fmt.Sprintf("Fingerprint: %s", out),
			"baseDomain": config.BaseDomain,
			"username":   username,
		})
	})

	r.Run()
}
