package scheduleds

import (
	"nav-rain-grid-go/services"
	"time"

	"github.com/robfig/cron/v3"
	commonScheduleds "github.com/wfu-work/nav-common-go-lib/scheduleds"
	"go.uber.org/zap"
)

const gridCalculateCron = "0 5 * * * *"

func RegisterGrid(timers commonScheduleds.Timer, options []cron.Option) {
	_, err := timers.AddTaskByFunc("GridRainCalculate", gridCalculateCron, func() {
		now := time.Now()
		zap.L().Info("格网降雨差分计算开始", zap.Time("time", now))
		results, err := services.GridCalculationServiceApp.CalculateEnabledGrids(now)
		if err != nil {
			zap.L().Error("格网降雨差分计算失败", zap.Error(err))
			return
		}
		gridCount := len(results)
		pointCount := 0
		for _, result := range results {
			pointCount += len(result.Points)
		}
		zap.L().Info("格网降雨差分计算完成",
			zap.Int("gridCount", gridCount),
			zap.Int("pointCount", pointCount),
		)
	}, "每小时第5分钟基于启用格网配置计算1/12/24小时格网中心点预测降雨", options...)
	if err != nil {
		zap.L().Error("注册格网降雨差分计算任务失败", zap.Error(err))
	}
}
