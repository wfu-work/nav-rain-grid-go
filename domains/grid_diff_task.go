package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

const (
	GridDiffTaskStatusSuccess = 1
	GridDiffTaskStatusFailed  = 2
)

const (
	GridDiffTaskNcStatusPending = 0
	GridDiffTaskNcStatusSuccess = 1
	GridDiffTaskNcStatusFailed  = 2
)

type GridDiffTask struct {
	domains.BaseDataEntity
	GridGuid         string  `json:"gridGuid" gorm:"index;comment:格网guid"`
	GridName         string  `json:"gridName" gorm:"comment:格网名称"`
	GridIdentifier   string  `json:"gridIdentifier" gorm:"index;comment:格网标识"`
	CoordinateSystem string  `json:"coordinateSystem" gorm:"comment:坐标系"`
	BaseTime         int64   `json:"baseTime" gorm:"index;comment:预测基准整点时间"`
	Resolution       float64 `json:"resolution" gorm:"comment:格网分辨率"`
	PointCount       int     `json:"pointCount" gorm:"comment:差分点数量"`
	Status           int     `json:"status" gorm:"comment:计算状态"`
	ErrorMsg         string  `json:"errorMsg" gorm:"comment:错误信息"`
	NcFilePath       string  `json:"ncFilePath" gorm:"comment:NC文件路径"`
	NcFileName       string  `json:"ncFileName" gorm:"comment:NC文件名"`
	NcFileSize       int64   `json:"ncFileSize" gorm:"comment:NC文件大小"`
	NcChecksum       string  `json:"ncChecksum" gorm:"comment:NC文件SHA256校验值"`
	NcStatus         int     `json:"ncStatus" gorm:"comment:NC文件生成状态"`
	NcErrorMsg       string  `json:"ncErrorMsg" gorm:"comment:NC文件生成错误信息"`
}

func (GridDiffTask) TableName() string {
	return "nav_rain_grid_diff_task"
}

func (s GridDiffTask) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
