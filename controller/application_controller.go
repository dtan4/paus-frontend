package controller

import (
	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/model/user"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ApplicationController struct {
	config *config.Config
}

// NewApplicationController creates new ApplicationController object
func NewApplicationController(config *config.Config) *ApplicationController {
	return &ApplicationController{
		config: config,
	}
}

// CurrentUser returns current login user
func (ac *ApplicationController) CurrentUser(c *gin.Context) string {
	session := sessions.Default(c)
	loginUser := session.Get("login")

	if loginUser == nil {
		return ""
	}

	if !user.Exists(loginUser.(string)) {
		session.Delete("login")
		return ""
	}

	return loginUser.(string)
}
