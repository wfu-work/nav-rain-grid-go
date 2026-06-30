package apis

import (
	"nav-rain-grid-go/domains"

	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/response"
	"go.uber.org/zap"
)

type ConfigApi struct{}

// Save 创建高级配置
// @Summary 创建高级配置
// @Description 创建高级配置
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body domains.Config true "高级配置"
// @Success 200 {object} response.Response{data=bool,msg=string}
// @Router /configs [post]
func (i ConfigApi) Save(c *gin.Context) {
	var entity domains.Config
	err := c.ShouldBindJSON(&entity)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = configService.SaveOrUpdate(entity)
	if err != nil {
		global.NAV_LOG.Error("创建失败!", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(true, c)
}

// GetConfig 获取高级配置
// @Summary 获取高级配置
// @Description 获取高级配置
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=domains.Config,msg=string}
// @Router /configs [get]
func (i ConfigApi) GetConfig(c *gin.Context) {
	t, err := configService.GetConfig()
	if err != nil {
		global.NAV_LOG.Error("获取失败", zap.Error(err))
		response.Fail(false, c)
		return
	}
	response.Ok(t, c)
}

// GetVersion 获取软件版本
// @Summary 获取软件版本
// @Description 获取软件版本
// @Tags 边缘计算模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=string,msg=string}
// @Router /configs/version [get]
func (i ConfigApi) GetVersion(c *gin.Context) {
	t := configService.GetVersion()
	response.Ok(t, c)
}
