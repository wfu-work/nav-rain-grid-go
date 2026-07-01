package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

const DefaultGridResolution = 0.01

const (
	GridStatusDisabled = 0
	GridStatusEnabled  = 1
)

type Grid struct {
	domains.BaseDataEntity
	Name        string  `json:"name" gorm:"index;comment:格网名称"`
	Sncodes     string  `json:"sncodes" gorm:"comment:设备号多选用英文逗号分割"`
	Resolution  float64 `json:"resolution" gorm:"default:0.01;comment:格网分辨率，默认0.01度（约1公里）"`
	MinDevice   int     `json:"minDevice" gorm:"default:3;comment:最少设备，默认3台"`
	MinDistance float64 `json:"minDistance" gorm:"default:3;comment:最小距离，默认3公里"`
	Status      int     `json:"status" gorm:"comment:启用/禁用"`
}

func (Grid) TableName() string {
	return "nav_rain_grid"
}

func (s Grid) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
