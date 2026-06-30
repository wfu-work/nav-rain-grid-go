package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

type Config struct {
	domains.BaseDataEntity
	Key   string `json:"key" gorm:"index;comment:键"`
	Value string `json:"value" gorm:"comment:值"`
}

func (Config) TableName() string {
	return "nav_sys_config"
}
