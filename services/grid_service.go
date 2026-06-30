package services

import (
	"errors"
	"fmt"
	"nav-rain-grid-go/domains"
	"strings"
	"time"

	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/services"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
)

type GridService struct {
	services.CrudService[domains.Grid]
}

var GridServiceApp = new(GridService)

func (s GridService) SaveOrUpdate(entity domains.Grid) error {
	entity.Name = strings.TrimSpace(entity.Name)
	if entity.Name == "" {
		return errors.New("格网名称不能为空")
	}
	if global.NAV_DB == nil {
		return errors.New("database is not initialized")
	}

	now := time.Now().UnixMilli()
	updateValues := map[string]interface{}{
		"name":         entity.Name,
		"sncodes":      strings.TrimSpace(entity.Sncodes),
		"resolution":   strings.TrimSpace(entity.Resolution),
		"min_device":   entity.MinDevice,
		"min_distance": entity.MinDistance,
		"status":       entity.Status,
		"update_time":  now,
	}

	if strings.TrimSpace(entity.Guid) != "" {
		return global.NAV_DB.Model(&domains.Grid{}).
			Where("guid = ?", entity.Guid).
			Updates(updateValues).Error
	}

	var existing domains.Grid
	err := global.NAV_DB.Where("name = ?", entity.Name).First(&existing).Error
	if err == nil {
		return global.NAV_DB.Model(&existing).Updates(updateValues).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.Create(entity)
}

func (s GridService) List(params map[string]string) (list interface{}, total int64, err error) {
	if global.NAV_DB == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	db := s.buildGridQuery(params)

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

	var results []domains.Grid
	err = db.Order(resolveGridOrder(params)).
		Limit(size).
		Offset(size * (page - 1)).
		Find(&results).Error
	return results, total, err
}

func (s GridService) Query(params map[string]string) ([]domains.Grid, error) {
	if global.NAV_DB == nil {
		return nil, errors.New("database is not initialized")
	}
	var results []domains.Grid
	err := s.buildGridQuery(params).Order(resolveGridOrder(params)).Find(&results).Error
	return results, err
}

func (s GridService) buildGridQuery(params map[string]string) *gorm.DB {
	db := global.NAV_DB.Model(new(domains.Grid))

	if name := strings.TrimSpace(params["name"]); name != "" {
		db = db.Where("name like ?", "%"+name+"%")
	}
	if sncode := strings.TrimSpace(params["sncode"]); sncode != "" {
		db = db.Where("sncodes like ?", "%"+sncode+"%")
	}
	if resolution := strings.TrimSpace(params["resolution"]); resolution != "" {
		db = db.Where("resolution = ?", resolution)
	}
	if status := strings.TrimSpace(params["status"]); status != "" {
		db = db.Where("status = ?", status)
	}
	if content := strings.TrimSpace(params["content"]); content != "" {
		like := "%" + content + "%"
		db = db.Where("name like ? or sncodes like ? or resolution like ?", like, like, like)
	}
	return db
}

func resolveGridOrder(params map[string]string) string {
	if asc := strings.TrimSpace(params["asc"]); asc != "" {
		if column, ok := gridOrderColumn(asc); ok {
			return fmt.Sprintf("%s asc", column)
		}
	}
	if desc := strings.TrimSpace(params["desc"]); desc != "" {
		if column, ok := gridOrderColumn(desc); ok {
			return fmt.Sprintf("%s desc", column)
		}
	}
	return "id desc"
}

func gridOrderColumn(field string) (string, bool) {
	switch utils.CamelToSnake(strings.TrimSpace(field)) {
	case "id":
		return "id", true
	case "name":
		return "name", true
	case "resolution":
		return "resolution", true
	case "min_device":
		return "min_device", true
	case "min_distance":
		return "min_distance", true
	case "status":
		return "status", true
	case "create_time":
		return "create_time", true
	case "update_time":
		return "update_time", true
	default:
		return "", false
	}
}
