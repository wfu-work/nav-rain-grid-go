package apis

import (
	"errors"
	"nav-rain-grid-go/domains"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	domains2 "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/response"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type VersionReleaseApi struct{}

type VersionReleaseLink struct {
	Release     domains.VersionRelease `json:"release"`
	DownloadUrl string                 `json:"downloadUrl"`
}

// Save 新增或更新版本发布
// @Summary 新增或更新版本发布
// @Description 新增或更新版本发布
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body domains.VersionRelease true "版本发布信息"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /version-release [post]
func (i VersionReleaseApi) Save(c *gin.Context) {
	var entity domains.VersionRelease
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := versionReleaseService.SaveOrUpdate(entity); err != nil {
		global.NAV_LOG.Error("保存版本发布失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// Update 更新版本发布
// @Summary 更新版本发布
// @Description 更新版本发布
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "版本发布guid"
// @Param data body domains.VersionRelease true "版本发布信息"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /version-release/{guid} [put]
func (i VersionReleaseApi) Update(c *gin.Context) {
	var entity domains.VersionRelease
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	entity.Guid = c.Param("guid")
	if err := versionReleaseService.SaveOrUpdate(entity); err != nil {
		global.NAV_LOG.Error("更新版本发布失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// Upload 上传版本文件
// @Summary 上传版本文件
// @Description 上传版本文件
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "版本文件"
// @Param version formData string true "版本号"
// @Success 200 {object} response.Response{data=domains.VersionRelease,msg=string}
// @Router /version-release/upload [post]
func (i VersionReleaseApi) Upload(c *gin.Context) {
	entity := versionReleaseFromForm(c)
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("版本文件不能为空", c)
		return
	}
	result, err := versionReleaseService.Upload(entity, file)
	if err != nil {
		global.NAV_LOG.Error("上传版本文件失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// UploadByGuid 上传指定版本文件
// @Summary 上传指定版本文件
// @Description 上传指定版本文件
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param guid path string true "版本发布guid"
// @Param file formData file true "版本文件"
// @Success 200 {object} response.Response{data=domains.VersionRelease,msg=string}
// @Router /version-release/{guid}/upload [post]
func (i VersionReleaseApi) UploadByGuid(c *gin.Context) {
	entity := versionReleaseFromForm(c)
	entity.Guid = c.Param("guid")
	if entity.Version == "" {
		if existing, err := versionReleaseService.GetByGuid(entity.Guid); err == nil {
			entity.Version = existing.Version
			entity.Name = existing.Name
			entity.AppName = existing.AppName
			entity.Platform = existing.Platform
			entity.Architecture = existing.Architecture
			entity.Description = existing.Description
			entity.ReleaseNote = existing.ReleaseNote
			entity.Status = existing.Status
			entity.ReleaseTime = existing.ReleaseTime
		}
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("版本文件不能为空", c)
		return
	}
	result, err := versionReleaseService.Upload(entity, file)
	if err != nil {
		global.NAV_LOG.Error("上传版本文件失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// DeleteByGuid 根据guid删除版本发布
// @Summary 根据guid删除版本发布
// @Description 根据guid删除版本发布
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "版本发布guid"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /version-release/{guid} [delete]
func (i VersionReleaseApi) DeleteByGuid(c *gin.Context) {
	guid := c.Param("guid")
	if err := versionReleaseService.DeleteByGuid(guid); err != nil {
		global.NAV_LOG.Error("删除版本发布失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// GetByGuid 根据guid获取版本发布
// @Summary 根据guid获取版本发布
// @Description 根据guid获取版本发布
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "版本发布guid"
// @Success 200 {object} response.Response{data=domains.VersionRelease,msg=string}
// @Router /version-release/{guid} [get]
func (i VersionReleaseApi) GetByGuid(c *gin.Context) {
	guid := c.Param("guid")
	t, err := versionReleaseService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取版本发布失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(t, c)
}

// Latest 获取最新已发布版本
// @Summary 获取最新已发布版本
// @Description 获取最新已发布版本
// @Tags 版本发布模块
// @Accept json
// @Produce json
// @Param data query any false "查询参数"
// @Success 200 {object} response.Response{data=VersionReleaseLink,msg=string}
// @Router /version-release/latest [get]
func (i VersionReleaseApi) Latest(c *gin.Context) {
	release, err := versionReleaseService.LatestPublished(queryParams(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.FailWithMessage("暂无已发布版本", c)
		return
	}
	if err != nil {
		global.NAV_LOG.Error("获取最新版本失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(VersionReleaseLink{
		Release:     release,
		DownloadUrl: buildVersionReleaseDownloadURL(c, release.Guid),
	}, c)
}

// Download 下载版本文件
// @Summary 下载版本文件
// @Description 下载版本文件
// @Tags 版本发布模块
// @Accept json
// @Produce application/octet-stream
// @Param guid path string true "版本发布guid"
// @Router /version-release/{guid}/download [get]
func (i VersionReleaseApi) Download(c *gin.Context) {
	guid := c.Param("guid")
	release, err := versionReleaseService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取版本发布失败", zap.Error(err))
		response.FailWithMessage("版本发布不存在", c)
		return
	}
	filePath := strings.TrimSpace(release.FilePath)
	if filePath == "" {
		response.FailWithMessage("版本文件路径为空", c)
		return
	}
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		global.NAV_LOG.Error("版本文件不存在", zap.String("filePath", filePath), zap.Error(err))
		response.FailWithMessage("版本文件不存在", c)
		return
	}

	fileName := strings.TrimSpace(release.FileName)
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Cache-Control", "no-cache")
	c.FileAttachment(filePath, fileName)
}

// List 分页获取版本发布列表
// @Summary 分页获取版本发布列表
// @Description 分页获取版本发布列表
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query domains.PageInfo true "页码, 每页大小"
// @Success 200 {object} response.Response{data=domains.PageResult,msg=string}
// @Router /version-release/list [get]
func (i VersionReleaseApi) List(c *gin.Context) {
	params := queryParams(c)
	list, total, err := versionReleaseService.List(params)
	if err != nil {
		global.NAV_LOG.Error("获取版本发布列表失败", zap.Error(err))
		response.Fail(nil, c)
		return
	}
	response.Ok(domains2.PageResult{
		Data:  list,
		Total: total,
		Page:  utils.Str2Int(params["page"]),
		Size:  utils.Str2Int(params["size"]),
	}, c)
}

// ListAll 获取全部版本发布列表
// @Summary 获取全部版本发布列表
// @Description 获取全部版本发布列表
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any false "查询参数"
// @Success 200 {object} response.Response{data=[]domains.VersionRelease,msg=string}
// @Router /version-release/list/all [get]
func (i VersionReleaseApi) ListAll(c *gin.Context) {
	result, err := versionReleaseService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("获取全部版本发布列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// Query 查询版本发布列表
// @Summary 查询版本发布列表
// @Description 查询版本发布列表
// @Tags 版本发布模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any true "参数"
// @Success 200 {object} response.Response{data=[]domains.VersionRelease,msg=string}
// @Router /version-release/query [get]
func (i VersionReleaseApi) Query(c *gin.Context) {
	result, err := versionReleaseService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("查询版本发布列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

func versionReleaseFromForm(c *gin.Context) domains.VersionRelease {
	entity := domains.VersionRelease{
		Version:      strings.TrimSpace(c.PostForm("version")),
		Name:         strings.TrimSpace(c.PostForm("name")),
		AppName:      strings.TrimSpace(c.PostForm("appName")),
		Platform:     strings.TrimSpace(c.PostForm("platform")),
		Architecture: strings.TrimSpace(c.PostForm("architecture")),
		Description:  strings.TrimSpace(c.PostForm("description")),
		ReleaseNote:  strings.TrimSpace(c.PostForm("releaseNote")),
	}
	entity.Guid = strings.TrimSpace(c.PostForm("guid"))
	if status := strings.TrimSpace(c.PostForm("status")); status != "" {
		if value, err := strconv.Atoi(status); err == nil {
			entity.Status = value
		}
	} else {
		entity.Status = domains.VersionReleaseStatusPublished
	}
	if releaseTime := strings.TrimSpace(c.PostForm("releaseTime")); releaseTime != "" {
		if value, err := strconv.ParseInt(releaseTime, 10, 64); err == nil {
			entity.ReleaseTime = value
		}
	}
	return entity
}

func buildVersionReleaseDownloadURL(c *gin.Context, guid string) string {
	currentPath := c.Request.URL.Path
	basePath := strings.TrimSuffix(currentPath, "/latest")
	if basePath == currentPath {
		basePath = "/version-release"
	}
	downloadPath := strings.TrimRight(basePath, "/") + "/" + url.PathEscape(guid) + "/download"
	return requestExternalBaseURL(c) + downloadPath
}
