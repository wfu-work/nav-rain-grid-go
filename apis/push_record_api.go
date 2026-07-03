package apis

import (
	"nav-rain-grid-go/domains"

	"github.com/gin-gonic/gin"
	domains2 "github.com/wfu-work/nav-common-go-lib/domains"
	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/response"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"go.uber.org/zap"
)

type PushRecordApi struct{}

// List 分页获取格网推送记录列表
// @Summary 分页获取格网推送记录列表
// @Description 分页获取格网推送记录列表
// @Tags 格网推送记录模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query domains.PageInfo true "页码, 每页大小"
// @Success 200 {object} response.Response{data=domains.PageResult,msg=string}
// @Router /push-record/list [get]
func (i PushRecordApi) List(c *gin.Context) {
	params := queryParams(c)
	list, total, err := pushRecordService.List(params)
	if err != nil {
		global.NAV_LOG.Error("获取格网推送记录列表失败", zap.Error(err))
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

// Query 查询格网推送记录列表
// @Summary 查询格网推送记录列表
// @Description 查询格网推送记录列表
// @Tags 格网推送记录模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any true "参数"
// @Success 200 {object} response.Response{data=[]domains.PushRecord,msg=string}
// @Router /push-record/query [get]
func (i PushRecordApi) Query(c *gin.Context) {
	result, err := pushRecordService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("查询格网推送记录列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

var _ = domains.PushRecord{}
