package main

import (
	"nav-rain-grid-go/inits"
)

// 这部分 @Tag 设置用于排序, 需要排序的接口请按照下面的格式添加
// swag init 对 @Tag 只会从入口文件解析, 默认 main.go
// 也可通过 --generalInfo flag 指定其他文件
// @Tag.Name 降雨格网系统
// @Tag.Description 降雨格网系统相关的接口列表

// @title                       PAN-MI Swagger API接口文档
// @version                     v1.0.0
// @description                 降雨格网系统接口文档
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @BasePath                    /
func main() {
	inits.Init()
}
