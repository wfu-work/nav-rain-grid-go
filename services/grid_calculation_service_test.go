package services

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nav-rain-grid-go/domains"

	"github.com/wfu-work/nav-common-go-lib/global"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGridCalculationServiceCalculateEnabledGrids(t *testing.T) {
	oldDB := global.NAV_DB
	oldNCOutputDir := gridDiffNCOutputDir
	defer func() {
		global.NAV_DB = oldDB
		gridDiffNCOutputDir = oldNCOutputDir
	}()
	gridDiffNCOutputDir = t.TempDir()

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
	if len(point.Forecast1H.Devices) == 0 {
		t.Fatal("expected weighted devices")
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
	if task.GridIdentifier != "ceshi" {
		t.Fatalf("unexpected task grid identifier: %v", task.GridIdentifier)
	}
	if task.CoordinateSystem != domains.DefaultGridCoordinateSystem {
		t.Fatalf("unexpected task coordinate system: %v", task.CoordinateSystem)
	}
	if task.NcStatus != domains.GridDiffTaskNcStatusSuccess {
		t.Fatalf("unexpected nc status: %v", task.NcStatus)
	}
	expectedNCFileName := "shouming_hourly_precipitation_forecast_ceshi_wgs84_" + formatGridNCBaseTime(baseTime) + ".nc"
	if task.NcFileName != expectedNCFileName {
		t.Fatalf("unexpected nc file name: got %s, want %s", task.NcFileName, expectedNCFileName)
	}
	if task.NcFileSize <= 0 {
		t.Fatalf("unexpected nc file size: %d", task.NcFileSize)
	}
	if len(task.NcChecksum) != 64 {
		t.Fatalf("unexpected nc checksum: %s", task.NcChecksum)
	}
	ncData, err := os.ReadFile(task.NcFilePath)
	if err != nil {
		t.Fatalf("read nc file: %v", err)
	}
	if int64(len(ncData)) != task.NcFileSize {
		t.Fatalf("unexpected nc file size on disk: got %d, want %d", len(ncData), task.NcFileSize)
	}
	if !bytes.HasPrefix(ncData, []byte{'C', 'D', 'F', 1}) {
		t.Fatalf("unexpected nc magic: %v", ncData[:4])
	}
	for _, variableName := range []string{"longitude", "latitude", "forecast_1h", "forecast_12h", "forecast_24h"} {
		if !bytes.Contains(ncData, []byte(variableName)) {
			t.Fatalf("nc file should contain variable %s", variableName)
		}
	}
	if filepath.Base(task.NcFilePath) != expectedNCFileName {
		t.Fatalf("unexpected nc file path: %s", task.NcFilePath)
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

func TestGridCalculationServiceSkipsBelowMinimumDevices(t *testing.T) {
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
	for _, device := range []domains.Device{
		{Sncode: "A", Lat: &latA, Lng: &lngA},
		{Sncode: "B", Lat: &latB, Lng: &lngB},
	} {
		if err := db.Create(&device).Error; err != nil {
			t.Fatalf("create device: %v", err)
		}
	}

	if err := db.Create(&domains.Grid{
		Name:       "设备不足格网",
		Sncodes:    "A,B",
		Resolution: 0.01,
		Status:     domains.GridStatusEnabled,
	}).Error; err != nil {
		t.Fatalf("create grid: %v", err)
	}

	now := time.Date(2026, 7, 1, 10, 30, 0, 0, time.Local)
	baseTime := alignGridBaseTime(now)
	hourMillis := int64(time.Hour / time.Millisecond)
	for _, sncode := range []string{"A", "B"} {
		for _, forecastHour := range []int{1, 12, 24} {
			if err := db.Create(&domains.Predict{
				Sncode:      sncode,
				Type:        forecastHour,
				BaseTime:    baseTime,
				Time:        baseTime + int64(forecastHour)*hourMillis,
				PredictRain: 10,
			}).Error; err != nil {
				t.Fatalf("create prediction: %v", err)
			}
		}
	}

	results, err := GridCalculationService{}.CalculateEnabledGrids(now)
	if err != nil {
		t.Fatalf("calculate grids: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("less than three devices should not calculate grid: %#v", results)
	}

	var taskCount int64
	if err := db.Model(&domains.GridDiffTask{}).Count(&taskCount).Error; err != nil {
		t.Fatalf("count grid diff tasks: %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("less than three devices should not save task, got %d", taskCount)
	}
}

func TestInterpolateGridForecastUsesFiveKilometerCircleWeights(t *testing.T) {
	center := gridCenter{lng: 0, lat: 0}
	devices := []gridDevicePredict{
		testGridDevicePredict("A", 0.01, 0, 10),
		testGridDevicePredict("B", 0.02, 0, 20),
		testGridDevicePredict("C", 0.10, 0, 1000),
	}

	forecast := interpolateGridForecast(center, devices, 1)
	if forecast == nil {
		t.Fatal("expected forecast")
	}
	if len(forecast.Devices) != 2 {
		t.Fatalf("expected two devices inside 5km circle, got %d", len(forecast.Devices))
	}

	weightA := distanceWeight(coordinateDistanceKm(center.lng, center.lat, 0.01, 0))
	weightB := distanceWeight(coordinateDistanceKm(center.lng, center.lat, 0.02, 0))
	expected := (10*weightA + 20*weightB) / (weightA + weightB)
	if math.Abs(forecast.PredictRain-expected) > 1e-9 {
		t.Fatalf("unexpected forecast rain: got %v, want %v", forecast.PredictRain, expected)
	}
	for _, device := range forecast.Devices {
		if device.Sncode == "C" {
			t.Fatal("device outside 5km circle should not be used")
		}
	}
}

func TestInterpolateGridForecastUsesSingleDeviceValueInNonOverlapArea(t *testing.T) {
	center := gridCenter{lng: 0, lat: 0}
	devices := []gridDevicePredict{
		testGridDevicePredict("A", 0.03, 0, 10),
		testGridDevicePredict("B", 0.10, 0, 1000),
	}

	forecast := interpolateGridForecast(center, devices, 1)
	if forecast == nil {
		t.Fatal("expected forecast")
	}
	if len(forecast.Devices) != 1 {
		t.Fatalf("expected one device inside non-overlap area, got %d", len(forecast.Devices))
	}
	if forecast.Devices[0].Sncode != "A" {
		t.Fatalf("unexpected device: %s", forecast.Devices[0].Sncode)
	}
	if math.Abs(forecast.PredictRain-10) > 1e-9 {
		t.Fatalf("single device area should use station value, got %v", forecast.PredictRain)
	}
	if math.Abs(forecast.Devices[0].Weight-1) > 1e-9 {
		t.Fatalf("single device weight should be 1, got %v", forecast.Devices[0].Weight)
	}
}

func TestBuildGridCentersUsesFiveKilometerDeviceCircles(t *testing.T) {
	lat, lng := 0.0, 0.0
	centers := buildGridCenters([]domains.Device{
		{Sncode: "A", Lat: &lat, Lng: &lng},
	}, 0.01)

	if len(centers) == 0 {
		t.Fatal("expected centers inside device circle")
	}
	for _, center := range centers {
		distance := coordinateDistanceKm(center.lng, center.lat, lng, lat)
		if distance > gridInfluenceRadiusKm+distanceEpsilonKm {
			t.Fatalf("center outside 5km circle: center=%#v distance=%v", center, distance)
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
