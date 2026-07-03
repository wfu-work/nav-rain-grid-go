package routers

import "github.com/gin-gonic/gin"

type PushRecordRouter struct{}

func (s *PushRecordRouter) InitPushRecordRouter(Router *gin.RouterGroup) {
	router := Router.Group("push-record")
	{
		router.GET("list", pushRecordApi.List)
		router.GET("query", pushRecordApi.Query)
	}
}
