package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/gin-gonic/gin"
)

func latestAppURLOfUser(uriScheme, baseDomain, username, appName string) string {
	identifier := username + "-" + appName

	return AppURL(uriScheme, identifier, baseDomain)
}

func main() {
	baseDomain := os.Getenv("BASE_DOMAIN")
	etcdEndpoint := os.Getenv("ETCD_ENDPOINT")
	uriScheme := os.Getenv("URI_SCHEME")

	if uriScheme == "" {
		uriScheme = "http"
	}

	config := client.Config{
		Endpoints:               []string{etcdEndpoint},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(config)

	if err != nil {
		log.Fatal(err)
	}

	keysAPI := client.NewKeysAPI(c)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"alert":      false,
			"error":      false,
			"message":    "",
			"baseDomain": baseDomain,
		})
	})

	r.GET("/users/:username", func(c *gin.Context) {
		username := c.Param("username")
		apps, err := Apps(keysAPI, username)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "user.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})
		} else {
			c.HTML(http.StatusOK, "user.tmpl", gin.H{
				"error": false,
				"user":  username,
				"apps":  apps,
			})
		}
	})

	r.GET("/users/:username/:appName", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")
		urls, err := AppURLs(keysAPI, uriScheme, baseDomain, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})
		} else {
			envs, err := EnvironmentVariables(keysAPI, username, appName)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
					"error":   true,
					"message": strings.Join([]string{"error: ", err.Error()}, ""),
				})
			} else {
				latestURL := latestAppURLOfUser(uriScheme, baseDomain, username, appName)

				c.HTML(http.StatusOK, "app.tmpl", gin.H{
					"error":     false,
					"user":      username,
					"app":       appName,
					"latestURL": latestURL,
					"urls":      urls,
					"envs":      envs,
				})
			}
		}
	})

	r.POST("/users/:username/:appName/envs", func(c *gin.Context) {
		appName := c.Param("appName")
		username := c.Param("username")
		key := c.PostForm("key")
		value := c.PostForm("value")

		err := AddEnvironmentVariable(keysAPI, username, appName, key, value)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})
		} else {
			c.Redirect(http.StatusMovedPermanently, "/users/"+username+"/"+appName)
		}
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

		if err = LoadDotenv(keysAPI, username, appName, dotenvFile); err != nil {
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
		} else {
			c.HTML(http.StatusCreated, "index.tmpl", gin.H{
				"alert":   true,
				"error":   false,
				"message": strings.Join([]string{"fingerprint: ", string(out)}, ""),
			})
		}
	})

	r.Run()
}
