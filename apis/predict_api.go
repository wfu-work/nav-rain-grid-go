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

type PredictApi struct{}

// DeleteByGuid 根据guid删除预测数据
// @Summary 根据预测数据guid删除预测数据
// @Description 根据预测数据guid删除预测数据
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "预测数据guid"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /predict/{guid} [delete]
func (i PredictApi) DeleteByGuid(c *gin.Context) {
	guid := c.Param("guid")
	err := predictService.DeleteByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("删除失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(true, c)
}

// DeleteByParams 根据参数删除预测数据
// @Summary 根据预测数据参数删除预测数据
// @Description 根据预测数据参数删除预测数据
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "预测数据参数"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /predict/params [delete]
func (i PredictApi) DeleteByParams(c *gin.Context) {
	query := c.Request.URL.Query()
	params := make(map[string]string)
	for key, val := range query {
		if len(val) > 0 {
			params[key] = val[0]
		}
	}
	err := predictService.DeleteByParams(params)
	if err != nil {
		global.NAV_LOG.Error("删除失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(true, c)
}

// GetByGuid 获取预测数据信息
// @Summary 根据预测数据guid获取预测数据
// @Description 根据预测数据guid获取预测数据
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "预测数据guid"
// @Success 200 {object} response.Response{data=domains.Qx,msg=string}
// @Router /predict/{guid} [get]
func (i PredictApi) GetByGuid(c *gin.Context) {
	guid := c.Param("guid")
	t, err := predictService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(t, c)
}

// List 分页获取预测数据列表
// @Summary 分页获取预测数据列表
// @Description 分页获取预测数据列表
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data  query domains.PageInfo true  "页码, 每页大小"
// @Success 200  {object}  response.Response{data=domains.PageResult,msg=string}  "分页获取预测数据列表,返回包括列表,总数,页码,每页数量"
// @Router /predict/list [get]
func (i PredictApi) List(c *gin.Context) {
	query := c.Request.URL.Query()
	params := make(map[string]string)
	for key, val := range query {
		if len(val) > 0 {
			params[key] = val[0]
		}
	}
	list, total, err := predictService.List(params)
	if err != nil {
		global.NAV_LOG.Error("获取失败!", zap.Error(err))
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

// Query 查询预测数据
// @Summary 查询预测数据
// @Description 查询预测数据
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data  query any true  "参数"
// @Success 200  {object}  response.Response{data=domains.PageResult,msg=string}  "气象数据集合"
// @Router /predict/query [get]
func (i PredictApi) Query(c *gin.Context) {
	query := c.Request.URL.Query()
	params := make(map[string]string)
	for key, val := range query {
		if len(val) > 0 {
			params[key] = val[0]
		}
	}
	result, err := predictService.Query(params)
	if err != nil {
		global.NAV_LOG.Error("查询气象数据异常", zap.Error(err))
		response.Fail(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// Last 获取最新的预测数据
// @Summary 获取最新的预测数据
// @Description 获取最新的预测数据
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data  query any true  "参数"
// @Success 200  {object}  response.Response{data=domains.PageResult,msg=string}  "最新一条气象数据"
// @Router /predict/last [get]
func (i PredictApi) Last(c *gin.Context) {
	query := c.Request.URL.Query()
	params := make(map[string]string)
	for key, val := range query {
		if len(val) > 0 {
			params[key] = val[0]
		}
	}
	result, err := predictService.SafeLast(domains.Predict{})
	if err != nil {
		global.NAV_LOG.Error("获取最新一条数据失败", zap.Error(err))
		response.Fail(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// Export 导出预测数据
// @Summary 导出预测数据
// @Description 导出预测数据
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data  query any true  "参数"
// @Success 200  {object}  response.Response{data=domains.PageResult,msg=string}  "导出预测数据"
// @Router /predict/export [get]
func (i PredictApi) Export(c *gin.Context) {
	query := c.Request.URL.Query()
	params := make(map[string]string)
	for key, val := range query {
		if len(val) > 0 {
			params[key] = val[0]
		}
	}
	err := predictService.Export(params, c)
	if err != nil {
		global.NAV_LOG.Error("导出数据失败", zap.Error(err))
		response.Fail(err.Error(), c)
		return
	}
}
