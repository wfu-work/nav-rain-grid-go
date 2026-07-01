package services

import (
	"testing"

	"nav-rain-grid-go/domains"

	"github.com/wfu-work/nav-common-go-lib/global"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGridSaveOrUpdateDefaultsResolution(t *testing.T) {
	oldDB := global.NAV_DB
	defer func() {
		global.NAV_DB = oldDB
	}()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&domains.Grid{}); err != nil {
		t.Fatalf("migrate grid table: %v", err)
	}
	global.NAV_DB = db

	service := GridService{}
	if err := service.SaveOrUpdate(domains.Grid{Name: "测试格网"}); err != nil {
		t.Fatalf("save grid: %v", err)
	}

	var created domains.Grid
	if err := db.Where("name = ?", "测试格网").First(&created).Error; err != nil {
		t.Fatalf("query created grid: %v", err)
	}
	if created.Resolution != domains.DefaultGridResolution {
		t.Fatalf("unexpected created resolution: %v", created.Resolution)
	}

	update := domains.Grid{Name: "测试格网更新"}
	update.Guid = created.Guid
	if err := service.SaveOrUpdate(update); err != nil {
		t.Fatalf("update grid: %v", err)
	}

	var updated domains.Grid
	if err := db.Where("guid = ?", created.Guid).First(&updated).Error; err != nil {
		t.Fatalf("query updated grid: %v", err)
	}
	if updated.Resolution != domains.DefaultGridResolution {
		t.Fatalf("unexpected updated resolution: %v", updated.Resolution)
	}

	if err := service.SaveOrUpdate(domains.Grid{Name: "两公里格网", Resolution: 0.02}); err != nil {
		t.Fatalf("save two kilometer grid: %v", err)
	}
	var twoKilometer domains.Grid
	if err := db.Where("name = ?", "两公里格网").First(&twoKilometer).Error; err != nil {
		t.Fatalf("query two kilometer grid: %v", err)
	}
	if twoKilometer.Resolution != 0.02 {
		t.Fatalf("unexpected two kilometer resolution: %v", twoKilometer.Resolution)
	}
}
