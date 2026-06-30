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
	SystemMonitorApi
}

var (
	configService        = services.ServiceGroupApp.ConfigService
	predictService       = services.ServiceGroupApp.PredictService
	deviceService        = services.ServiceGroupApp.DeviceService
	gridService          = services.ServiceGroupApp.GridService
	systemMonitorService = services.ServiceGroupApp.SystemMonitorService
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
