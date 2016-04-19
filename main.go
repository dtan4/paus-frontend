package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

func apps(keysAPI client.KeysAPI, username string) ([]string, error) {
	resp, err := keysAPI.Get(context.Background(), "/paus/users/"+username+"/", &client.GetOptions{Sort: true})

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, node := range resp.Node.Nodes {
		appName := strings.Replace(node.Key, "/paus/users/"+username+"/", "", 1)
		result = append(result, appName)
	}

	return result, nil
}

func appURL(uriScheme, identifier, baseDomain string) string {
	return uriScheme + "://" + identifier + "." + baseDomain
}

func appURLs(keysAPI client.KeysAPI, uriScheme, baseDomain, username, appName string) ([]string, error) {
	resp, err := keysAPI.Get(context.Background(), "/paus/users/"+username+"/"+appName+"/revisions/", &client.GetOptions{Sort: true})

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, node := range resp.Node.Nodes {
		revision := strings.Replace(node.Key, "/paus/users/"+username+"/"+appName+"/revisions/", "", 1)
		identifier := username + "-" + appName + "-" + revision
		result = append(result, appURL(uriScheme, identifier, baseDomain))
	}

	return result, nil
}

func latestAppURLOfUser(uriScheme, baseDomain, username, appName string) string {
	identifier := username + "-" + appName
	return appURL(uriScheme, identifier, baseDomain)
}

func environmentVariables(keysAPI client.KeysAPI, username, appName string) (*map[string]string, error) {
	resp, err := keysAPI.Get(context.Background(), "/paus/users/"+username+"/"+appName+"/envs/", &client.GetOptions{Sort: true})

	if err != nil {
		return nil, err
	}

	result := map[string]string{}

	for _, node := range resp.Node.Nodes {
		key := strings.Replace(node.Key, "/paus/users/"+username+"/"+appName+"/envs/", "", 1)
		value := node.Value
		result[key] = value
	}

	return &result, nil
}

func addEnvironmentVariable(keysAPI client.KeysAPI, username, appName, key, value string) error {
	_, err := keysAPI.Set(context.Background(), "/paus/users/"+username+"/"+appName+"/envs/"+key, value, nil)

	return err
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
		apps, err := apps(keysAPI, username)

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
		urls, err := appURLs(keysAPI, uriScheme, baseDomain, username, appName)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})
		} else {
			envs, err := environmentVariables(keysAPI, username, appName)
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

		err := addEnvironmentVariable(keysAPI, username, appName, key, value)

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

		file, _, err := c.Request.FormFile("dotenv")

		if err != nil {
			c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})

			return
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			envKeyValue := strings.Split(scanner.Text(), "=")
			key, value := envKeyValue[0], strings.Join(envKeyValue[1:], "=")

			fmt.Printf("%s = %s\n", key, value)

			if key == "" {
				continue
			}

			err := addEnvironmentVariable(keysAPI, username, appName, key, value)

			if err != nil {
				c.HTML(http.StatusInternalServerError, "app.tmpl", gin.H{
					"alert":   true,
					"error":   true,
					"message": strings.Join([]string{"error: ", err.Error()}, ""),
				})

				return
			}
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
