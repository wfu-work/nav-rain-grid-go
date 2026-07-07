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
	PushRecordService
	SystemMonitorService
	VersionReleaseService
}

type HasBaseData interface {
	GetBaseData() domains.BaseDataEntity
}
