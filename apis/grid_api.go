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

type GridApi struct{}

// Save 新增或更新格网
// @Summary 新增或更新格网
// @Description 新增或更新格网
// @Tags 格网模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body domains.Grid true "格网信息"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /grid [post]
func (i GridApi) Save(c *gin.Context) {
	var entity domains.Grid
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := gridService.SaveOrUpdate(entity); err != nil {
		global.NAV_LOG.Error("保存格网失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// Update 更新格网
// @Summary 更新格网
// @Description 更新格网
// @Tags 格网模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "格网guid"
// @Param data body domains.Grid true "格网信息"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /grid/{guid} [put]
func (i GridApi) Update(c *gin.Context) {
	var entity domains.Grid
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	entity.Guid = c.Param("guid")
	if err := gridService.SaveOrUpdate(entity); err != nil {
		global.NAV_LOG.Error("更新格网失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// DeleteByGuid 根据guid删除格网
// @Summary 根据guid删除格网
// @Description 根据guid删除格网
// @Tags 格网模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "格网guid"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /grid/{guid} [delete]
func (i GridApi) DeleteByGuid(c *gin.Context) {
	guid := c.Param("guid")
	err := gridService.DeleteByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("删除格网失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(true, c)
}

// GetByGuid 根据guid获取格网
// @Summary 根据guid获取格网
// @Description 根据guid获取格网
// @Tags 格网模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "格网guid"
// @Success 200 {object} response.Response{data=domains.Grid,msg=string}
// @Router /grid/{guid} [get]
func (i GridApi) GetByGuid(c *gin.Context) {
	guid := c.Param("guid")
	t, err := gridService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取格网失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(t, c)
}

// List 分页获取格网列表
// @Summary 分页获取格网列表
// @Description 分页获取格网列表
// @Tags 格网模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query domains.PageInfo true "页码, 每页大小"
// @Success 200 {object} response.Response{data=domains.PageResult,msg=string}
// @Router /grid/list [get]
func (i GridApi) List(c *gin.Context) {
	params := queryParams(c)
	list, total, err := gridService.List(params)
	if err != nil {
		global.NAV_LOG.Error("获取格网列表失败", zap.Error(err))
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

// ListAll 获取全部格网列表
// @Summary 获取全部格网列表
// @Description 获取全部格网列表
// @Tags 格网模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any false "查询参数"
// @Success 200 {object} response.Response{data=[]domains.Grid,msg=string}
// @Router /grid/list/all [get]
func (i GridApi) ListAll(c *gin.Context) {
	result, err := gridService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("获取全部格网列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// Query 查询格网列表
// @Summary 查询格网列表
// @Description 查询格网列表
// @Tags 格网模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any true "参数"
// @Success 200 {object} response.Response{data=[]domains.Grid,msg=string}
// @Router /grid/query [get]
func (i GridApi) Query(c *gin.Context) {
	result, err := gridService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("查询格网列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}
