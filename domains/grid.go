package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

type Grid struct {
	domains.BaseDataEntity
	Name        string  `json:"name" gorm:"index;comment:格网名称"`
	Sncodes     string  `json:"sncodes" gorm:"comment:设备号多选用英文逗号分割"`
	Resolution  string  `json:"resolution" gorm:"comment:格网分辨率"`
	MinDevice   int     `json:"min_device" gorm:"default:3;comment:最少设备，默认3台"`
	MinDistance float64 `json:"min_distance" gorm:"default:3;comment:最小距离，默认3公里"`
	Status      int     `json:"status" gorm:"comment:启用/禁用"`
}

func (Grid) TableName() string {
	return "nav_rain_grid"
}

func (s Grid) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
