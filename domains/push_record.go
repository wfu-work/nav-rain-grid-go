package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

const (
	PushRecordStatusSuccess = 1
	PushRecordStatusFailed  = 2
)

type PushRecord struct {
	domains.BaseDataEntity
	Key             string `json:"key" gorm:"index;comment:请求key，格网guid"`
	GridGuid        string `json:"gridGuid" gorm:"index;comment:格网guid"`
	GridName        string `json:"gridName" gorm:"comment:格网名称"`
	GridIdentifier  string `json:"gridIdentifier" gorm:"index;comment:格网标识"`
	TaskGuid        string `json:"taskGuid" gorm:"index;comment:格网任务guid"`
	BaseTime        int64  `json:"baseTime" gorm:"index;comment:预测基准整点时间"`
	Method          string `json:"method" gorm:"comment:请求方法"`
	Path            string `json:"path" gorm:"index;comment:请求路径"`
	Query           string `json:"query" gorm:"type:text;comment:请求参数"`
	ClientIP        string `json:"clientIp" gorm:"index;comment:客户端IP"`
	RemoteAddr      string `json:"remoteAddr" gorm:"comment:远端地址"`
	UserAgent       string `json:"userAgent" gorm:"type:text;comment:User-Agent"`
	Referer         string `json:"referer" gorm:"type:text;comment:Referer"`
	XForwardedFor   string `json:"xForwardedFor" gorm:"type:text;comment:X-Forwarded-For"`
	XForwardedHost  string `json:"xForwardedHost" gorm:"comment:X-Forwarded-Host"`
	XForwardedProto string `json:"xForwardedProto" gorm:"comment:X-Forwarded-Proto"`
	Status          int    `json:"status" gorm:"index;comment:请求状态，1成功，2失败"`
	HttpStatus      int    `json:"httpStatus" gorm:"comment:HTTP状态码"`
	ResponseCode    int    `json:"responseCode" gorm:"comment:业务响应码"`
	ErrorMsg        string `json:"errorMsg" gorm:"type:text;comment:错误信息"`
	ResponseInfo    string `json:"responseInfo" gorm:"type:text;comment:返回信息JSON"`
	NcFileName      string `json:"ncFileName" gorm:"comment:NC文件名"`
	NcFileSize      int64  `json:"ncFileSize" gorm:"comment:NC文件大小"`
	NcChecksum      string `json:"ncChecksum" gorm:"comment:NC文件SHA256校验值"`
	DownloadUrl     string `json:"downloadUrl" gorm:"type:text;comment:下载链接"`
	RequestTime     int64  `json:"requestTime" gorm:"index;comment:请求时间"`
	CostMillis      int64  `json:"costMillis" gorm:"comment:处理耗时毫秒"`
}

func (PushRecord) TableName() string {
	return "nav_rain_push_record"
}

func (s PushRecord) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
