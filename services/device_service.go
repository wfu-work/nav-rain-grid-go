package services

import (
	"errors"
	"fmt"
	"nav-rain-grid-go/domains"
	"strconv"
	"strings"
	"time"

	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/services"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
)

type DeviceService struct {
	services.CrudService[domains.Device]
}

var DeviceServiceApp = new(DeviceService)

func (s DeviceService) SaveOrUpdate(entity domains.Device) error {
	entity.Sncode = strings.TrimSpace(entity.Sncode)
	if entity.Sncode == "" {
		return errors.New("设备号不能为空")
	}
	if global.NAV_DB == nil {
		return errors.New("database is not initialized")
	}

	now := time.Now().UnixMilli()
	updateValues := map[string]interface{}{
		"sncode":      entity.Sncode,
		"alias":       entity.Alias,
		"type":        entity.Type,
		"lat":         entity.Lat,
		"lng":         entity.Lng,
		"alt":         entity.Alt,
		"gsw":         entity.Gsw,
		"rain":        entity.Rain,
		"status":      entity.Status,
		"last_time":   entity.LastTime,
		"update_time": now,
	}

	if strings.TrimSpace(entity.Guid) != "" {
		return global.NAV_DB.Model(&domains.Device{}).
			Where("guid = ?", entity.Guid).
			Updates(updateValues).Error
	}

	var existing domains.Device
	err := global.NAV_DB.Where("sncode = ?", entity.Sncode).First(&existing).Error
	if err == nil {
		return global.NAV_DB.Model(&existing).Updates(updateValues).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.Create(entity)
}

func (s DeviceService) GetBySncode(sncode string) (*domains.Device, error) {
	sncode = strings.TrimSpace(sncode)
	if sncode == "" {
		return nil, errors.New("设备号不能为空")
	}
	return s.GetByField("sncode", sncode)
}

func (s DeviceService) List(params map[string]string) (list interface{}, total int64, err error) {
	if global.NAV_DB == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	db := s.buildDeviceQuery(params)

	page := utils.Str2Int(params["page"])
	size := utils.Str2Int(params["size"])
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var results []domains.Device
	err = db.Order(resolveDeviceOrder(params)).
		Limit(size).
		Offset(size * (page - 1)).
		Find(&results).Error
	return results, total, err
}

func (s DeviceService) Query(params map[string]string) ([]domains.Device, error) {
	if global.NAV_DB == nil {
		return nil, errors.New("database is not initialized")
	}
	var results []domains.Device
	err := s.buildDeviceQuery(params).Order(resolveDeviceOrder(params)).Find(&results).Error
	return results, err
}

func (s DeviceService) buildDeviceQuery(params map[string]string) *gorm.DB {
	db := global.NAV_DB.Model(new(domains.Device))

	if sncode := strings.TrimSpace(params["sncode"]); sncode != "" {
		db = db.Where("sncode = ?", sncode)
	}
	if alias := strings.TrimSpace(params["alias"]); alias != "" {
		db = db.Where("alias like ?", "%"+alias+"%")
	}
	if deviceType := strings.TrimSpace(params["type"]); deviceType != "" {
		db = db.Where("type = ?", deviceType)
	}
	if status := strings.TrimSpace(params["status"]); status != "" {
		db = db.Where("status = ?", status)
	}
	if gsw, ok := parseOptionalBool(params["gsw"]); ok {
		db = db.Where("gsw = ?", gsw)
	}
	if rain, ok := parseOptionalBool(params["rain"]); ok {
		db = db.Where("rain = ?", rain)
	}
	if content := strings.TrimSpace(params["content"]); content != "" {
		like := "%" + content + "%"
		db = db.Where("sncode like ? or alias like ? or type like ?", like, like, like)
	}
	return db
}

func parseOptionalBool(value string) (bool, bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "undefined" {
		return false, false
	}
	result, err := strconv.ParseBool(value)
	return result, err == nil
}

func resolveDeviceOrder(params map[string]string) string {
	if asc := strings.TrimSpace(params["asc"]); asc != "" {
		if column, ok := deviceOrderColumn(asc); ok {
			return fmt.Sprintf("%s asc", column)
		}
	}
	if desc := strings.TrimSpace(params["desc"]); desc != "" {
		if column, ok := deviceOrderColumn(desc); ok {
			return fmt.Sprintf("%s desc", column)
		}
	}
	return "id desc"
}

func deviceOrderColumn(field string) (string, bool) {
	switch utils.CamelToSnake(strings.TrimSpace(field)) {
	case "id":
		return "id", true
	case "sncode":
		return "sncode", true
	case "alias":
		return "alias", true
	case "type":
		return "type", true
	case "status":
		return "status", true
	case "last_time":
		return "last_time", true
	case "create_time":
		return "create_time", true
	case "update_time":
		return "update_time", true
	default:
		return "", false
	}
}
