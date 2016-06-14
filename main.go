package main

import (
	"os"

	"github.com/dtan4/paus-frontend/controller"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

const (
	AppName = "paus"
)

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

	if err = etcd.Init(config); err != nil {
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
	sessionController := controller.NewSessionController(config, etcd, oauthConf)

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

	r.POST("/apps/:appName/envs/upload", envController.Upload)

	r.GET("/update-keys", sessionController.UpdateKeys)
	r.GET("/signin", sessionController.SignIn)
	r.GET("/signout", sessionController.SignOut)
	r.GET("/oauth/callback", sessionController.Callback)

	r.Run()
}
