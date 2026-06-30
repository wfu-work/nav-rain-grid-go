package mqtt

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"nav-rain-grid-go/domains"

	"github.com/wfu-work/nav-common-go-lib/global"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

var registerDeviceHeartbeatOnce sync.Once

func RegisterDeviceHeartbeatHandler() {
	registerDeviceHeartbeatOnce.Do(func() {
		BrokerServiceApp.AddMessageHandler(HandleDeviceHeartbeat)
	})
}

func HandleDeviceHeartbeat(clientID string, topic string, payload []byte) {
	sncode := ExtractSNCode(topic, payload)
	if sncode == "" {
		return
	}

	if err := SaveDeviceHeartbeat(sncode); err != nil {
		zap.L().Error("保存设备心跳失败",
			zap.String("clientId", clientID),
			zap.String("topic", topic),
			zap.String("sncode", sncode),
			zap.Error(err),
		)
		return
	}

	zap.L().Info("设备心跳已保存",
		zap.String("clientId", clientID),
		zap.String("topic", topic),
		zap.String("sncode", sncode),
	)
}

func SaveDeviceHeartbeat(sncode string) error {
	sncode = strings.TrimSpace(sncode)
	if sncode == "" {
		return nil
	}
	if global.NAV_DB == nil {
		return fmt.Errorf("database is not initialized")
	}

	now := time.Now().UnixMilli()
	device := domains.Device{
		Sncode:   sncode,
		Status:   domains.DeviceStatusOnline,
		LastTime: now,
	}
	return global.NAV_DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "sncode"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":      domains.DeviceStatusOnline,
			"last_time":   now,
			"update_time": now,
		}),
	}).Create(&device).Error
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
