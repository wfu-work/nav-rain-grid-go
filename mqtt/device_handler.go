package mqtt

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"nav-rain-grid-go/domains"

	"github.com/wfu-work/nav-common-go-lib/global"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

var registerDeviceHeartbeatOnce sync.Once

type DeviceHeartbeatPayload struct {
	Sncode string
	Alias  string
	Lat    *float64
	Lng    *float64
	Alt    *float64
}

func RegisterDeviceHeartbeatHandler() {
	registerDeviceHeartbeatOnce.Do(func() {
		BrokerServiceApp.AddMessageHandler(HandleDeviceHeartbeat)
	})
}

func HandleDeviceHeartbeat(clientID string, topic string, payload []byte) {
	heartbeat := ParseDeviceHeartbeat(topic, payload)
	if heartbeat.Sncode == "" {
		return
	}

	if err := SaveDeviceHeartbeat(heartbeat); err != nil {
		zap.L().Error("保存设备心跳失败",
			zap.String("clientId", clientID),
			zap.String("topic", topic),
			zap.String("sncode", heartbeat.Sncode),
			zap.Error(err),
		)
		return
	}

	zap.L().Info("设备心跳已保存",
		zap.String("clientId", clientID),
		zap.String("topic", topic),
		zap.String("sncode", heartbeat.Sncode),
	)
}

func SaveDeviceHeartbeat(heartbeat DeviceHeartbeatPayload) error {
	heartbeat.Sncode = strings.TrimSpace(heartbeat.Sncode)
	heartbeat.Alias = strings.TrimSpace(heartbeat.Alias)
	if heartbeat.Sncode == "" {
		return nil
	}
	if global.NAV_DB == nil {
		return fmt.Errorf("database is not initialized")
	}

	now := time.Now().UnixMilli()
	device := domains.Device{
		Sncode:   heartbeat.Sncode,
		Alias:    heartbeat.Alias,
		Lat:      heartbeat.Lat,
		Lng:      heartbeat.Lng,
		Status:   domains.DeviceStatusOnline,
		LastTime: now,
	}
	updateValues := map[string]interface{}{
		"status":      domains.DeviceStatusOnline,
		"last_time":   now,
		"update_time": now,
	}
	if heartbeat.Alias != "" {
		updateValues["alias"] = heartbeat.Alias
	}
	if heartbeat.Lat != nil {
		updateValues["lat"] = heartbeat.Lat
	}
	if heartbeat.Lng != nil {
		updateValues["lng"] = heartbeat.Lng
	}
	if heartbeat.Alt != nil {
		updateValues["alt"] = heartbeat.Alt
	}
	return global.NAV_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sncode"}},
		DoUpdates: clause.Assignments(updateValues),
	}).Create(&device).Error
}

func ParseDeviceHeartbeat(topic string, payload []byte) DeviceHeartbeatPayload {
	if heartbeat, ok := parseDeviceHeartbeatFromJSON(payload); ok {
		if heartbeat.Sncode != "" {
			return heartbeat
		}
	}
	return DeviceHeartbeatPayload{
		Sncode: ExtractSNCode(topic, payload),
	}
}

func parseDeviceHeartbeatFromJSON(payload []byte) (DeviceHeartbeatPayload, bool) {
	body := strings.TrimSpace(string(payload))
	if body == "" {
		return DeviceHeartbeatPayload{}, false
	}

	var value interface{}
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return DeviceHeartbeatPayload{}, false
	}

	data, ok := value.(map[string]interface{})
	if !ok {
		return DeviceHeartbeatPayload{}, false
	}
	return findDeviceHeartbeat(data), true
}

func findDeviceHeartbeat(data map[string]interface{}) DeviceHeartbeatPayload {
	heartbeat := DeviceHeartbeatPayload{
		Sncode: findSNCode(data),
		Alias:  findStringValue(data, []string{"alias", "name", "deviceName", "device_name"}),
		Lat:    findFloatValue(data, []string{"lat", "latitude"}),
		Lng:    findFloatValue(data, []string{"lng", "lon", "longitude"}),
	}
	if heartbeat.Sncode != "" && heartbeat.hasDeviceInfo() {
		return heartbeat
	}

	for _, value := range data {
		if child, ok := value.(map[string]interface{}); ok {
			childHeartbeat := findDeviceHeartbeat(child)
			heartbeat = mergeDeviceHeartbeat(heartbeat, childHeartbeat)
			if heartbeat.Sncode != "" && heartbeat.hasDeviceInfo() {
				return heartbeat
			}
		}
	}
	return heartbeat
}

func (h DeviceHeartbeatPayload) hasDeviceInfo() bool {
	return h.Alias != "" || h.Lat != nil || h.Lng != nil
}

func mergeDeviceHeartbeat(base DeviceHeartbeatPayload, child DeviceHeartbeatPayload) DeviceHeartbeatPayload {
	if base.Sncode == "" {
		base.Sncode = child.Sncode
	}
	if base.Alias == "" {
		base.Alias = child.Alias
	}
	if base.Lat == nil {
		base.Lat = child.Lat
	}
	if base.Lng == nil {
		base.Lng = child.Lng
	}
	return base
}

func ExtractSNCode(topic string, payload []byte) string {
	if sncode := extractSNCodeFromJSON(payload); sncode != "" {
		return sncode
	}
	if sncode := extractSNCodeFromText(payload); sncode != "" {
		return sncode
	}
	return extractSNCodeFromTopic(topic)
}

func extractSNCodeFromJSON(payload []byte) string {
	body := strings.TrimSpace(string(payload))
	if body == "" {
		return ""
	}

	var value interface{}
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return normalizeSNCode(v)
	case map[string]interface{}:
		return findSNCode(v)
	default:
		return ""
	}
}

func findSNCode(data map[string]interface{}) string {
	keys := []string{"sncode", "snCode", "sn_code", "SNCode", "SN_CODE", "sn", "SN"}
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if sncode := normalizeSNCode(value); sncode != "" {
				return sncode
			}
		}
	}
	for _, value := range data {
		if child, ok := value.(map[string]interface{}); ok {
			if sncode := findSNCode(child); sncode != "" {
				return sncode
			}
		}
	}
	return ""
}

func normalizeSNCode(value interface{}) string {
	switch v := value.(type) {
	case string:
		sncode := strings.TrimSpace(v)
		if isValidSNCode(sncode) {
			return sncode
		}
		return ""
	case float64:
		return normalizeSNCode(fmt.Sprintf("%.0f", v))
	case json.Number:
		return normalizeSNCode(v.String())
	default:
		return ""
	}
}

func extractSNCodeFromText(payload []byte) string {
	body := strings.TrimSpace(string(payload))
	if body == "" || strings.HasPrefix(body, "{") || strings.HasPrefix(body, "[") {
		return ""
	}

	for _, pair := range strings.FieldsFunc(body, func(r rune) bool {
		return r == ',' || r == ';' || r == '&' || r == '\n' || r == '\r'
	}) {
		parts := strings.FieldsFunc(pair, func(r rune) bool {
			return r == '=' || r == ':'
		})
		if len(parts) == 2 && isSNCodeKey(parts[0]) {
			return normalizeSNCode(parts[1])
		}
	}
	return normalizeSNCode(body)
}

func extractSNCodeFromTopic(topic string) string {
	segments := strings.Split(topic, "/")
	for _, segment := range segments {
		parts := strings.FieldsFunc(segment, func(r rune) bool {
			return r == '=' || r == ':'
		})
		if len(parts) == 2 && isSNCodeKey(parts[0]) {
			return normalizeSNCode(parts[1])
		}
	}
	return ""
}

func findStringValue(data map[string]interface{}, keys []string) string {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if text := normalizeString(value); text != "" {
				return text
			}
		}
	}
	return ""
}

func findFloatValue(data map[string]interface{}, keys []string) *float64 {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if number, ok := normalizeFloat(value); ok {
				return &number
			}
		}
	}
	return nil
}

func normalizeString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return strings.TrimSpace(v.String())
	case float64:
		return strings.TrimSpace(strconv.FormatFloat(v, 'f', -1, 64))
	default:
		return ""
	}
}

func normalizeFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func isSNCodeKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "sncode", "sn_code", "sn":
		return true
	default:
		return false
	}
}

func isValidSNCode(sncode string) bool {
	if sncode == "" || len(sncode) > 50 {
		return false
	}
	return !strings.ContainsAny(sncode, " \t\r\n{}[]")
}
