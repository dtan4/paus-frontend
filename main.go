package main

import (
	"log"
	"net/http"
	"os/exec"
	"strings"

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
		apps, err := Apps(etcd, username)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "user.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		c.HTML(http.StatusOK, "user.tmpl", gin.H{
			"error": false,
			"user":  username,
			"apps":  apps,
		})
	})

	r.GET("/users/:username/:appName", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")
		urls, err := AppURLs(etcd, config.URIScheme, config.BaseDomain, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		envs, err := EnvironmentVariables(etcd, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		buildArgs, err := BuildArgs(etcd, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		latestURL := LatestAppURLOfUser(config.URIScheme, config.BaseDomain, username, appName)

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

	r.POST("/users/:username/:appName/build-args", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := AddBuildArg(etcd, username, appName, key, value)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/"+appName)
	})

	r.POST("/users/:username/:appName/envs", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := AddEnvironmentVariable(etcd, username, appName, key, value)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/"+appName)
	})

	r.POST("/users/:username/:appName/envs/upload", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")

		dotenvFile, _, err := c.Request.FormFile("dotenv")

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		if err = LoadDotenv(etcd, username, appName, dotenvFile); err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/"+appName)
	})

	r.POST("/submit", func(c *gin.Context) {
		username := c.PostForm("username")
		pubKey := c.PostForm("pubKey")

		// libcompose does not support `docker-compose run`...
		out, err := exec.Command("docker-compose", "-p", "paus", "run", "--rm", "gitreceive-upload-key", username, pubKey).Output()

		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		c.HTML(http.StatusCreated, "index.tmpl", gin.H{
			"alert":   true,
			"error":   false,
			"message": strings.Join([]string{"fingerprint: ", string(out)}, ""),
		})
	})

	r.Run()
}
