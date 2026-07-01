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
	GridDiffTaskRouter
	GridDiffPointRouter
	SystemMonitorRouter
}

var (
	configApi        = apis.ApiGroupApp.ConfigApi
	predictApi       = apis.ApiGroupApp.PredictApi
	deviceApi        = apis.ApiGroupApp.DeviceApi
	gridApi          = apis.ApiGroupApp.GridApi
	gridDiffTaskApi  = apis.ApiGroupApp.GridDiffTaskApi
	gridDiffPointApi = apis.ApiGroupApp.GridDiffPointApi
	systemMonitorApi = apis.ApiGroupApp.SystemMonitorApi
)

func AppRouterInit(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	RouterGroupApp.InitConfigRouter(privateGroup, publicGroup)
	RouterGroupApp.InitPredictRouter(privateGroup)
	RouterGroupApp.InitDeviceRouter(privateGroup)
	RouterGroupApp.InitGridRouter(privateGroup)
	RouterGroupApp.InitGridDiffTaskRouter(privateGroup)
	RouterGroupApp.InitGridDiffPointRouter(privateGroup)
	RouterGroupApp.InitSystemMonitorRouter(privateGroup)
}
