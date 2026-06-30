package mqtt

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"nav-rain-grid-go/domains"
	"nav-rain-grid-go/services"

	"github.com/wfu-work/nav-common-go-lib/global"
	"go.uber.org/zap"
)

var registerPredictOnce sync.Once

type PredictPayload struct {
	Sncode      string
	BaseTime    int64
	Predictions []predictItem
}

type predictItem struct {
	Hour      int
	Rain      float64
	RainLevel int
	Time      int64
}

func RegisterPredictHandler() {
	registerPredictOnce.Do(func() {
		BrokerServiceApp.AddMessageHandler(HandlePredictMessage)
	})
}

func HandlePredictMessage(clientID string, topic string, payload []byte) {
	data, ok := ParsePredictPayload(topic, payload)
	if !ok {
		return
	}

	if err := SavePredictPayload(data); err != nil {
		zap.L().Error("保存MQTT降雨预测数据失败",
			zap.String("clientID", clientID),
			zap.String("topic", topic),
			zap.String("sncode", data.Sncode),
			zap.Error(err),
		)
		return
	}

	zap.L().Info("MQTT降雨预测数据已保存",
		zap.String("clientID", clientID),
		zap.String("topic", topic),
		zap.String("sncode", data.Sncode),
		zap.Int("count", len(data.Predictions)),
	)
}

func SavePredictPayload(data PredictPayload) error {
	if global.NAV_DB == nil {
		return fmt.Errorf("database is not initialized")
	}
	if data.Sncode == "" || data.BaseTime == 0 || len(data.Predictions) == 0 {
		return nil
	}

	for _, item := range data.Predictions {
		if item.Hour == 0 {
			continue
		}
		predictTime := item.Time
		if predictTime == 0 {
			predictTime = time.UnixMilli(data.BaseTime).Add(time.Duration(item.Hour) * time.Hour).UnixMilli()
		}
		result := services.PredictServiceApp.CreateOne(data.BaseTime, domains.Predict{
			BaseTime:         data.BaseTime,
			Time:             predictTime,
			Sncode:           data.Sncode,
			PredictRain:      item.Rain,
			PredictRainLevel: item.RainLevel,
			Type:             item.Hour,
		})
		if result == nil {
			return fmt.Errorf("保存%d小时预测降雨失败", item.Hour)
		}
	}
	return nil
}

func ParsePredictPayload(topic string, payload []byte) (PredictPayload, bool) {
	body := strings.TrimSpace(string(payload))
	if body == "" {
		return PredictPayload{}, false
	}

	var value interface{}
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return PredictPayload{}, false
	}

	root, ok := value.(map[string]interface{})
	if !ok {
		return PredictPayload{}, false
	}

	data := PredictPayload{
		Sncode:   ExtractSNCode(topic, payload),
		BaseTime: extractBaseTime(root),
	}
	if data.BaseTime == 0 {
		data.BaseTime = time.Now().UnixMilli()
	}

	data.Predictions = append(data.Predictions, extractPredictItems(root)...)
	if len(data.Predictions) == 0 {
		return PredictPayload{}, false
	}
	if data.Sncode == "" {
		return PredictPayload{}, false
	}
	return data, true
}

func extractBaseTime(data map[string]interface{}) int64 {
	keys := []string{"baseTime", "base_time", "time", "timestamp", "ts"}
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if timestamp := normalizeTimestamp(value); timestamp != 0 {
				return timestamp
			}
		}
	}
	return 0
}

func extractPredictItems(data map[string]interface{}) []predictItem {
	items := make([]predictItem, 0, 3)
	items = append(items, extractPredictItemsFromArray(data)...)
	items = append(items, extractPredictItemsFromMap(data)...)
	for _, hour := range []int{1, 12, 24} {
		if rain, ok := extractRainByHour(data, hour); ok {
			items = upsertPredictItem(items, predictItem{
				Hour:      hour,
				Rain:      rain,
				RainLevel: extractRainLevelByHour(data, hour),
			})
		}
	}
	return items
}

func extractPredictItemsFromMap(data map[string]interface{}) []predictItem {
	keys := []string{"rain", "rains", "predict", "predictRain", "predict_rain", "prediction", "predictions"}
	items := make([]predictItem, 0, 3)
	for _, key := range keys {
		raw, ok := data[key]
		if !ok {
			continue
		}
		values, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		for _, hour := range []int{1, 12, 24} {
			if rain, ok := extractRainByHour(values, hour); ok {
				items = upsertPredictItem(items, predictItem{
					Hour:      hour,
					Rain:      rain,
					RainLevel: extractRainLevelByHour(values, hour),
				})
			}
		}
	}
	return items
}

func extractPredictItemsFromArray(data map[string]interface{}) []predictItem {
	keys := []string{"predictions", "predicts", "predictList", "predict_list", "rainPredictions", "rain_predictions", "data"}
	items := make([]predictItem, 0, 3)
	for _, key := range keys {
		raw, ok := data[key]
		if !ok {
			continue
		}
		values, ok := raw.([]interface{})
		if !ok {
			continue
		}
		for _, value := range values {
			itemMap, ok := value.(map[string]interface{})
			if !ok {
				continue
			}
			hour := extractHour(itemMap)
			rain, ok := extractRain(itemMap)
			if hour == 0 || !ok {
				continue
			}
			items = upsertPredictItem(items, predictItem{
				Hour:      hour,
				Rain:      rain,
				RainLevel: extractRainLevel(itemMap),
				Time:      extractPredictTime(itemMap),
			})
		}
	}
	return items
}

func extractHour(data map[string]interface{}) int {
	keys := []string{"hour", "hours", "type", "forecastHour", "forecast_hour", "duration"}
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if hour, ok := numberToInt(value); ok && isPredictHour(hour) {
				return hour
			}
			if hour, ok := parseHourString(value); ok {
				return hour
			}
		}
	}
	return 0
}

func extractRain(data map[string]interface{}) (float64, bool) {
	keys := []string{"predictRain", "predict_rain", "rain", "rainfall", "value", "val"}
	for _, key := range keys {
		if value, ok := data[key]; ok {
			return numberToFloat(value)
		}
	}
	return 0, false
}

func extractRainLevel(data map[string]interface{}) int {
	keys := []string{"PredictLevel", "predictLevel", "predict_level", "predictRainLevel", "predict_rain_level", "rainLevel", "rain_level", "level"}
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if level, ok := numberToInt(value); ok {
				return level
			}
		}
	}
	return 0
}

func extractRainByHour(data map[string]interface{}, hour int) (float64, bool) {
	keys := rainKeys(hour)
	for _, key := range keys {
		if value, ok := data[key]; ok {
			return numberToFloat(value)
		}
	}
	return 0, false
}

func extractRainLevelByHour(data map[string]interface{}, hour int) int {
	keys := rainLevelKeys(hour)
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if level, ok := numberToInt(value); ok {
				return level
			}
		}
	}
	return extractRainLevel(data)
}

func rainKeys(hour int) []string {
	return []string{
		fmt.Sprintf("rain%d", hour),
		fmt.Sprintf("rain_%d", hour),
		fmt.Sprintf("rain%dh", hour),
		fmt.Sprintf("rain_%dh", hour),
		fmt.Sprintf("h%d", hour),
		strconv.Itoa(hour),
		fmt.Sprintf("%dh", hour),
		fmt.Sprintf("predictRain%d", hour),
		fmt.Sprintf("predictRain%dH", hour),
		fmt.Sprintf("predict_rain_%d", hour),
		fmt.Sprintf("predict_rain_%dh", hour),
		fmt.Sprintf("predict%d", hour),
		fmt.Sprintf("predict_%d", hour),
	}
}

func rainLevelKeys(hour int) []string {
	return []string{
		fmt.Sprintf("PredictLevel%d", hour),
		fmt.Sprintf("PredictLevel%dH", hour),
		fmt.Sprintf("predictLevel%d", hour),
		fmt.Sprintf("predictLevel%dH", hour),
		fmt.Sprintf("predict_level_%d", hour),
		fmt.Sprintf("predict_level_%dh", hour),
		fmt.Sprintf("predictRainLevel%d", hour),
		fmt.Sprintf("predictRainLevel%dH", hour),
		fmt.Sprintf("predict_rain_level_%d", hour),
		fmt.Sprintf("predict_rain_level_%dh", hour),
		fmt.Sprintf("rainLevel%d", hour),
		fmt.Sprintf("rainLevel%dH", hour),
		fmt.Sprintf("rain_level_%d", hour),
		fmt.Sprintf("rain_level_%dh", hour),
		fmt.Sprintf("level%d", hour),
		fmt.Sprintf("level_%d", hour),
		fmt.Sprintf("level%dh", hour),
		fmt.Sprintf("level_%dh", hour),
	}
}

func extractPredictTime(data map[string]interface{}) int64 {
	keys := []string{"time", "timestamp", "predictTime", "predict_time"}
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if timestamp := normalizeTimestamp(value); timestamp != 0 {
				return timestamp
			}
		}
	}
	return 0
}

func upsertPredictItem(items []predictItem, item predictItem) []predictItem {
	for i := range items {
		if items[i].Hour == item.Hour {
			items[i] = item
			return items
		}
	}
	return append(items, item)
}

func isPredictHour(hour int) bool {
	return hour == 1 || hour == 12 || hour == 24
}

func normalizeTimestamp(value interface{}) int64 {
	switch v := value.(type) {
	case json.Number:
		return normalizeTimestamp(v.String())
	case float64:
		return normalizeTimestamp(strconv.FormatFloat(v, 'f', -1, 64))
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return 0
		}
		if parsed, err := strconv.ParseInt(text, 10, 64); err == nil {
			return fixTimestampUnit(parsed)
		}
		if parsed, err := time.Parse(time.RFC3339, text); err == nil {
			return parsed.UnixMilli()
		}
		for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02"} {
			if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
				return parsed.UnixMilli()
			}
		}
	}
	return 0
}

func fixTimestampUnit(value int64) int64 {
	if value <= 0 {
		return 0
	}
	if value < 1_000_000_000_000 {
		return value * 1000
	}
	return value
}

func numberToFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func numberToInt(value interface{}) (int, bool) {
	f, ok := numberToFloat(value)
	if !ok {
		return 0, false
	}
	return int(f), true
}

func parseHourString(value interface{}) (int, bool) {
	text, ok := value.(string)
	if !ok {
		return 0, false
	}
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.TrimSuffix(text, "hour")
	text = strings.TrimSuffix(text, "hours")
	text = strings.TrimSuffix(text, "h")
	hour, err := strconv.Atoi(text)
	if err != nil || !isPredictHour(hour) {
		return 0, false
	}
	return hour, true
}
