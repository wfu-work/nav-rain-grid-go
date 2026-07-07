package domains

import (
	"log"
	"os"

	"github.com/wfu-work/nav-common-go-lib/global"
	"go.uber.org/zap"
)

func RegisterTables() {
	db := global.NAV_DB
	err := db.AutoMigrate(
		Config{},
		Predict{},
		Device{},
		Grid{},
		GridDiffTask{},
		GridDiffPoint{},
		PushRecord{},
		VersionRelease{},
	)
	if err != nil {
		global.NAV_LOG.Error("register business table failed", zap.Error(err))
		os.Exit(0)
	}
	log.Println("register business table success")
}
