package routers

import (
	"nav-rain-grid-go/apis"

	"github.com/gin-gonic/gin"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	ConfigRouter
	PredictRouter
	DeviceRouter
}

var (
	configApi  = apis.ApiGroupApp.ConfigApi
	predictApi = apis.ApiGroupApp.PredictApi
	deviceApi  = apis.ApiGroupApp.DeviceApi
)

func AppRouterInit(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	RouterGroupApp.InitConfigRouter(privateGroup, publicGroup)
	RouterGroupApp.InitPredictRouter(privateGroup)
	RouterGroupApp.InitDeviceRouter(privateGroup)
}
