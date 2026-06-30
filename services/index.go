package services

import (
	"github.com/wfu-work/nav-common-go-lib/domains"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	ConfigService
}

type HasBaseData interface {
	GetBaseData() domains.BaseDataEntity
}
