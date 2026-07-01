package services

import (
	"math"
	"testing"
	"time"

	"nav-rain-grid-go/domains"

	"github.com/wfu-work/nav-common-go-lib/global"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGridCalculationServiceCalculateEnabledGrids(t *testing.T) {
	oldDB := global.NAV_DB
	defer func() {
		global.NAV_DB = oldDB
	}()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&domains.Grid{}, &domains.Device{}, &domains.Predict{}, &domains.GridDiffTask{}, &domains.GridDiffPoint{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	global.NAV_DB = db

	latA, lngA := 30.000, 114.000
	latB, lngB := 30.000, 114.010
	latC, lngC := 30.010, 114.000
	devices := []domains.Device{
		{Sncode: "A", Lat: &latA, Lng: &lngA},
		{Sncode: "B", Lat: &latB, Lng: &lngB},
		{Sncode: "C", Lat: &latC, Lng: &lngC},
	}
	for _, device := range devices {
		if err := db.Create(&device).Error; err != nil {
			t.Fatalf("create device: %v", err)
		}
	}

	grid := domains.Grid{
		Name:       "测试格网",
		Sncodes:    "A,B,C",
		Resolution: 0.01,
		Status:     domains.GridStatusEnabled,
	}
	if err := db.Create(&grid).Error; err != nil {
		t.Fatalf("create grid: %v", err)
	}

	now := time.Date(2026, 7, 1, 10, 30, 0, 0, time.Local)
	baseTime := alignGridBaseTime(now)
	hourMillis := int64(time.Hour / time.Millisecond)
	predictions := []domains.Predict{
		{Sncode: "A", Type: 1, BaseTime: baseTime, Time: baseTime + hourMillis, PredictRain: 10},
		{Sncode: "B", Type: 1, BaseTime: baseTime, Time: baseTime + hourMillis, PredictRain: 20},
		{Sncode: "C", Type: 1, BaseTime: baseTime, Time: baseTime + hourMillis, PredictRain: 30},
		{Sncode: "A", Type: 12, BaseTime: baseTime, Time: baseTime + 12*hourMillis, PredictRain: 1},
		{Sncode: "B", Type: 12, BaseTime: baseTime, Time: baseTime + 12*hourMillis, PredictRain: 2},
		{Sncode: "C", Type: 12, BaseTime: baseTime, Time: baseTime + 12*hourMillis, PredictRain: 3},
		{Sncode: "A", Type: 24, BaseTime: baseTime, Time: baseTime + 24*hourMillis, PredictRain: 4},
		{Sncode: "B", Type: 24, BaseTime: baseTime, Time: baseTime + 24*hourMillis, PredictRain: 5},
		{Sncode: "C", Type: 24, BaseTime: baseTime, Time: baseTime + 24*hourMillis, PredictRain: 6},
	}
	for _, prediction := range predictions {
		if err := db.Create(&prediction).Error; err != nil {
			t.Fatalf("create prediction: %v", err)
		}
	}

	results, err := GridCalculationService{}.CalculateEnabledGrids(now)
	if err != nil {
		t.Fatalf("calculate grids: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("unexpected result count: %d", len(results))
	}
	if len(results[0].Points) == 0 {
		t.Fatal("expected grid points")
	}

	point := results[0].Points[0]
	if point.Forecast1H == nil || point.Forecast12H == nil || point.Forecast24H == nil {
		t.Fatalf("expected 1/12/24 hour forecasts: %#v", point)
	}
	if len(point.Forecast1H.Devices) != 3 {
		t.Fatalf("expected three weighted devices, got %d", len(point.Forecast1H.Devices))
	}
	if point.Forecast1H.PredictRain <= 0 {
		t.Fatalf("unexpected forecast rain: %v", point.Forecast1H.PredictRain)
	}

	var task domains.GridDiffTask
	if err := db.Where("grid_guid = ? AND base_time = ?", results[0].GridGuid, results[0].BaseTime).First(&task).Error; err != nil {
		t.Fatalf("query grid diff task: %v", err)
	}
	if task.PointCount != len(results[0].Points) {
		t.Fatalf("unexpected task point count: got %d, want %d", task.PointCount, len(results[0].Points))
	}
	var pointCount int64
	if err := db.Model(&domains.GridDiffPoint{}).Where("task_guid = ?", task.Guid).Count(&pointCount).Error; err != nil {
		t.Fatalf("count grid diff points: %v", err)
	}
	if pointCount != int64(len(results[0].Points)) {
		t.Fatalf("unexpected saved point count: got %d, want %d", pointCount, len(results[0].Points))
	}

	if _, err := (GridCalculationService{}).CalculateEnabledGrids(now); err != nil {
		t.Fatalf("recalculate grids: %v", err)
	}
	var taskCount int64
	if err := db.Model(&domains.GridDiffTask{}).Where("grid_guid = ? AND base_time = ?", results[0].GridGuid, results[0].BaseTime).Count(&taskCount).Error; err != nil {
		t.Fatalf("count grid diff tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("same grid and baseTime should keep one task, got %d", taskCount)
	}
	if err := db.Model(&domains.GridDiffPoint{}).Where("task_guid = ?", task.Guid).Count(&pointCount).Error; err != nil {
		t.Fatalf("count recalculated grid diff points: %v", err)
	}
	if pointCount != int64(len(results[0].Points)) {
		t.Fatalf("recalculate should replace points: got %d, want %d", pointCount, len(results[0].Points))
	}
}

func TestGridCalculationServiceDoesNotFallbackToPreviousBaseTime(t *testing.T) {
	oldDB := global.NAV_DB
	defer func() {
		global.NAV_DB = oldDB
	}()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&domains.Grid{}, &domains.Device{}, &domains.Predict{}, &domains.GridDiffTask{}, &domains.GridDiffPoint{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	global.NAV_DB = db

	latA, lngA := 30.000, 114.000
	latB, lngB := 30.000, 114.010
	latC, lngC := 30.010, 114.000
	for _, device := range []domains.Device{
		{Sncode: "A", Lat: &latA, Lng: &lngA},
		{Sncode: "B", Lat: &latB, Lng: &lngB},
		{Sncode: "C", Lat: &latC, Lng: &lngC},
	} {
		if err := db.Create(&device).Error; err != nil {
			t.Fatalf("create device: %v", err)
		}
	}

	if err := db.Create(&domains.Grid{
		Name:       "测试格网",
		Sncodes:    "A,B,C",
		Resolution: 0.01,
		Status:     domains.GridStatusEnabled,
	}).Error; err != nil {
		t.Fatalf("create grid: %v", err)
	}

	now := time.Date(2026, 7, 1, 10, 30, 0, 0, time.Local)
	previousBaseTime := alignGridBaseTime(now.Add(-time.Hour))
	hourMillis := int64(time.Hour / time.Millisecond)
	for _, sncode := range []string{"A", "B", "C"} {
		for _, forecastHour := range []int{1, 12, 24} {
			if err := db.Create(&domains.Predict{
				Sncode:      sncode,
				Type:        forecastHour,
				BaseTime:    previousBaseTime,
				Time:        previousBaseTime + int64(forecastHour)*hourMillis,
				PredictRain: 99,
			}).Error; err != nil {
				t.Fatalf("create previous prediction: %v", err)
			}
		}
	}

	results, err := GridCalculationService{}.CalculateEnabledGrids(now)
	if err != nil {
		t.Fatalf("calculate grids: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("previous baseTime predictions should not be used: %#v", results)
	}
	var taskCount int64
	if err := db.Model(&domains.GridDiffTask{}).Count(&taskCount).Error; err != nil {
		t.Fatalf("count grid diff tasks: %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("no complete result should not save task, got %d", taskCount)
	}
}

func TestInterpolateGridForecastUsesNearestThreeDistanceWeights(t *testing.T) {
	center := gridCenter{lng: 0, lat: 0}
	devices := []gridDevicePredict{
		testGridDevicePredict("A", 1, 0, 10),
		testGridDevicePredict("B", 2, 0, 20),
		testGridDevicePredict("C", 3, 0, 30),
		testGridDevicePredict("D", 100, 0, 1000),
	}

	forecast := interpolateGridForecast(center, devices, 1)
	if forecast == nil {
		t.Fatal("expected forecast")
	}

	weights := []float64{1, 0.5, 1.0 / 3.0}
	expected := (10*weights[0] + 20*weights[1] + 30*weights[2]) / (weights[0] + weights[1] + weights[2])
	if math.Abs(forecast.PredictRain-expected) > 1e-9 {
		t.Fatalf("unexpected forecast rain: got %v, want %v", forecast.PredictRain, expected)
	}
	for _, device := range forecast.Devices {
		if device.Sncode == "D" {
			t.Fatal("far device should not be used")
		}
	}
}

func testGridDevicePredict(sncode string, lng, lat, rain float64) gridDevicePredict {
	return gridDevicePredict{
		device: domains.Device{
			Sncode: sncode,
			Lng:    &lng,
			Lat:    &lat,
		},
		predicted: map[int]domains.Predict{
			1: {
				Sncode:      sncode,
				Type:        1,
				Time:        time.Now().UnixMilli(),
				PredictRain: rain,
			},
		},
	}
}
