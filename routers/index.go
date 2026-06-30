package routers

import (
	"nav-rain-grid-go/apis"

	"github.com/gin-gonic/gin"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	ConfigRouter
}

var (
	configApi = apis.ApiGroupApp.ConfigApi
)

func AppRouterInit(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	RouterGroupApp.InitConfigRouter(privateGroup, publicGroup)
}
