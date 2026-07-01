package routers

import "github.com/gin-gonic/gin"

type GridDiffPointRouter struct{}

func (s *GridDiffPointRouter) InitGridDiffPointRouter(Router *gin.RouterGroup) {
	router := Router.Group("grid-diff-point")
	{
		router.GET("list/all", gridDiffPointApi.ListAll) // 获取全部格网差分点列表
		router.GET("list", gridDiffPointApi.List)        // 分页获取格网差分点列表
		router.GET("query", gridDiffPointApi.Query)      // 查询格网差分点列表
		router.GET(":guid", gridDiffPointApi.GetByGuid)  // 根据guid查询格网差分点
	}
}
