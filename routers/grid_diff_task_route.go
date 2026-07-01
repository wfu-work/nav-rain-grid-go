package routers

import "github.com/gin-gonic/gin"

type GridDiffTaskRouter struct{}

func (s *GridDiffTaskRouter) InitGridDiffTaskRouter(Router *gin.RouterGroup) {
	router := Router.Group("grid-diff-task")
	{
		router.GET("list/all", gridDiffTaskApi.ListAll) // 获取全部格网差分任务列表
		router.GET("list", gridDiffTaskApi.List)        // 分页获取格网差分任务列表
		router.GET("query", gridDiffTaskApi.Query)      // 查询格网差分任务列表
		router.GET(":guid", gridDiffTaskApi.GetByGuid)  // 根据guid查询格网差分任务
	}
}
