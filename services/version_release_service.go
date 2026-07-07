package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"nav-rain-grid-go/configs"
	"nav-rain-grid-go/domains"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/wfu-work/nav-common-go-lib/global"
	"github.com/wfu-work/nav-common-go-lib/services"
	"github.com/wfu-work/nav-common-go-lib/utils"
	"gorm.io/gorm"
)

const defaultVersionReleasePath = configs.DefaultVersionPath

type VersionReleaseService struct {
	services.CrudService[domains.VersionRelease]
}

type versionReleaseFile struct {
	Path     string
	Name     string
	Size     int64
	Checksum string
}

var VersionReleaseServiceApp = new(VersionReleaseService)

func (s VersionReleaseService) SaveOrUpdate(entity domains.VersionRelease) error {
	if global.NAV_DB == nil {
		return errors.New("database is not initialized")
	}
	entity = normalizeVersionRelease(entity)
	if entity.Version == "" {
		return errors.New("版本号不能为空")
	}

	updateValues := versionReleaseMetadataUpdates(entity)
	if strings.TrimSpace(entity.Guid) != "" {
		return global.NAV_DB.Model(&domains.VersionRelease{}).
			Where("guid = ?", entity.Guid).
			Updates(updateValues).Error
	}

	var existing domains.VersionRelease
	err := versionReleaseIdentityQuery(global.NAV_DB, entity).First(&existing).Error
	if err == nil {
		return global.NAV_DB.Model(&existing).Updates(updateValues).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.Create(entity)
}

func (s VersionReleaseService) Upload(entity domains.VersionRelease, fileHeader *multipart.FileHeader) (domains.VersionRelease, error) {
	if global.NAV_DB == nil {
		return domains.VersionRelease{}, errors.New("database is not initialized")
	}
	if fileHeader == nil {
		return domains.VersionRelease{}, errors.New("版本文件不能为空")
	}
	entity = normalizeVersionRelease(entity)
	if entity.Version == "" {
		return domains.VersionRelease{}, errors.New("版本号不能为空")
	}
	if entity.Status != domains.VersionReleaseStatusDisabled {
		entity.Status = domains.VersionReleaseStatusPublished
	}
	if entity.ReleaseTime <= 0 {
		entity.ReleaseTime = time.Now().UnixMilli()
	}

	fileInfo, err := saveVersionReleaseFile(entity.Version, fileHeader)
	if err != nil {
		return domains.VersionRelease{}, err
	}
	entity.FilePath = fileInfo.Path
	entity.FileName = fileInfo.Name
	entity.FileSize = fileInfo.Size
	entity.Checksum = fileInfo.Checksum

	result, oldFilePath, err := s.upsertUploadedRelease(entity)
	if err != nil {
		_ = os.Remove(fileInfo.Path)
		return domains.VersionRelease{}, err
	}
	if oldFilePath != "" && oldFilePath != fileInfo.Path {
		_ = removeVersionReleaseFile(oldFilePath)
	}
	return result, nil
}

func (s VersionReleaseService) DeleteByGuid(guid string) error {
	if global.NAV_DB == nil {
		return errors.New("database is not initialized")
	}
	var release domains.VersionRelease
	if err := global.NAV_DB.Where("guid = ?", strings.TrimSpace(guid)).First(&release).Error; err != nil {
		return err
	}
	if strings.TrimSpace(release.FilePath) != "" {
		_ = removeVersionReleaseFile(release.FilePath)
	}
	return s.CrudService.DeleteByGuid(guid)
}

func (s VersionReleaseService) LatestPublished(params map[string]string) (domains.VersionRelease, error) {
	if global.NAV_DB == nil {
		return domains.VersionRelease{}, errors.New("database is not initialized")
	}
	var result domains.VersionRelease
	err := s.buildVersionReleaseQuery(params).
		Where("status = ?", domains.VersionReleaseStatusPublished).
		Where("file_path <> ''").
		Order("release_time desc, update_time desc, id desc").
		First(&result).Error
	return result, err
}

func (s VersionReleaseService) List(params map[string]string) (list interface{}, total int64, err error) {
	if global.NAV_DB == nil {
		return nil, 0, errors.New("database is not initialized")
	}
	db := s.buildVersionReleaseQuery(params)

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

	var results []domains.VersionRelease
	err = db.Order(resolveVersionReleaseOrder(params)).
		Limit(size).
		Offset(size * (page - 1)).
		Find(&results).Error
	return results, total, err
}

func (s VersionReleaseService) Query(params map[string]string) ([]domains.VersionRelease, error) {
	if global.NAV_DB == nil {
		return nil, errors.New("database is not initialized")
	}
	var results []domains.VersionRelease
	err := s.buildVersionReleaseQuery(params).Order(resolveVersionReleaseOrder(params)).Find(&results).Error
	return results, err
}

func (s VersionReleaseService) upsertUploadedRelease(entity domains.VersionRelease) (domains.VersionRelease, string, error) {
	var result domains.VersionRelease
	var oldFilePath string
	err := global.NAV_DB.Transaction(func(tx *gorm.DB) error {
		var existing domains.VersionRelease
		var queryErr error
		if strings.TrimSpace(entity.Guid) != "" {
			queryErr = tx.Where("guid = ?", entity.Guid).First(&existing).Error
			if queryErr != nil {
				return queryErr
			}
		} else {
			queryErr = versionReleaseIdentityQuery(tx, entity).First(&existing).Error
		}

		updateValues := versionReleaseMetadataUpdates(entity)
		updateValues["file_path"] = entity.FilePath
		updateValues["file_name"] = entity.FileName
		updateValues["file_size"] = entity.FileSize
		updateValues["checksum"] = entity.Checksum

		if queryErr == nil {
			oldFilePath = strings.TrimSpace(existing.FilePath)
			if err := tx.Model(&existing).Updates(updateValues).Error; err != nil {
				return err
			}
			return tx.Where("guid = ?", existing.Guid).First(&result).Error
		}
		if !errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return queryErr
		}
		if err := tx.Create(&entity).Error; err != nil {
			return err
		}
		result = entity
		return nil
	})
	return result, oldFilePath, err
}

func (s VersionReleaseService) buildVersionReleaseQuery(params map[string]string) *gorm.DB {
	db := global.NAV_DB.Model(new(domains.VersionRelease))

	if version := strings.TrimSpace(params["version"]); version != "" {
		db = db.Where("version_no = ?", version)
	}
	if name := strings.TrimSpace(params["name"]); name != "" {
		db = db.Where("name like ?", "%"+name+"%")
	}
	if appName := strings.TrimSpace(params["appName"]); appName != "" {
		db = db.Where("app_name = ?", appName)
	}
	if platform := strings.TrimSpace(params["platform"]); platform != "" {
		db = db.Where("platform = ?", platform)
	}
	if architecture := strings.TrimSpace(params["architecture"]); architecture != "" {
		db = db.Where("architecture = ?", architecture)
	}
	if status := strings.TrimSpace(params["status"]); status != "" {
		if value, err := strconv.Atoi(status); err == nil {
			db = db.Where("status = ?", value)
		}
	}
	if content := strings.TrimSpace(params["content"]); content != "" {
		like := "%" + content + "%"
		db = db.Where("version_no like ? or name like ? or app_name like ? or platform like ? or architecture like ? or file_name like ? or description like ? or release_note like ?", like, like, like, like, like, like, like, like)
	}
	return db
}

func normalizeVersionRelease(entity domains.VersionRelease) domains.VersionRelease {
	entity.Guid = strings.TrimSpace(entity.Guid)
	entity.Version = strings.TrimSpace(entity.Version)
	entity.Name = strings.TrimSpace(entity.Name)
	entity.AppName = strings.TrimSpace(entity.AppName)
	entity.Platform = strings.ToLower(strings.TrimSpace(entity.Platform))
	entity.Architecture = strings.ToLower(strings.TrimSpace(entity.Architecture))
	entity.Description = strings.TrimSpace(entity.Description)
	entity.ReleaseNote = strings.TrimSpace(entity.ReleaseNote)
	if entity.Name == "" {
		entity.Name = entity.Version
	}
	if entity.Status != domains.VersionReleaseStatusDraft &&
		entity.Status != domains.VersionReleaseStatusPublished &&
		entity.Status != domains.VersionReleaseStatusDisabled {
		entity.Status = domains.VersionReleaseStatusPublished
	}
	if entity.ReleaseTime <= 0 && entity.Status == domains.VersionReleaseStatusPublished {
		entity.ReleaseTime = time.Now().UnixMilli()
	}
	return entity
}

func versionReleaseMetadataUpdates(entity domains.VersionRelease) map[string]interface{} {
	return map[string]interface{}{
		"version_no":   entity.Version,
		"name":         entity.Name,
		"app_name":     entity.AppName,
		"platform":     entity.Platform,
		"architecture": entity.Architecture,
		"description":  entity.Description,
		"release_note": entity.ReleaseNote,
		"status":       entity.Status,
		"release_time": entity.ReleaseTime,
		"update_time":  time.Now().UnixMilli(),
	}
}

func versionReleaseIdentityQuery(db *gorm.DB, entity domains.VersionRelease) *gorm.DB {
	return db.Where("version_no = ? AND app_name = ? AND platform = ? AND architecture = ?", entity.Version, entity.AppName, entity.Platform, entity.Architecture)
}

func saveVersionReleaseFile(version string, fileHeader *multipart.FileHeader) (versionReleaseFile, error) {
	baseDir := versionReleaseBaseDir()
	versionDir := filepath.Join(baseDir, safeVersionPathPart(version))
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return versionReleaseFile{}, err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return versionReleaseFile{}, err
	}
	defer src.Close()

	originalName := safeVersionFileName(fileHeader.Filename)
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), originalName)
	tmp, err := os.CreateTemp(versionDir, ".upload-*")
	if err != nil {
		return versionReleaseFile{}, err
	}
	tmpPath := tmp.Name()
	hasher := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(tmp, hasher), src)
	closeErr := tmp.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return versionReleaseFile{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return versionReleaseFile{}, closeErr
	}

	filePath := filepath.Join(versionDir, fileName)
	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath)
		return versionReleaseFile{}, err
	}

	return versionReleaseFile{
		Path:     filepath.ToSlash(filePath),
		Name:     originalName,
		Size:     size,
		Checksum: hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

func versionReleaseBaseDir() string {
	path := defaultVersionReleasePath
	if global.NAV_VIPER != nil {
		if configured := strings.TrimSpace(global.NAV_VIPER.GetString("rain.version-path")); configured != "" {
			path = configured
		}
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return filepath.Clean(path)
}

func safeVersionPathPart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('_')
	}
	result := strings.Trim(builder.String(), "._-")
	if result == "" {
		return "unknown"
	}
	return result
}

func safeVersionFileName(value string) string {
	value = strings.TrimSpace(filepath.Base(value))
	if value == "." || value == string(filepath.Separator) || value == "" {
		return "version.bin"
	}
	value = strings.ReplaceAll(value, string(filepath.Separator), "_")
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "\\", "_")
	if strings.Trim(value, "._- ") == "" {
		return "version.bin"
	}
	return value
}

func removeVersionReleaseFile(filePath string) error {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	baseDir := versionReleaseBaseDir()
	rel, err := filepath.Rel(baseDir, absFilePath)
	if err != nil {
		return err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return errors.New("版本文件不在版本目录内，拒绝删除")
	}
	if err := os.Remove(absFilePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func resolveVersionReleaseOrder(params map[string]string) string {
	if asc := strings.TrimSpace(params["asc"]); asc != "" {
		if column, ok := versionReleaseOrderColumn(asc); ok {
			return fmt.Sprintf("%s asc", column)
		}
	}
	if desc := strings.TrimSpace(params["desc"]); desc != "" {
		if column, ok := versionReleaseOrderColumn(desc); ok {
			return fmt.Sprintf("%s desc", column)
		}
	}
	return "release_time desc, id desc"
}

func versionReleaseOrderColumn(field string) (string, bool) {
	switch utils.CamelToSnake(strings.TrimSpace(field)) {
	case "id":
		return "id", true
	case "version", "version_no":
		return "version_no", true
	case "name":
		return "name", true
	case "app_name":
		return "app_name", true
	case "platform":
		return "platform", true
	case "architecture":
		return "architecture", true
	case "file_name":
		return "file_name", true
	case "file_size":
		return "file_size", true
	case "status":
		return "status", true
	case "release_time":
		return "release_time", true
	case "create_time":
		return "create_time", true
	case "update_time":
		return "update_time", true
	default:
		return "", false
	}
}
