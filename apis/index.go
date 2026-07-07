package apis

import (
	"nav-rain-grid-go/services"

	"github.com/gin-gonic/gin"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	ConfigApi
	PredictApi
	DeviceApi
	GridApi
	GridDiffTaskApi
	GridDiffPointApi
	PushRecordApi
	SystemMonitorApi
	VersionReleaseApi
}

var (
	configService         = services.ServiceGroupApp.ConfigService
	predictService        = services.ServiceGroupApp.PredictService
	deviceService         = services.ServiceGroupApp.DeviceService
	gridService           = services.ServiceGroupApp.GridService
	gridDiffTaskService   = services.ServiceGroupApp.GridDiffTaskService
	gridDiffPointService  = services.ServiceGroupApp.GridDiffPointService
	pushRecordService     = services.ServiceGroupApp.PushRecordService
	systemMonitorService  = services.ServiceGroupApp.SystemMonitorService
	versionReleaseService = services.ServiceGroupApp.VersionReleaseService
)

func queryParams(c *gin.Context) map[string]string {
	query := c.Request.URL.Query()
	params := make(map[string]string)
	for key, val := range query {
		if len(val) > 0 {
			params[key] = val[0]
		}
	}
	return params
}
