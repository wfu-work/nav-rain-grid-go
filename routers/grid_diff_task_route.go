package routers

import "github.com/gin-gonic/gin"

type GridDiffTaskRouter struct{}

func (s *GridDiffTaskRouter) InitGridDiffTaskRouter(Router *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	router := Router.Group("grid-diff-task")
	v1PublicRouter := publicGroup.Group("v1/grid-diff")
	{
		router.GET("list/all", gridDiffTaskApi.ListAll) // 获取全部格网差分任务列表
		router.GET("list", gridDiffTaskApi.List)        // 分页获取格网差分任务列表
		router.GET("query", gridDiffTaskApi.Query)      // 查询格网差分任务列表
		router.GET("nc/latest", gridDiffTaskApi.LatestNCLink)
		router.GET("nc/download/:guid", gridDiffTaskApi.DownloadNC)
		router.GET(":guid/nc/download", gridDiffTaskApi.DownloadNC)
		router.GET(":guid", gridDiffTaskApi.GetByGuid) // 根据guid查询格网差分任务
	}
	{
		v1PublicRouter.GET("nc/latest", gridDiffTaskApi.LatestNCLinkV1)
		v1PublicRouter.GET("nc/download/:guid", gridDiffTaskApi.DownloadNC)
	}
}
