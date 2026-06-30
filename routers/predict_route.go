package routers

import (
	"github.com/gin-gonic/gin"
)

type PredictRouter struct{}

func (s *PredictRouter) InitPredictRouter(Router *gin.RouterGroup) {
	routerLogger := Router.Group("predict")
	router := Router.Group("predict")
	{
		routerLogger.DELETE("params", predictApi.DeleteByParams) // 根据参数删除气象预测
		routerLogger.DELETE(":guid", predictApi.DeleteByGuid)    // 删除气象预测
	}
	{
		router.GET(":guid", predictApi.GetByGuid) // 查询气象预测
		router.GET("list", predictApi.List)       // 分页获取气象预测列表
		router.GET("query", predictApi.Query)     // 获取气象预测列表
		router.GET("last", predictApi.Last)       // 获取最新气象预测
		router.GET("export", predictApi.Export)   // 导出气象预测数据
	}
}
