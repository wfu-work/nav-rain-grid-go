package routers

import "github.com/gin-gonic/gin"

type VersionReleaseRouter struct{}

func (s *VersionReleaseRouter) InitVersionReleaseRouter(Router *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	routerLogger := Router.Group("version-release")
	router := Router.Group("version-release")
	publicRouter := publicGroup.Group("version-release")
	{
		routerLogger.POST("", versionReleaseApi.Save)                     // 新增或更新版本发布
		routerLogger.PUT(":guid", versionReleaseApi.Update)               // 更新版本发布
		routerLogger.POST("upload", versionReleaseApi.Upload)             // 上传版本文件
		routerLogger.POST(":guid/upload", versionReleaseApi.UploadByGuid) // 上传指定版本文件
		routerLogger.DELETE(":guid", versionReleaseApi.DeleteByGuid)      // 删除版本发布
	}
	{
		router.GET("list/all", versionReleaseApi.ListAll) // 获取全部版本发布列表
		router.GET("list", versionReleaseApi.List)        // 分页获取版本发布列表
		router.GET("query", versionReleaseApi.Query)      // 查询版本发布列表
		router.GET(":guid", versionReleaseApi.GetByGuid)  // 根据guid查询版本发布
	}
	{
		publicRouter.GET("latest", versionReleaseApi.Latest)           // 获取最新已发布版本
		publicRouter.GET(":guid/download", versionReleaseApi.Download) // 下载版本文件
	}
}
