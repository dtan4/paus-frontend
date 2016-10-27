package server

import (
	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/controller"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

const (
	AppName = "paus"
)

// Run starts web server
func Run(config *config.Config) {
	oauthConf := oauth2.Config{
		ClientID:     config.GitHubClientID,
		ClientSecret: config.GitHubClientSecret,
		Scopes:       []string{"user", "read:public_key"},
		Endpoint:     githuboauth.Endpoint,
	}

	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	sessionStore := sessions.NewCookieStore([]byte(config.SecretKeyBase))

	r := gin.Default()
	r.Use(sessions.Sessions(AppName, sessionStore))
	r.Static("/assets", "assets")
	r.LoadHTMLGlob("templates/*")

	rootController := controller.NewRootController(config)
	appController := controller.NewAppController(config)
	argController := controller.NewArgController(config)
	envController := controller.NewEnvController(config)
	healthcheckController := controller.NewHealthcheckController(config)
	sessionController := controller.NewSessionController(config, oauthConf)

	r.GET("/", rootController.Index)

	r.GET("/signin", sessionController.SignIn)
	r.GET("/signout", sessionController.SignOut)
	r.GET("/oauth/callback", sessionController.Callback)
	r.GET("/update-keys", sessionController.UpdateKeys)

	r.GET("/apps", appController.Index)
	r.POST("/apps", appController.New)

	r.GET("/apps/:appName", appController.Get)

	r.POST("/apps/:appName/build-args", argController.New)
	r.POST("/apps/:appName/build-args/delete", argController.Delete) // TODO: DELETE /apps/:appName/build-args

	r.POST("/apps/:appName/envs", envController.New)
	r.POST("/apps/:appName/envs/delete", envController.Delete) // TODO: DELETE /apps/:appName/envs
	r.POST("/apps/:appName/envs/upload", envController.Upload)

	r.POST("/apps/:appName/healthcheck", healthcheckController.Update)

	r.Run()
}
