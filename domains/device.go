package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

type Device struct {
	domains.BaseDataEntity
	Sncode   string   `json:"sncode" gorm:"size:50;uniqueIndex;comment:设备号"`
	Alias    string   `json:"alias" gorm:"comment:别名"`
	Type     string   `json:"type" gorm:"comment:设备类型"`
	Lat      *float64 `json:"lat" gorm:"comment:纬度"`
	Lng      *float64 `json:"lng" gorm:"comment:经度"`
	Alt      *float64 `json:"alt" gorm:"comment:高程"`
	Gsw      *bool    `json:"gsw" gorm:"comment:是否开启北斗水位计"`
	Rain     *bool    `json:"rain" gorm:"comment:是否开启降雨预测"`
	Status   int      `json:"status" gorm:"comment:在线状态"`
	LastTime int64    `json:"last_time" gorm:"comment:最后在线时间"`
}

func (Device) TableName() string {
	return "nav_rain_device"
}

func (s Device) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
