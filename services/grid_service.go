package services

import (
	"errors"
	"fmt"
	"hash/crc32"
	"nav-rain-grid-go/domains"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"
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
	entity.Resolution = normalizeGridResolution(entity.Resolution)
	entity.MinDevice = normalizeGridMinDevice(entity.MinDevice)
	entity.CoordinateSystem = normalizeGridCoordinateSystem(entity.CoordinateSystem)
	entity.GridIdentifier = normalizeGridIdentifier(entity.GridIdentifier, entity.Name)

	now := time.Now().UnixMilli()
	updateValues := map[string]interface{}{
		"name":              entity.Name,
		"grid_identifier":   entity.GridIdentifier,
		"coordinate_system": entity.CoordinateSystem,
		"sncodes":           strings.TrimSpace(entity.Sncodes),
		"resolution":        entity.Resolution,
		"min_device":        entity.MinDevice,
		"min_distance":      entity.MinDistance,
		"status":            entity.Status,
		"update_time":       now,
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

func normalizeGridResolution(resolution float64) float64 {
	if resolution <= 0 {
		return domains.DefaultGridResolution
	}
	return resolution
}

func normalizeGridCoordinateSystem(coordinateSystem string) string {
	value := strings.ToLower(strings.TrimSpace(coordinateSystem))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	if value == "" || value == "84" || value == "wgs84" || value == "wgs1984" {
		return domains.DefaultGridCoordinateSystem
	}
	normalized := buildGridIdentifier(value, false)
	if normalized == "" {
		return domains.DefaultGridCoordinateSystem
	}
	return normalized
}

func normalizeGridIdentifier(identifier string, gridName string) string {
	if normalized := buildGridIdentifier(identifier, false); normalized != "" {
		return normalized
	}
	if normalized := buildGridIdentifier(removeGridNameCommonWords(gridName), true); normalized != "" {
		return normalized
	}
	checksum := crc32.ChecksumIEEE([]byte(gridName))
	return fmt.Sprintf("grid_%08x", checksum)
}

func removeGridNameCommonWords(value string) string {
	result := strings.TrimSpace(value)
	for _, word := range []string{"格网配置", "网格配置", "格网", "网格", "降雨", "雨量", "预测", "配置"} {
		result = strings.ReplaceAll(result, word, "")
	}
	if strings.TrimSpace(result) == "" {
		return value
	}
	return result
}

func buildGridIdentifier(value string, compactChinese bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	args := pinyin.NewArgs()
	tokens := make([]string, 0)
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, r := range value {
		if isIdentifierASCII(r) {
			current.WriteRune(unicode.ToLower(r))
			continue
		}
		if pys := pinyin.SinglePinyin(r, args); len(pys) > 0 {
			if !compactChinese {
				flush()
			}
			current.WriteString(pys[0])
			if !compactChinese {
				flush()
			}
			continue
		}
		if r == '_' || r == '-' || unicode.IsSpace(r) {
			flush()
		}
	}
	flush()
	return strings.Join(tokens, "_")
}

func isIdentifierASCII(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
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
		if value, err := strconv.ParseFloat(resolution, 64); err == nil {
			db = db.Where("resolution = ?", value)
		}
	}
	if gridIdentifier := strings.TrimSpace(params["gridIdentifier"]); gridIdentifier != "" {
		db = db.Where("grid_identifier = ?", gridIdentifier)
	}
	if coordinateSystem := strings.TrimSpace(params["coordinateSystem"]); coordinateSystem != "" {
		db = db.Where("coordinate_system = ?", coordinateSystem)
	}
	if status := strings.TrimSpace(params["status"]); status != "" {
		db = db.Where("status = ?", status)
	}
	if content := strings.TrimSpace(params["content"]); content != "" {
		like := "%" + content + "%"
		db = db.Where("name like ? or grid_identifier like ? or coordinate_system like ? or sncodes like ? or resolution like ?", like, like, like, like, like)
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
	case "grid_identifier":
		return "grid_identifier", true
	case "coordinate_system":
		return "coordinate_system", true
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
