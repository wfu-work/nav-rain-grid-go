package apis

import (
	"nav-rain-grid-go/mqtt"

	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/response"
)

type SystemMonitorApi struct{}

func (a SystemMonitorApi) Runtime(c *gin.Context) {
	item, err := systemMonitorService.Runtime()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(item, c)
}

func (a SystemMonitorApi) Mqtt(c *gin.Context) {
	response.Ok(mqtt.BrokerServiceApp.Status(), c)
}
