package routers

import (
	"github.com/gin-gonic/gin"
)

type GridRouter struct{}

func (s *GridRouter) InitGridRouter(Router *gin.RouterGroup) {
	routerLogger := Router.Group("grid")
	router := Router.Group("grid")
	{
		routerLogger.POST("", gridApi.Save)                // 新增或更新格网
		routerLogger.PUT(":guid", gridApi.Update)          // 更新格网
		routerLogger.DELETE(":guid", gridApi.DeleteByGuid) // 删除格网
	}
	{
		router.GET("list/all", gridApi.ListAll) // 获取全部格网列表
		router.GET("list", gridApi.List)        // 分页获取格网列表
		router.GET("query", gridApi.Query)      // 查询格网列表
		router.GET(":guid", gridApi.GetByGuid)  // 根据guid查询格网
	}
}
