package inits

import (
	_ "embed"
	"fmt"
	"nav-rain-grid-go/domains"
	"nav-rain-grid-go/mqtt"
	"nav-rain-grid-go/routers"
	scheduleds2 "nav-rain-grid-go/scheduleds"
	"nav-rain-grid-go/utils"
	"nav-rain-grid-go/webs"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/wfu-work/nav-common-go-lib/inits"
	"github.com/wfu-work/nav-common-go-lib/scheduleds"
)

//go:embed config.yaml
var defaultConfig []byte

func Init() {
	if err := utils.NewDefaultConfigManager(defaultConfig).Ensure(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "prepare config failed: %v\n", err)
		os.Exit(1)
	}
	sysInit := inits.SysInit{}
	sysInit.OnTableInit(func() {
		domains.RegisterTables()
	})
	sysInit.OnRouterInit(func(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
		routers.AppRouterInit(publicGroup, privateGroup)
	})
	sysInit.OnOtherInit(func() {
		mqtt.InitMqtt()
	})
	sysInit.OnScheInit(func(timers scheduleds.Timer, options []cron.Option) {
		scheduleds2.Init(timers, options)
	})
	sysInit.OnWebInit(func(router *gin.Engine) {
		_ = webs.InitStatic(router)
	})
	sysInit.OnShutInit(func() {
		mqtt.BrokerServiceApp.Stop()
	})
	sysInit.Init()
}
