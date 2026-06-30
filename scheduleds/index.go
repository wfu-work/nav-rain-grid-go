package scheduleds

import (
	"github.com/robfig/cron/v3"
	"github.com/wfu-work/nav-common-go-lib/scheduleds"
)

func Init(timers scheduleds.Timer, options []cron.Option) {
	RegisterDeviceStatusCheck(timers, options)
}
