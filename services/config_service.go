package services

import (
	"log"
	"nav-rain-grid-go/domains"

	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/services"
)

type ConfigService struct {
	services.CrudService[domains.Config]
}

var ConfigServiceApp = new(ConfigService)

const (
	Version = "1.0.0"
)

func init() {
	log.Println("当前软件版本：" + Version)
}

func (s ConfigService) GetVersion() string {
	return Version
}

func (s ConfigService) SaveOrUpdate(config domains.Config) error {
	bean, _ := s.SafeFirst(domains.Config{})
	if bean != nil {
		config.Guid = bean.Guid
		return global.NAV_DB.Model(&domains.Config{}).Where("guid = ?", config.Guid).
			Select("*").
			Updates(config).Error
	}
	return s.Create(config)
}

func (s ConfigService) GetConfig() (*domains.Config, error) {
	return s.SafeFirst(domains.Config{})
}
