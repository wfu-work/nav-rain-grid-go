package routers

import (
	"github.com/gin-gonic/gin"
)

type ConfigRouter struct{}

func (s *ConfigRouter) InitConfigRouter(Router *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	routerLogger := Router.Group("configs")
	publicRouter := publicGroup.Group("configs")
	{
		routerLogger.POST("", configApi.Save) // 新增高级配置
	}
	{
		publicRouter.GET("", configApi.GetConfig) // 查询高级配置
		publicRouter.GET("version", configApi.GetVersion)
		publicRouter.GET("current-version", configApi.CurrentVersion)
	}
}
