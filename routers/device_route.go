package routers

import (
	"github.com/gin-gonic/gin"
)

type DeviceRouter struct{}

func (s *DeviceRouter) InitDeviceRouter(Router *gin.RouterGroup) {
	routerLogger := Router.Group("device")
	router := Router.Group("device")
	{
		routerLogger.POST("", deviceApi.Save)                // 新增或更新设备
		routerLogger.PUT(":guid", deviceApi.Update)          // 更新设备
		routerLogger.DELETE(":guid", deviceApi.DeleteByGuid) // 删除设备
	}
	{
		router.GET("list/all", deviceApi.ListAll)           // 获取全部设备列表
		router.GET("list", deviceApi.List)                  // 分页获取设备列表
		router.GET("query", deviceApi.Query)                // 查询设备列表
		router.GET("sncode/:sncode", deviceApi.GetBySncode) // 根据设备号查询设备
		router.GET(":guid", deviceApi.GetByGuid)            // 根据guid查询设备
	}
}
