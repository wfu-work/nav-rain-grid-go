package services

import (
	"errors"
	"fmt"
	"nav-rain-grid-go/domains"
	"strconv"
	"strings"

	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/services"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
)

type PushRecordService struct {
	services.CrudService[domains.PushRecord]
}

var PushRecordServiceApp = new(PushRecordService)

func (s PushRecordService) Record(entity domains.PushRecord) error {
	if global.NAV_DB == nil {
		return errors.New("database is not initialized")
	}
	return s.Create(entity)
}

func (s PushRecordService) List(params map[string]string) (list interface{}, total int64, err error) {
	if global.NAV_DB == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	db := s.buildPushRecordQuery(params)

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

	var results []domains.PushRecord
	err = db.Order(resolvePushRecordOrder(params)).
		Limit(size).
		Offset(size * (page - 1)).
		Find(&results).Error
	return results, total, err
}

func (s PushRecordService) Query(params map[string]string) ([]domains.PushRecord, error) {
	if global.NAV_DB == nil {
		return nil, errors.New("database is not initialized")
	}
	var results []domains.PushRecord
	err := s.buildPushRecordQuery(params).Order(resolvePushRecordOrder(params)).Find(&results).Error
	return results, err
}

func (s PushRecordService) buildPushRecordQuery(params map[string]string) *gorm.DB {
	db := global.NAV_DB.Model(new(domains.PushRecord))

	if key := strings.TrimSpace(params["key"]); key != "" {
		db = db.Where("key = ?", key)
	}
	if gridGuid := strings.TrimSpace(params["gridGuid"]); gridGuid != "" {
		db = db.Where("grid_guid = ?", gridGuid)
	}
	if gridName := strings.TrimSpace(params["gridName"]); gridName != "" {
		db = db.Where("grid_name like ?", "%"+gridName+"%")
	}
	if gridIdentifier := strings.TrimSpace(params["gridIdentifier"]); gridIdentifier != "" {
		db = db.Where("grid_identifier = ?", gridIdentifier)
	}
	if taskGuid := strings.TrimSpace(params["taskGuid"]); taskGuid != "" {
		db = db.Where("task_guid = ?", taskGuid)
	}
	if status := strings.TrimSpace(params["status"]); status != "" {
		if value, err := strconv.Atoi(status); err == nil {
			db = db.Where("status = ?", value)
		}
	}
	if clientIP := strings.TrimSpace(params["clientIp"]); clientIP != "" {
		db = db.Where("client_ip like ?", "%"+clientIP+"%")
	}
	if content := strings.TrimSpace(params["content"]); content != "" {
		like := "%" + content + "%"
		db = db.Where(
			"key like ? or grid_guid like ? or grid_name like ? or grid_identifier like ? or task_guid like ? or client_ip like ? or nc_file_name like ? or error_msg like ?",
			like, like, like, like, like, like, like, like,
		)
	}
	return db
}

func resolvePushRecordOrder(params map[string]string) string {
	if asc := strings.TrimSpace(params["asc"]); asc != "" {
		if column, ok := pushRecordOrderColumn(asc); ok {
			return fmt.Sprintf("%s asc", column)
		}
	}
	if desc := strings.TrimSpace(params["desc"]); desc != "" {
		if column, ok := pushRecordOrderColumn(desc); ok {
			return fmt.Sprintf("%s desc", column)
		}
	}
	return "request_time desc, id desc"
}

func pushRecordOrderColumn(field string) (string, bool) {
	switch utils.CamelToSnake(strings.TrimSpace(field)) {
	case "id":
		return "id", true
	case "key":
		return "key", true
	case "grid_guid":
		return "grid_guid", true
	case "grid_name":
		return "grid_name", true
	case "grid_identifier":
		return "grid_identifier", true
	case "task_guid":
		return "task_guid", true
	case "base_time":
		return "base_time", true
	case "status":
		return "status", true
	case "http_status":
		return "http_status", true
	case "response_code":
		return "response_code", true
	case "client_ip":
		return "client_ip", true
	case "request_time":
		return "request_time", true
	case "cost_millis":
		return "cost_millis", true
	case "create_time":
		return "create_time", true
	case "update_time":
		return "update_time", true
	default:
		return "", false
	}
}
