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

type GridDiffTaskService struct {
	services.CrudService[domains.GridDiffTask]
}

var GridDiffTaskServiceApp = new(GridDiffTaskService)

func (s GridDiffTaskService) List(params map[string]string) (list interface{}, total int64, err error) {
	if global.NAV_DB == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	db := s.buildGridDiffTaskQuery(params)

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

	var results []domains.GridDiffTask
	err = db.Order(resolveGridDiffTaskOrder(params)).
		Limit(size).
		Offset(size * (page - 1)).
		Find(&results).Error
	return results, total, err
}

func (s GridDiffTaskService) Query(params map[string]string) ([]domains.GridDiffTask, error) {
	if global.NAV_DB == nil {
		return nil, errors.New("database is not initialized")
	}
	var results []domains.GridDiffTask
	err := s.buildGridDiffTaskQuery(params).Order(resolveGridDiffTaskOrder(params)).Find(&results).Error
	return results, err
}

func (s GridDiffTaskService) LatestSuccessNCByGridGuid(gridGuid string) (domains.GridDiffTask, error) {
	if global.NAV_DB == nil {
		return domains.GridDiffTask{}, errors.New("database is not initialized")
	}
	gridGuid = strings.TrimSpace(gridGuid)
	if gridGuid == "" {
		return domains.GridDiffTask{}, errors.New("grid guid is required")
	}
	var result domains.GridDiffTask
	err := global.NAV_DB.
		Where("grid_guid = ? AND nc_status = ? AND nc_file_path <> ''", gridGuid, domains.GridDiffTaskNcStatusSuccess).
		Order("base_time desc, id desc").
		First(&result).Error
	return result, err
}

func (s GridDiffTaskService) buildGridDiffTaskQuery(params map[string]string) *gorm.DB {
	db := global.NAV_DB.Model(new(domains.GridDiffTask))

	if gridGuid := strings.TrimSpace(params["gridGuid"]); gridGuid != "" {
		db = db.Where("grid_guid = ?", gridGuid)
	}
	if gridName := strings.TrimSpace(params["gridName"]); gridName != "" {
		db = db.Where("grid_name like ?", "%"+gridName+"%")
	}
	if gridIdentifier := strings.TrimSpace(params["gridIdentifier"]); gridIdentifier != "" {
		db = db.Where("grid_identifier = ?", gridIdentifier)
	}
	if coordinateSystem := strings.TrimSpace(params["coordinateSystem"]); coordinateSystem != "" {
		db = db.Where("coordinate_system = ?", coordinateSystem)
	}
	if baseTime := strings.TrimSpace(params["baseTime"]); baseTime != "" {
		if value, err := strconv.ParseInt(baseTime, 10, 64); err == nil {
			db = db.Where("base_time = ?", value)
		}
	}
	if status := strings.TrimSpace(params["status"]); status != "" {
		if value, err := strconv.Atoi(status); err == nil {
			db = db.Where("status = ?", value)
		}
	}
	if ncStatus := strings.TrimSpace(params["ncStatus"]); ncStatus != "" {
		if value, err := strconv.Atoi(ncStatus); err == nil {
			db = db.Where("nc_status = ?", value)
		}
	}
	if content := strings.TrimSpace(params["content"]); content != "" {
		like := "%" + content + "%"
		db = db.Where("grid_name like ? or grid_identifier like ? or coordinate_system like ? or nc_file_name like ? or error_msg like ? or nc_error_msg like ?", like, like, like, like, like, like)
	}
	return db
}

func resolveGridDiffTaskOrder(params map[string]string) string {
	if asc := strings.TrimSpace(params["asc"]); asc != "" {
		if column, ok := gridDiffTaskOrderColumn(asc); ok {
			return fmt.Sprintf("%s asc", column)
		}
	}
	if desc := strings.TrimSpace(params["desc"]); desc != "" {
		if column, ok := gridDiffTaskOrderColumn(desc); ok {
			return fmt.Sprintf("%s desc", column)
		}
	}
	return "base_time desc, id desc"
}

func gridDiffTaskOrderColumn(field string) (string, bool) {
	switch utils.CamelToSnake(strings.TrimSpace(field)) {
	case "id":
		return "id", true
	case "grid_guid":
		return "grid_guid", true
	case "grid_name":
		return "grid_name", true
	case "grid_identifier":
		return "grid_identifier", true
	case "coordinate_system":
		return "coordinate_system", true
	case "base_time":
		return "base_time", true
	case "resolution":
		return "resolution", true
	case "point_count":
		return "point_count", true
	case "status":
		return "status", true
	case "nc_status":
		return "nc_status", true
	case "nc_file_size":
		return "nc_file_size", true
	case "create_time":
		return "create_time", true
	case "update_time":
		return "update_time", true
	default:
		return "", false
	}
}
