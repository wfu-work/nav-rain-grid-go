package domains

import "github.com/wfu-work/nav-common-go-lib/domains"

const (
	VersionReleaseStatusDraft     = 0
	VersionReleaseStatusPublished = 1
	VersionReleaseStatusDisabled  = 2
)

type VersionRelease struct {
	domains.BaseDataEntity
	Version      string `json:"version" gorm:"column:version_no;index;comment:版本号"`
	Name         string `json:"name" gorm:"comment:版本名称"`
	AppName      string `json:"appName" gorm:"index;comment:应用名称"`
	Platform     string `json:"platform" gorm:"index;comment:平台"`
	Architecture string `json:"architecture" gorm:"index;comment:架构"`
	Description  string `json:"description" gorm:"comment:版本描述"`
	ReleaseNote  string `json:"releaseNote" gorm:"comment:发布说明"`
	FilePath     string `json:"filePath" gorm:"comment:版本文件路径"`
	FileName     string `json:"fileName" gorm:"comment:版本文件名"`
	FileSize     int64  `json:"fileSize" gorm:"comment:版本文件大小"`
	Checksum     string `json:"checksum" gorm:"comment:版本文件SHA256校验值"`
	Status       int    `json:"status" gorm:"index;comment:发布状态"`
	ReleaseTime  int64  `json:"releaseTime" gorm:"index;comment:发布时间"`
}

func (VersionRelease) TableName() string {
	return "nav_rain_version_release"
}

func (s VersionRelease) GetBaseData() domains.BaseDataEntity {
	return s.BaseDataEntity
}
