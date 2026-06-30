package scheduleds

import (
	"time"

	"nav-rain-grid-go/services"

	"github.com/robfig/cron/v3"
	commonScheduleds "github.com/wfu-work/nav-common-go-lib/scheduleds"
	"go.uber.org/zap"
)

const deviceHeartbeatTimeout = 10 * time.Minute

func RegisterDeviceStatusCheck(timers commonScheduleds.Timer, options []cron.Option) {
	_, err := timers.AddTaskByFunc("DeviceStatusCheck", "@every 1m", func() {
		count, err := services.DeviceServiceApp.MarkOfflineExpired(deviceHeartbeatTimeout)
		if err != nil {
			zap.L().Error("设备离线状态检测失败", zap.Error(err))
			return
		}
		if count > 0 {
			zap.L().Info("设备离线状态检测完成", zap.Int64("offlineCount", count))
		}
	}, "每分钟检测超过10分钟未心跳设备并设置离线", options...)
	if err != nil {
		zap.L().Error("注册设备离线状态检测任务失败", zap.Error(err))
	}
}
