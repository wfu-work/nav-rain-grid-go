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
	GridRouter
	SystemMonitorRouter
}

var (
	configApi        = apis.ApiGroupApp.ConfigApi
	predictApi       = apis.ApiGroupApp.PredictApi
	deviceApi        = apis.ApiGroupApp.DeviceApi
	gridApi          = apis.ApiGroupApp.GridApi
	systemMonitorApi = apis.ApiGroupApp.SystemMonitorApi
)

func AppRouterInit(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	RouterGroupApp.InitConfigRouter(privateGroup, publicGroup)
	RouterGroupApp.InitPredictRouter(privateGroup)
	RouterGroupApp.InitDeviceRouter(privateGroup)
	RouterGroupApp.InitGridRouter(privateGroup)
	RouterGroupApp.InitSystemMonitorRouter(privateGroup)
}
