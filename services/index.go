package services

import (
	"github.com/wfu-work/nav-common-go-lib/domains"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	ConfigService
	PredictService
	DeviceService
	GridService
	GridCalculationService
	GridDiffTaskService
	GridDiffPointService
	SystemMonitorService
}

type HasBaseData interface {
	GetBaseData() domains.BaseDataEntity
}
