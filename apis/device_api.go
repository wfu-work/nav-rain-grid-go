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

type DeviceApi struct{}

// Save 新增或更新设备
// @Summary 新增或更新设备
// @Description 新增或更新设备
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body domains.Device true "设备信息"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /device [post]
func (i DeviceApi) Save(c *gin.Context) {
	var entity domains.Device
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := deviceService.SaveOrUpdate(entity); err != nil {
		global.NAV_LOG.Error("保存设备失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// Update 更新设备
// @Summary 更新设备
// @Description 更新设备
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "设备guid"
// @Param data body domains.Device true "设备信息"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /device/{guid} [put]
func (i DeviceApi) Update(c *gin.Context) {
	var entity domains.Device
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	entity.Guid = c.Param("guid")
	if err := deviceService.SaveOrUpdate(entity); err != nil {
		global.NAV_LOG.Error("更新设备失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(true, c)
}

// DeleteByGuid 根据guid删除设备
// @Summary 根据guid删除设备
// @Description 根据guid删除设备
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "设备guid"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /device/{guid} [delete]
func (i DeviceApi) DeleteByGuid(c *gin.Context) {
	guid := c.Param("guid")
	err := deviceService.DeleteByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("删除设备失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(true, c)
}

// GetByGuid 根据guid获取设备
// @Summary 根据guid获取设备
// @Description 根据guid获取设备
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param guid path string true "设备guid"
// @Success 200 {object} response.Response{data=domains.Device,msg=string}
// @Router /device/{guid} [get]
func (i DeviceApi) GetByGuid(c *gin.Context) {
	guid := c.Param("guid")
	t, err := deviceService.GetByGuid(guid)
	if err != nil {
		global.NAV_LOG.Error("获取设备失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(t, c)
}

// GetBySncode 根据设备号获取设备
// @Summary 根据设备号获取设备
// @Description 根据设备号获取设备
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sncode path string true "设备号"
// @Success 200 {object} response.Response{data=domains.Device,msg=string}
// @Router /device/sncode/{sncode} [get]
func (i DeviceApi) GetBySncode(c *gin.Context) {
	sncode := c.Param("sncode")
	t, err := deviceService.GetBySncode(sncode)
	if err != nil {
		global.NAV_LOG.Error("获取设备失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(t, c)
}

// List 分页获取设备列表
// @Summary 分页获取设备列表
// @Description 分页获取设备列表
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query domains.PageInfo true "页码, 每页大小"
// @Success 200 {object} response.Response{data=domains.PageResult,msg=string}
// @Router /device/list [get]
func (i DeviceApi) List(c *gin.Context) {
	params := queryParams(c)
	list, total, err := deviceService.List(params)
	if err != nil {
		global.NAV_LOG.Error("获取设备列表失败", zap.Error(err))
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

// ListAll 获取全部设备列表
// @Summary 获取全部设备列表
// @Description 获取全部设备列表
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any false "查询参数"
// @Success 200 {object} response.Response{data=[]domains.Device,msg=string}
// @Router /device/list/all [get]
func (i DeviceApi) ListAll(c *gin.Context) {
	result, err := deviceService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("获取全部设备列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}

// Query 查询设备列表
// @Summary 查询设备列表
// @Description 查询设备列表
// @Tags 设备模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data query any true "参数"
// @Success 200 {object} response.Response{data=[]domains.Device,msg=string}
// @Router /device/query [get]
func (i DeviceApi) Query(c *gin.Context) {
	result, err := deviceService.Query(queryParams(c))
	if err != nil {
		global.NAV_LOG.Error("查询设备列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(result, c)
}
