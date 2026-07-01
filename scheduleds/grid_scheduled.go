package scheduleds

import (
	"nav-rain-grid-go/services"
	"time"

	"github.com/robfig/cron/v3"
	commonScheduleds "github.com/wfu-work/nav-common-go-lib/scheduleds"
	"go.uber.org/zap"
)

const gridCalculateDelay = 5 * time.Minute

func RegisterGrid(timers commonScheduleds.Timer, options []cron.Option) {
	_, err := timers.AddTaskByFunc("GridRainCalculate", "@every 1h", func() {
		time.Sleep(gridCalculateDelay)
		results, err := services.GridCalculationServiceApp.CalculateEnabledGrids(time.Now())
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
			zap.Duration("delay", gridCalculateDelay),
		)
	}, "每小时延迟5分钟后基于启用格网配置计算1/12/24小时格网中心点预测降雨", options...)
	if err != nil {
		zap.L().Error("注册格网降雨差分计算任务失败", zap.Error(err))
	}
}
