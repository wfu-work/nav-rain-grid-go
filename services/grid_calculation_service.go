package services

import (
	"errors"
	"math"
	"nav-rain-grid-go/domains"
	"sort"
	"strings"
	"time"

	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
)

const (
	gridNearestDeviceCount = 3
)

type GridCalculationService struct{}

var GridCalculationServiceApp = new(GridCalculationService)

type gridDevicePredict struct {
	device    domains.Device
	predicted map[int]domains.Predict
}

type gridWeightedDevice struct {
	device   domains.Device
	predict  domains.Predict
	distance float64
	weight   float64
}

func (s GridCalculationService) CalculateEnabledGrids(now time.Time) ([]domains.GridDiffResult, error) {
	if global.NAV_DB == nil {
		return nil, errors.New("database is not initialized")
	}

	var grids []domains.Grid
	if err := global.NAV_DB.
		Where("status = ?", domains.GridStatusEnabled).
		Find(&grids).Error; err != nil {
		return nil, err
	}

	results := make([]domains.GridDiffResult, 0, len(grids))
	for _, grid := range grids {
		result, ok, err := s.calculateGrid(grid, now)
		if err != nil {
			return nil, err
		}
		if ok {
			if err := s.saveGridDiffResult(result); err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (s GridCalculationService) calculateGrid(grid domains.Grid, now time.Time) (domains.GridDiffResult, bool, error) {
	resolution := normalizeGridResolution(grid.Resolution)
	sncodes := parseGridSncodes(grid.Sncodes)
	minDevice := normalizeGridMinDevice(grid.MinDevice)
	if len(sncodes) < minDevice {
		return domains.GridDiffResult{}, false, nil
	}

	devices, err := s.queryGridDevices(sncodes)
	if err != nil {
		return domains.GridDiffResult{}, false, err
	}
	if len(devices) < minDevice {
		return domains.GridDiffResult{}, false, nil
	}

	baseTime := alignGridBaseTime(now)
	predictions, err := s.queryPredictionsByBaseTime(sncodes, baseTime)
	if err != nil {
		return domains.GridDiffResult{}, false, err
	}

	devicePredicts := make([]gridDevicePredict, 0, len(devices))
	for _, device := range devices {
		devicePredicts = append(devicePredicts, gridDevicePredict{
			device:    device,
			predicted: predictions[device.Sncode],
		})
	}

	centers := buildGridCenters(devices, resolution)
	if len(centers) == 0 {
		return domains.GridDiffResult{}, false, nil
	}

	result := domains.GridDiffResult{
		GridGuid:   grid.Guid,
		GridName:   grid.Name,
		BaseTime:   baseTime,
		Resolution: resolution,
		Points:     make([]domains.GridDiffPointResult, 0, len(centers)),
	}
	for _, center := range centers {
		point := domains.GridDiffPointResult{
			CenterLng: center.lng,
			CenterLat: center.lat,
		}
		point.Forecast1H = interpolateGridForecast(center, devicePredicts, 1)
		point.Forecast12H = interpolateGridForecast(center, devicePredicts, 12)
		point.Forecast24H = interpolateGridForecast(center, devicePredicts, 24)
		if point.Forecast1H != nil && point.Forecast12H != nil && point.Forecast24H != nil {
			result.Points = append(result.Points, point)
		}
	}
	if len(result.Points) == 0 {
		return domains.GridDiffResult{}, false, nil
	}
	return result, true, nil
}

func (s GridCalculationService) queryGridDevices(sncodes []string) ([]domains.Device, error) {
	var devices []domains.Device
	err := global.NAV_DB.
		Where("sncode IN ?", sncodes).
		Where("lat IS NOT NULL AND lng IS NOT NULL").
		Find(&devices).Error
	return devices, err
}

func (s GridCalculationService) queryPredictionsByBaseTime(sncodes []string, baseTime int64) (map[string]map[int]domains.Predict, error) {
	var predictions []domains.Predict
	err := global.NAV_DB.
		Where("sncode IN ?", sncodes).
		Where("type IN ?", []int{1, 12, 24}).
		Where("base_time = ?", baseTime).
		Order("update_time desc, id desc").
		Find(&predictions).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[int]domains.Predict)
	for _, prediction := range predictions {
		if result[prediction.Sncode] == nil {
			result[prediction.Sncode] = make(map[int]domains.Predict)
		}
		if _, exists := result[prediction.Sncode][prediction.Type]; !exists {
			result[prediction.Sncode][prediction.Type] = prediction
		}
	}
	return result, nil
}

func alignGridBaseTime(now time.Time) int64 {
	aligned := time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), 0, 0, 0,
		now.Location(),
	)
	return aligned.UnixMilli()
}

func (s GridCalculationService) saveGridDiffResult(result domains.GridDiffResult) error {
	return global.NAV_DB.Transaction(func(tx *gorm.DB) error {
		taskGuid, err := upsertGridDiffTask(tx, result)
		if err != nil {
			return err
		}
		if err := tx.Unscoped().Where("task_guid = ?", taskGuid).Delete(&domains.GridDiffPoint{}).Error; err != nil {
			return err
		}
		points := buildGridDiffPointRows(taskGuid, result)
		if len(points) == 0 {
			return nil
		}
		return tx.Create(&points).Error
	})
}

func upsertGridDiffTask(tx *gorm.DB, result domains.GridDiffResult) (string, error) {
	var existing domains.GridDiffTask
	err := tx.Where("grid_guid = ? AND base_time = ?", result.GridGuid, result.BaseTime).First(&existing).Error
	updateValues := map[string]interface{}{
		"grid_guid":   result.GridGuid,
		"grid_name":   result.GridName,
		"base_time":   result.BaseTime,
		"resolution":  result.Resolution,
		"point_count": len(result.Points),
		"status":      domains.GridDiffTaskStatusSuccess,
		"error_msg":   "",
		"update_time": time.Now().UnixMilli(),
	}
	if err == nil {
		return existing.Guid, tx.Model(&existing).Updates(updateValues).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	task := domains.GridDiffTask{
		GridGuid:   result.GridGuid,
		GridName:   result.GridName,
		BaseTime:   result.BaseTime,
		Resolution: result.Resolution,
		PointCount: len(result.Points),
		Status:     domains.GridDiffTaskStatusSuccess,
	}
	if err := tx.Create(&task).Error; err != nil {
		return "", err
	}
	return task.Guid, nil
}

func buildGridDiffPointRows(taskGuid string, result domains.GridDiffResult) []domains.GridDiffPoint {
	points := make([]domains.GridDiffPoint, 0, len(result.Points))
	for _, point := range result.Points {
		if point.Forecast1H == nil || point.Forecast12H == nil || point.Forecast24H == nil {
			continue
		}
		points = append(points, domains.GridDiffPoint{
			TaskGuid:            taskGuid,
			GridGuid:            result.GridGuid,
			GridName:            result.GridName,
			BaseTime:            result.BaseTime,
			CenterLng:           point.CenterLng,
			CenterLat:           point.CenterLat,
			PredictTime1H:       point.Forecast1H.Time,
			PredictRain1H:       point.Forecast1H.PredictRain,
			PredictRainLevel1H:  point.Forecast1H.PredictRainLevel,
			PredictTime12H:      point.Forecast12H.Time,
			PredictRain12H:      point.Forecast12H.PredictRain,
			PredictRainLevel12H: point.Forecast12H.PredictRainLevel,
			PredictTime24H:      point.Forecast24H.Time,
			PredictRain24H:      point.Forecast24H.PredictRain,
			PredictRainLevel24H: point.Forecast24H.PredictRainLevel,
		})
	}
	return points
}

type gridCenter struct {
	lng float64
	lat float64
}

func buildGridCenters(devices []domains.Device, resolution float64) []gridCenter {
	resolution = normalizeGridResolution(resolution)
	minLng, maxLng := math.MaxFloat64, -math.MaxFloat64
	minLat, maxLat := math.MaxFloat64, -math.MaxFloat64
	for _, device := range devices {
		if device.Lng == nil || device.Lat == nil {
			continue
		}
		minLng = math.Min(minLng, *device.Lng)
		maxLng = math.Max(maxLng, *device.Lng)
		minLat = math.Min(minLat, *device.Lat)
		maxLat = math.Max(maxLat, *device.Lat)
	}
	if minLng == math.MaxFloat64 || minLat == math.MaxFloat64 {
		return nil
	}

	startLngIndex := int(math.Floor(minLng / resolution))
	endLngIndex := int(math.Floor(maxLng / resolution))
	startLatIndex := int(math.Floor(minLat / resolution))
	endLatIndex := int(math.Floor(maxLat / resolution))

	centers := make([]gridCenter, 0, (endLngIndex-startLngIndex+1)*(endLatIndex-startLatIndex+1))
	for lngIndex := startLngIndex; lngIndex <= endLngIndex; lngIndex++ {
		for latIndex := startLatIndex; latIndex <= endLatIndex; latIndex++ {
			centers = append(centers, gridCenter{
				lng: roundCoordinate((float64(lngIndex) + 0.5) * resolution),
				lat: roundCoordinate((float64(latIndex) + 0.5) * resolution),
			})
		}
	}
	return centers
}

func interpolateGridForecast(center gridCenter, devicePredicts []gridDevicePredict, forecastHour int) *domains.GridDiffForecast {
	candidates := make([]gridWeightedDevice, 0, len(devicePredicts))
	for _, devicePredict := range devicePredicts {
		prediction, ok := devicePredict.predicted[forecastHour]
		if !ok || devicePredict.device.Lng == nil || devicePredict.device.Lat == nil {
			continue
		}
		distance := coordinateDistance(center.lng, center.lat, *devicePredict.device.Lng, *devicePredict.device.Lat)
		weight := distanceWeight(distance)
		candidates = append(candidates, gridWeightedDevice{
			device:   devicePredict.device,
			predict:  prediction,
			distance: distance,
			weight:   weight,
		})
	}
	if len(candidates) < gridNearestDeviceCount {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})
	nearest := candidates[:gridNearestDeviceCount]

	var totalWeight float64
	for _, candidate := range nearest {
		totalWeight += candidate.weight
	}
	if totalWeight <= 0 {
		return nil
	}

	var predictRain float64
	devices := make([]domains.GridDiffDeviceWeight, 0, gridNearestDeviceCount)
	for _, candidate := range nearest {
		normalizedWeight := candidate.weight / totalWeight
		predictRain += candidate.predict.PredictRain * normalizedWeight
		devices = append(devices, domains.GridDiffDeviceWeight{
			Sncode:      candidate.device.Sncode,
			Distance:    candidate.distance,
			Weight:      normalizedWeight,
			PredictRain: candidate.predict.PredictRain,
		})
	}

	return &domains.GridDiffForecast{
		ForecastHour:     forecastHour,
		Time:             nearest[0].predict.Time,
		PredictRain:      predictRain,
		PredictRainLevel: utils.GetLevel(predictRain),
		Devices:          devices,
	}
}

func parseGridSncodes(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		sncode := strings.TrimSpace(part)
		if sncode == "" {
			continue
		}
		if _, exists := seen[sncode]; exists {
			continue
		}
		seen[sncode] = struct{}{}
		result = append(result, sncode)
	}
	return result
}

func normalizeGridMinDevice(minDevice int) int {
	if minDevice < gridNearestDeviceCount {
		return gridNearestDeviceCount
	}
	return minDevice
}

func coordinateDistance(lng1, lat1, lng2, lat2 float64) float64 {
	lngDelta := lng1 - lng2
	latDelta := lat1 - lat2
	return math.Sqrt(lngDelta*lngDelta + latDelta*latDelta)
}

func distanceWeight(distance float64) float64 {
	if distance <= 0 {
		return 1e12
	}
	return 1 / distance
}

func roundCoordinate(value float64) float64 {
	return math.Round(value*1e8) / 1e8
}
