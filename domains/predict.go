package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

type Predict struct {
	domains.BaseDataEntity
	BaseTime         int64   `json:"baseTime" gorm:"index;comment:基准时间"`
	Time             int64   `json:"time" gorm:"index;comment:时间"`
	Sncode           string  `json:"sncode" gorm:"size:50;comment:设备号"`
	PredictRain      float64 `json:"predictRain" description:"预测雨量"`
	PredictRainLevel int     `json:"predictRainLevel" description:"预测雨量分类"`
	Type             int     `json:"type" description:"预测雨量类型"`
}

type PredictGroup struct {
	BaseTime    int64     `json:"baseTime"`
	PredictList []Predict `json:"predictList"`
}

func (Predict) TableName() string {
	return "nav_rain_predict"
}

func (s Predict) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
