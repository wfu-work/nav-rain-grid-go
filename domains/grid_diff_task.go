package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

const (
	GridDiffTaskStatusSuccess = 1
	GridDiffTaskStatusFailed  = 2
)

type GridDiffTask struct {
	domains.BaseDataEntity
	GridGuid   string  `json:"gridGuid" gorm:"index;comment:格网guid"`
	GridName   string  `json:"gridName" gorm:"comment:格网名称"`
	BaseTime   int64   `json:"baseTime" gorm:"index;comment:预测基准整点时间"`
	Resolution float64 `json:"resolution" gorm:"comment:格网分辨率"`
	PointCount int     `json:"pointCount" gorm:"comment:差分点数量"`
	Status     int     `json:"status" gorm:"comment:计算状态"`
	ErrorMsg   string  `json:"errorMsg" gorm:"comment:错误信息"`
}

func (GridDiffTask) TableName() string {
	return "nav_rain_grid_diff_task"
}

func (s GridDiffTask) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
