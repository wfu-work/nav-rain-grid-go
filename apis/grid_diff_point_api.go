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

type GridDiffPointApi struct{}

// GetByGuid 根据guid获取格网差分点
// @Summary 根据guid获取格网差分点
// @Description 根据guid获取格网差分点
// @Tags 格网差分点模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "点位guid"
// @Success 200 {object} response.Response{data=domains.GridDiffPoint,msg=string}
// @Router /grid-diff-point/{guid} [get]
func (i GridDiffPointApi) GetByGuid(c *gin.Context) {
	guid := c.Param("guid")
	t, err := gridDiffPointService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取格网差分点失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(t, c)
}

// List 分页获取格网差分点列表
// @Summary 分页获取格网差分点列表
// @Description 分页获取格网差分点列表
// @Tags 格网差分点模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query domains.PageInfo true "页码, 每页大小"
// @Success 200 {object} response.Response{data=domains.PageResult,msg=string}
// @Router /grid-diff-point/list [get]
func (i GridDiffPointApi) List(c *gin.Context) {
	params := queryParams(c)
	list, total, err := gridDiffPointService.List(params)
	if err != nil {
		global.NAV_LOG.Error("获取格网差分点列表失败", zap.Error(err))
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

// ListAll 获取全部格网差分点列表
// @Summary 获取全部格网差分点列表
// @Description 获取全部格网差分点列表
// @Tags 格网差分点模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any false "查询参数"
// @Success 200 {object} response.Response{data=[]domains.GridDiffPoint,msg=string}
// @Router /grid-diff-point/list/all [get]
func (i GridDiffPointApi) ListAll(c *gin.Context) {
	result, err := gridDiffPointService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("获取全部格网差分点列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// Query 查询格网差分点列表
// @Summary 查询格网差分点列表
// @Description 查询格网差分点列表
// @Tags 格网差分点模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any true "参数"
// @Success 200 {object} response.Response{data=[]domains.GridDiffPoint,msg=string}
// @Router /grid-diff-point/query [get]
func (i GridDiffPointApi) Query(c *gin.Context) {
	result, err := gridDiffPointService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("查询格网差分点列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

var _ = domains.GridDiffPoint{}
