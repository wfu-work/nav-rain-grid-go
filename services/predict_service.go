package services

import (
	"errors"
	"fmt"
	"nav-rain-grid-go/domains"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/services"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type PredictService struct {
	services.CrudService[domains.Predict]
}

var PredictServiceApp = new(PredictService)

func (s PredictService) CreateOne(baseTime int64, entity domains.Predict) *domains.Predict {
	one := &domains.Predict{}
	err := global.NAV_DB.Where("base_time = ? and time = ? and sncode = ?", baseTime, entity.Time, entity.Sncode).First(one).Error
	if err == nil {
		err = global.NAV_DB.Model(one).Updates(map[string]interface{}{
			"base_time":          baseTime,
			"time":               entity.Time,
			"sncode":             entity.Sncode,
			"predict_rain":       entity.PredictRain,
			"predict_rain_level": entity.PredictRainLevel,
			"type":               entity.Type,
			"update_time":        time.Now().UnixMilli(),
		}).Error
		if err != nil {
			fmt.Println("数据库中覆盖雨量预测数据执行失败:", err)
			return nil
		}
		one, err = s.GetByGuid(one.Guid)
		return one
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("数据库中查询雨量预测数据执行失败:", err)
		return nil
	}
	if entity.Guid == "" {
		entity.Guid = strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	err = s.Create(entity)
	if err != nil {
		fmt.Println("数据库中插入雨量预测数据执行失败:", err)
		return nil
	}
	one, err = s.GetByGuid(entity.Guid)
	return one
}

func (s PredictService) DeleteByParams(params map[string]string) error {
	db := global.NAV_DB.Model(new(domains.Predict))
	hasCondition := false

	sncode := strings.TrimSpace(params["sncode"])
	if sncode != "" && sncode != "undefined" {
		db = db.Where("sncode = ?", sncode)
		hasCondition = true
	}
	time := strings.TrimSpace(params["time"])
	if time != "" && sncode != "undefined" {
		timeValue, err := strconv.ParseInt(time, 10, 64)
		if err != nil {
			return fmt.Errorf("time参数格式错误: %w", err)
		}
		db = db.Where("time = ?", timeValue)
		hasCondition = true
	}
	baseTime := strings.TrimSpace(params["baseTime"])
	if time != "" && sncode != "undefined" {
		timeValue, err := strconv.ParseInt(baseTime, 10, 64)
		if err != nil {
			return fmt.Errorf("time参数格式错误: %w", err)
		}
		db = db.Where("base_time = ?", timeValue)
		hasCondition = true
	}
	startTime := strings.TrimSpace(params["startTime"])
	if startTime != "" {
		timeValue, err := strconv.ParseInt(startTime, 10, 64)
		if err != nil {
			return fmt.Errorf("startTime参数格式错误: %w", err)
		}
		db = db.Where("time >= ?", timeValue)
		hasCondition = true
	}
	endTime := strings.TrimSpace(params["endTime"])
	if endTime != "" {
		timeValue, err := strconv.ParseInt(endTime, 10, 64)
		if err != nil {
			return fmt.Errorf("endTime参数格式错误: %w", err)
		}
		db = db.Where("time <= ?", timeValue)
		hasCondition = true
	}
	for key, value := range params {
		value = strings.TrimSpace(value)
		if key == "sncode" || key == "startTime" || key == "endTime" || value == "" || value == "undefined" {
			continue
		}
		db = db.Where(utils.CamelToSnake(key)+" = ?", value)
		hasCondition = true
	}
	if !hasCondition {
		return errors.New("删除参数不能为空")
	}
	return db.Delete(new(domains.Predict)).Error
}

func (s PredictService) List(params map[string]string) (list interface{}, total int64, err error) {
	limit := utils.Str2Int(params["size"])
	offset := utils.Str2Int(params["size"]) * (utils.Str2Int(params["page"]) - 1)
	err = s.buildPredictBaseQuery(params).Distinct("base_time").Count(&total).Error
	if err != nil {
		return
	}

	var baseTimes []int64
	err = s.buildPredictBaseQuery(params).
		Distinct("base_time").
		Order("base_time desc").
		Limit(limit).
		Offset(offset).
		Pluck("base_time", &baseTimes).Error
	if err != nil {
		return
	}
	if len(baseTimes) == 0 {
		return []domains.PredictGroup{}, total, nil
	}

	var results []domains.Predict
	err = s.buildPredictBaseQuery(params).
		Where("base_time IN ?", baseTimes).
		Order("base_time desc, time asc").
		Find(&results).Error
	if err != nil {
		return
	}
	return groupPredicts(results), total, nil
}

func (s PredictService) Query(params map[string]string) ([]domains.PredictGroup, error) {
	results, err := s.queryPredictDetails(params)
	if err != nil {
		return nil, err
	}
	return groupPredicts(results), nil
}

func (s PredictService) queryPredictDetails(params map[string]string) ([]domains.Predict, error) {
	var results []domains.Predict
	db := s.buildPredictBaseQuery(params)
	for key, value := range params {
		if key != "sncode" && key != "startTime" && key != "endTime" {
			db = db.Where(utils.CamelToSnake(key)+" = ?", value)
		}
	}
	err := db.Order("base_time ASC, time ASC").Find(&results).Error
	return results, err
}

func (s PredictService) buildPredictBaseQuery(params map[string]string) *gorm.DB {
	db := global.NAV_DB.Model(new(domains.Predict))
	sncode := params["sncode"]
	if sncode != "" {
		db = db.Where("sncode = ?", sncode)
	}
	startTime := params["startTime"]
	if startTime != "" {
		db = db.Where("base_time >= ?", startTime)
	}
	endTime := params["endTime"]
	if endTime != "" {
		db = db.Where("base_time <= ?", endTime)
	}
	return db
}

func groupPredicts(results []domains.Predict) []domains.PredictGroup {
	groups := make([]domains.PredictGroup, 0)
	groupIndex := make(map[int64]int)
	for _, item := range results {
		index, ok := groupIndex[item.BaseTime]
		if !ok {
			groups = append(groups, domains.PredictGroup{
				BaseTime:    item.BaseTime,
				PredictList: []domains.Predict{},
			})
			index = len(groups) - 1
			groupIndex[item.BaseTime] = index
		}
		groups[index].PredictList = append(groups[index].PredictList, item)
	}
	return groups
}

func (s PredictService) updateRealRain(timestampMillis int64, off int, rainFallNow float64) error {
	baseTime := time.UnixMilli(timestampMillis).Add(-time.Duration(off) * time.Hour).UnixMilli()
	return global.NAV_DB.Model(&domains.Predict{}).
		Where("base_time = ? AND time = ?", baseTime, timestampMillis).
		Updates(map[string]interface{}{
			"real_rain":       rainFallNow,
			"real_rain_level": utils.GetLevel(rainFallNow),
			"update_time":     time.Now().UnixMilli(),
		}).Error
}

func (s PredictService) Export(params map[string]string, c *gin.Context) error {
	results, err := s.queryPredictDetails(params)
	if err != nil {
		return err
	}
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("关闭Excel文件失败:", err)
		}
	}()
	headers := []string{"时间", "预测降雨量"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue("Sheet1", cell, header)
	}
	for i, data := range results {
		row := i + 2
		values := []interface{}{
			time.UnixMilli(data.Time).Format("2006-01-02 15:04:05"),
			data.PredictRain,
		}
		for j, value := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			_ = f.SetCellValue("Sheet1", cell, value)
		}
	}
	filename := "气象预测数据.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Cache-Control", "no-cache")
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "导出失败"})
		return fmt.Errorf("写入Excel文件到响应失败: %v", err)
	}
	return nil
}

func getRainCategory(level int) string {
	switch level {
	case 0:
		return "无雨"
	case 1:
		return "小雨"
	case 2:
		return "中雨"
	case 3:
		return "大雨"
	default:
		return "未知"
	}
}
