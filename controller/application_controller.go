package controller

import (
	"github.com/dtan4/paus-frontend/model/user"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ApplicationController struct {
	config *server.Config
	etcd   *store.Etcd
}

func NewApplicationController(config *server.Config, etcd *store.Etcd) *ApplicationController {
	return &ApplicationController{
		config: config,
		etcd:   etcd,
	}
}

func (self *ApplicationController) CurrentUser(c *gin.Context) string {
	session := sessions.Default(c)
	token := session.Get("token")

	if token == nil {
		return ""
	}

	return user.GetLoginUser(self.etcd, token.(string))
}
