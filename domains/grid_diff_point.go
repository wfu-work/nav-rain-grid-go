package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

type GridDiffPoint struct {
	domains.BaseDataEntity
	TaskGuid            string  `json:"taskGuid" gorm:"index;comment:差分任务guid"`
	GridGuid            string  `json:"gridGuid" gorm:"index;comment:格网guid"`
	GridName            string  `json:"gridName" gorm:"comment:格网名称"`
	BaseTime            int64   `json:"baseTime" gorm:"index;comment:预测基准整点时间"`
	CenterLng           float64 `json:"centerLng" gorm:"comment:格网中心经度"`
	CenterLat           float64 `json:"centerLat" gorm:"comment:格网中心纬度"`
	PredictTime1H       int64   `json:"predictTime1h" gorm:"comment:1小时预测时间"`
	PredictRain1H       float64 `json:"predictRain1h" gorm:"comment:1小时预测降雨值"`
	PredictRainLevel1H  int     `json:"predictRainLevel1h" gorm:"comment:1小时预测降雨等级"`
	PredictTime12H      int64   `json:"predictTime12h" gorm:"comment:12小时预测时间"`
	PredictRain12H      float64 `json:"predictRain12h" gorm:"comment:12小时预测降雨值"`
	PredictRainLevel12H int     `json:"predictRainLevel12h" gorm:"comment:12小时预测降雨等级"`
	PredictTime24H      int64   `json:"predictTime24h" gorm:"comment:24小时预测时间"`
	PredictRain24H      float64 `json:"predictRain24h" gorm:"comment:24小时预测降雨值"`
	PredictRainLevel24H int     `json:"predictRainLevel24h" gorm:"comment:24小时预测降雨等级"`
}

func (GridDiffPoint) TableName() string {
	return "nav_rain_grid_diff_point"
}

func (s GridDiffPoint) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
