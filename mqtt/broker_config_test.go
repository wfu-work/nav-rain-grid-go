package mqtt

import (
	"nav-rain-grid-go/domains"
	"testing"

	"github.com/spf13/viper"
	"github.com/wfu-work/nav-common-go-lib/global"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoadConfigUsesDatabaseRainGridSettings(t *testing.T) {
	oldDB := global.NAV_DB
	oldViper := global.NAV_VIPER
	t.Cleanup(func() {
		global.NAV_DB = oldDB
		global.NAV_VIPER = oldViper
	})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&domains.Config{}); err != nil {
		t.Fatalf("migrate config: %v", err)
	}
	global.NAV_DB = db

	v := viper.New()
	v.Set("mqtt.enable", false)
	v.Set("mqtt.host", "127.0.0.1")
	v.Set("mqtt.port", 2883)
	global.NAV_VIPER = v

	cfg := loadConfig()
	if cfg.Enable {
		t.Fatalf("expected yaml mqtt enable false before db settings")
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 2883 {
		t.Fatalf("unexpected yaml mqtt config: %+v", cfg)
	}

	if err := db.Create(&domains.Config{
		Key:   RainGridSettingsKey,
		Value: `{"mqttEnable":true,"mqttPort":3883}`,
	}).Error; err != nil {
		t.Fatalf("create db settings: %v", err)
	}

	cfg = loadConfig()
	if !cfg.Enable {
		t.Fatalf("expected db mqtt enable true")
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 3883 {
		t.Fatalf("unexpected db mqtt config: %+v", cfg)
	}
}

func TestLoadConfigFallsBackToYamlWithoutDatabaseSettings(t *testing.T) {
	oldDB := global.NAV_DB
	oldViper := global.NAV_VIPER
	t.Cleanup(func() {
		global.NAV_DB = oldDB
		global.NAV_VIPER = oldViper
	})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&domains.Config{}); err != nil {
		t.Fatalf("migrate config: %v", err)
	}
	global.NAV_DB = db

	v := viper.New()
	v.Set("mqtt.enable", true)
	v.Set("mqtt.port", 4883)
	global.NAV_VIPER = v

	cfg := loadConfig()
	if !cfg.Enable || cfg.Port != 4883 {
		t.Fatalf("unexpected fallback mqtt config: %+v", cfg)
	}
}
