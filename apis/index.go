package apis

import (
	"nav-rain-grid-go/services"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	ConfigApi
}

var (
	configService = services.ServiceGroupApp.ConfigService
)
