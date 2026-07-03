package apis

import (
	"encoding/json"
	"errors"
	"nav-rain-grid-go/domains"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	domains2 "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/response"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GridDiffTaskApi struct{}

type GridDiffTaskNCLink struct {
	Key              string `json:"key"`
	GridGuid         string `json:"gridGuid"`
	GridName         string `json:"gridName"`
	GridIdentifier   string `json:"gridIdentifier"`
	CoordinateSystem string `json:"coordinateSystem"`
	TaskGuid         string `json:"taskGuid"`
	BaseTime         int64  `json:"baseTime"`
	NcFileName       string `json:"ncFileName"`
	NcFileSize       int64  `json:"ncFileSize"`
	NcChecksum       string `json:"ncChecksum"`
	DownloadUrl      string `json:"downloadUrl"`
}

type GridDiffTaskNCLinkV1 struct {
	BaseTime    int64  `json:"baseTime"`
	NcFileName  string `json:"ncFileName"`
	NcFileSize  int64  `json:"ncFileSize"`
	NcChecksum  string `json:"ncChecksum"`
	DownloadUrl string `json:"downloadUrl"`
}

// GetByGuid 根据guid获取格网差分任务
// @Summary 根据guid获取格网差分任务
// @Description 根据guid获取格网差分任务
// @Tags 格网差分任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "任务guid"
// @Success 200 {object} response.Response{data=domains.GridDiffTask,msg=string}
// @Router /grid-diff-task/{guid} [get]
func (i GridDiffTaskApi) GetByGuid(c *gin.Context) {
	guid := c.Param("guid")
	t, err := gridDiffTaskService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取格网差分任务失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(t, c)
}

// LatestNCLink 根据格网guid获取最新NC文件下载链接
// @Summary 根据格网guid获取最新NC文件下载链接
// @Description 根据格网guid获取最新NC文件下载链接
// @Tags 格网差分任务模块
// @Accept json
// @Produce json
// @Param key query string true "格网guid"
// @Success 200 {object} response.Response{data=GridDiffTaskNCLink,msg=string}
// @Router /grid-diff-task/nc/latest [get]
func (i GridDiffTaskApi) LatestNCLink(c *gin.Context) {
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		response.FailWithMessage("参数key不能为空", c)
		return
	}

	task, err := gridDiffTaskService.LatestSuccessNCByGridGuid(key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.FailWithMessage("暂无成功生成的NC文件", c)
		return
	}
	if err != nil {
		global.NAV_LOG.Error("获取最新NC文件失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.Ok(GridDiffTaskNCLink{
		Key:              key,
		GridGuid:         task.GridGuid,
		GridName:         task.GridName,
		GridIdentifier:   task.GridIdentifier,
		CoordinateSystem: task.CoordinateSystem,
		TaskGuid:         task.Guid,
		BaseTime:         task.BaseTime,
		NcFileName:       task.NcFileName,
		NcFileSize:       task.NcFileSize,
		NcChecksum:       task.NcChecksum,
		DownloadUrl:      buildGridDiffTaskNCDownloadURL(c, task.Guid),
	}, c)
}

// LatestNCLinkV1 根据格网guid获取最新NC文件下载链接
// @Summary 根据格网guid获取最新NC文件下载链接
// @Description 根据格网guid获取最新NC文件下载链接
// @Tags 格网差分任务模块
// @Accept json
// @Produce json
// @Param key query string true "格网guid"
// @Success 200 {object} GridDiffTaskNCLinkV1
// @Router /v1/grid-diff-task/nc/latest [get]
func (i GridDiffTaskApi) LatestNCLinkV1(c *gin.Context) {
	start := time.Now()
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		msg := "参数key不能为空"
		i.recordLatestNCLinkV1Push(c, start, key, nil, nil, domains.PushRecordStatusFailed, http.StatusOK, http.StatusBadRequest, msg)
		response.FailWithMessage(msg, c)
		return
	}

	task, err := gridDiffTaskService.LatestSuccessNCByGridGuid(key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		msg := "暂无成功生成的NC文件"
		i.recordLatestNCLinkV1Push(c, start, key, nil, nil, domains.PushRecordStatusFailed, http.StatusOK, http.StatusBadRequest, msg)
		response.FailWithMessage(msg, c)
		return
	}
	if err != nil {
		global.NAV_LOG.Error("获取最新NC文件失败", zap.Error(err))
		msg := err.Error()
		i.recordLatestNCLinkV1Push(c, start, key, nil, nil, domains.PushRecordStatusFailed, http.StatusOK, http.StatusBadRequest, msg)
		response.FailWithMessage(msg, c)
		return
	}

	result := GridDiffTaskNCLinkV1{
		BaseTime:    task.BaseTime,
		NcFileName:  task.NcFileName,
		NcFileSize:  task.NcFileSize,
		NcChecksum:  task.NcChecksum,
		DownloadUrl: buildGridDiffTaskNCDownloadURL(c, task.Guid),
	}
	i.recordLatestNCLinkV1Push(c, start, key, &task, &result, domains.PushRecordStatusSuccess, http.StatusOK, http.StatusOK, "")
	c.JSON(http.StatusOK, result)
}

func (i GridDiffTaskApi) recordLatestNCLinkV1Push(
	c *gin.Context,
	start time.Time,
	key string,
	task *domains.GridDiffTask,
	result *GridDiffTaskNCLinkV1,
	status int,
	httpStatus int,
	responseCode int,
	errorMsg string,
) {
	record := domains.PushRecord{
		Key:             key,
		Method:          c.Request.Method,
		Path:            c.Request.URL.Path,
		Query:           c.Request.URL.RawQuery,
		ClientIP:        c.ClientIP(),
		RemoteAddr:      c.Request.RemoteAddr,
		UserAgent:       c.GetHeader("User-Agent"),
		Referer:         c.GetHeader("Referer"),
		XForwardedFor:   c.GetHeader("X-Forwarded-For"),
		XForwardedHost:  c.GetHeader("X-Forwarded-Host"),
		XForwardedProto: c.GetHeader("X-Forwarded-Proto"),
		Status:          status,
		HttpStatus:      httpStatus,
		ResponseCode:    responseCode,
		ErrorMsg:        errorMsg,
		RequestTime:     start.UnixMilli(),
		CostMillis:      time.Since(start).Milliseconds(),
	}
	if task != nil {
		record.GridGuid = task.GridGuid
		record.GridName = task.GridName
		record.GridIdentifier = task.GridIdentifier
		record.TaskGuid = task.Guid
		record.BaseTime = task.BaseTime
		record.NcFileName = task.NcFileName
		record.NcFileSize = task.NcFileSize
		record.NcChecksum = task.NcChecksum
	}
	if result != nil {
		record.NcFileName = result.NcFileName
		record.NcFileSize = result.NcFileSize
		record.NcChecksum = result.NcChecksum
		record.DownloadUrl = result.DownloadUrl
		record.ResponseInfo = marshalJSONSilent(result)
	} else if errorMsg != "" {
		record.ResponseInfo = marshalJSONSilent(response.Response{
			Code: responseCode,
			Data: map[string]interface{}{},
			Msg:  errorMsg,
		})
	}
	if err := pushRecordService.Record(record); err != nil {
		global.NAV_LOG.Error("记录v1格网NC推送请求失败", zap.Error(err))
	}
}

func marshalJSONSilent(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

// DownloadNC 下载格网差分任务生成的NC文件
// @Summary 下载格网差分任务生成的NC文件
// @Description 下载格网差分任务生成的NC文件
// @Tags 格网差分任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce application/octet-stream
// @Param guid path string true "任务guid"
// @Router /grid-diff-task/{guid}/nc/download [get]
func (i GridDiffTaskApi) DownloadNC(c *gin.Context) {
	guid := c.Param("guid")
	task, err := gridDiffTaskService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取格网差分任务失败", zap.Error(err))
		response.FailWithMessage("格网任务不存在", c)
		return
	}
	if task.NcStatus != domains.GridDiffTaskNcStatusSuccess {
		response.FailWithMessage("NC文件尚未生成成功", c)
		return
	}
	filePath := strings.TrimSpace(task.NcFilePath)
	if filePath == "" {
		response.FailWithMessage("NC文件路径为空", c)
		return
	}
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		global.NAV_LOG.Error("NC文件不存在", zap.String("filePath", filePath), zap.Error(err))
		response.FailWithMessage("NC文件不存在", c)
		return
	}

	fileName := strings.TrimSpace(task.NcFileName)
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	c.Header("Content-Type", "application/x-netcdf")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Cache-Control", "no-cache")
	c.FileAttachment(filePath, fileName)
}

func buildGridDiffTaskNCDownloadURL(c *gin.Context, taskGuid string) string {
	currentPath := c.Request.URL.Path
	basePath := strings.TrimSuffix(currentPath, "/nc/latest")
	if basePath == currentPath {
		basePath = "/grid-diff-task"
	}
	downloadPath := strings.TrimRight(basePath, "/") + "/nc/download/" + url.PathEscape(taskGuid)
	return requestExternalBaseURL(c) + downloadPath
}

func requestExternalBaseURL(c *gin.Context) string {
	proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if proto != "" {
		proto = strings.TrimSpace(strings.Split(proto, ",")[0])
	}
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host != "" {
		host = strings.TrimSpace(strings.Split(host, ",")[0])
	}
	if host == "" {
		host = c.Request.Host
	}
	return proto + "://" + host
}

// List 分页获取格网差分任务列表
// @Summary 分页获取格网差分任务列表
// @Description 分页获取格网差分任务列表
// @Tags 格网差分任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query domains.PageInfo true "页码, 每页大小"
// @Success 200 {object} response.Response{data=domains.PageResult,msg=string}
// @Router /grid-diff-task/list [get]
func (i GridDiffTaskApi) List(c *gin.Context) {
	params := queryParams(c)
	list, total, err := gridDiffTaskService.List(params)
	if err != nil {
		global.NAV_LOG.Error("获取格网差分任务列表失败", zap.Error(err))
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

// ListAll 获取全部格网差分任务列表
// @Summary 获取全部格网差分任务列表
// @Description 获取全部格网差分任务列表
// @Tags 格网差分任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any false "查询参数"
// @Success 200 {object} response.Response{data=[]domains.GridDiffTask,msg=string}
// @Router /grid-diff-task/list/all [get]
func (i GridDiffTaskApi) ListAll(c *gin.Context) {
	result, err := gridDiffTaskService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("获取全部格网差分任务列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// Query 查询格网差分任务列表
// @Summary 查询格网差分任务列表
// @Description 查询格网差分任务列表
// @Tags 格网差分任务模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any true "参数"
// @Success 200 {object} response.Response{data=[]domains.GridDiffTask,msg=string}
// @Router /grid-diff-task/query [get]
func (i GridDiffTaskApi) Query(c *gin.Context) {
	result, err := gridDiffTaskService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("查询格网差分任务列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

var _ = domains.GridDiffTask{}
