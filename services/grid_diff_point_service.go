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

type GridDiffPointService struct {
	services.CrudService[domains.GridDiffPoint]
}

var GridDiffPointServiceApp = new(GridDiffPointService)

func (s GridDiffPointService) List(params map[string]string) (list interface{}, total int64, err error) {
	if global.NAV_DB == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	db := s.buildGridDiffPointQuery(params)

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

	var results []domains.GridDiffPoint
	err = db.Order(resolveGridDiffPointOrder(params)).
		Limit(size).
		Offset(size * (page - 1)).
		Find(&results).Error
	return results, total, err
}

func (s GridDiffPointService) Query(params map[string]string) ([]domains.GridDiffPoint, error) {
	if global.NAV_DB == nil {
		return nil, errors.New("database is not initialized")
	}
	var results []domains.GridDiffPoint
	err := s.buildGridDiffPointQuery(params).Order(resolveGridDiffPointOrder(params)).Find(&results).Error
	return results, err
}

func (s GridDiffPointService) buildGridDiffPointQuery(params map[string]string) *gorm.DB {
	db := global.NAV_DB.Model(new(domains.GridDiffPoint))

	if taskGuid := strings.TrimSpace(params["taskGuid"]); taskGuid != "" {
		db = db.Where("task_guid = ?", taskGuid)
	}
	if gridGuid := strings.TrimSpace(params["gridGuid"]); gridGuid != "" {
		db = db.Where("grid_guid = ?", gridGuid)
	}
	if gridName := strings.TrimSpace(params["gridName"]); gridName != "" {
		db = db.Where("grid_name like ?", "%"+gridName+"%")
	}
	if baseTime := strings.TrimSpace(params["baseTime"]); baseTime != "" {
		if value, err := strconv.ParseInt(baseTime, 10, 64); err == nil {
			db = db.Where("base_time = ?", value)
		}
	}
	if content := strings.TrimSpace(params["content"]); content != "" {
		like := "%" + content + "%"
		db = db.Where("grid_name like ? or task_guid like ? or grid_guid like ?", like, like, like)
	}
	return db
}

func resolveGridDiffPointOrder(params map[string]string) string {
	if asc := strings.TrimSpace(params["asc"]); asc != "" {
		if column, ok := gridDiffPointOrderColumn(asc); ok {
			return fmt.Sprintf("%s asc", column)
		}
	}
	if desc := strings.TrimSpace(params["desc"]); desc != "" {
		if column, ok := gridDiffPointOrderColumn(desc); ok {
			return fmt.Sprintf("%s desc", column)
		}
	}
	return "base_time desc, id asc"
}

func gridDiffPointOrderColumn(field string) (string, bool) {
	switch utils.CamelToSnake(strings.TrimSpace(field)) {
	case "id":
		return "id", true
	case "task_guid":
		return "task_guid", true
	case "grid_guid":
		return "grid_guid", true
	case "grid_name":
		return "grid_name", true
	case "base_time":
		return "base_time", true
	case "center_lng":
		return "center_lng", true
	case "center_lat":
		return "center_lat", true
	case "predict_rain_1_h":
		return "predict_rain1_h", true
	case "predict_rain_12_h":
		return "predict_rain12_h", true
	case "predict_rain_24_h":
		return "predict_rain24_h", true
	case "create_time":
		return "create_time", true
	case "update_time":
		return "update_time", true
	default:
		return "", false
	}
}
