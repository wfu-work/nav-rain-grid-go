package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"mime/multipart"
	"nav-rain-grid-go/domains"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/wfu-work/nav-common-go-lib/global"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestVersionReleaseServiceUploadUsesConfiguredPath(t *testing.T) {
	oldDB := global.NAV_DB
	oldViper := global.NAV_VIPER
	defer func() {
		global.NAV_DB = oldDB
		global.NAV_VIPER = oldViper
	}()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&domains.VersionRelease{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	global.NAV_DB = db

	versionDir := t.TempDir()
	v := viper.New()
	v.Set("rain.version-path", versionDir)
	global.NAV_VIPER = v

	service := VersionReleaseService{}
	firstData := []byte("first version package")
	firstHeader := testVersionReleaseFileHeader(t, "rain-grid.zip", firstData)
	first, err := service.Upload(domains.VersionRelease{
		Version:      "v1.0.0",
		AppName:      "nav-rain-grid",
		Platform:     "linux",
		Architecture: "amd64",
	}, firstHeader)
	if err != nil {
		t.Fatalf("upload first release: %v", err)
	}
	assertVersionReleaseFile(t, versionDir, first, firstData)

	secondData := []byte("second version package")
	secondHeader := testVersionReleaseFileHeader(t, "rain-grid-v1.0.0.zip", secondData)
	second, err := service.Upload(domains.VersionRelease{
		Version:      "v1.0.0",
		AppName:      "nav-rain-grid",
		Platform:     "linux",
		Architecture: "amd64",
	}, secondHeader)
	if err != nil {
		t.Fatalf("upload second release: %v", err)
	}
	assertVersionReleaseFile(t, versionDir, second, secondData)
	if first.Guid != second.Guid {
		t.Fatalf("same version identity should update existing release: first=%s second=%s", first.Guid, second.Guid)
	}
	if _, err := os.Stat(first.FilePath); !os.IsNotExist(err) {
		t.Fatalf("old version file should be replaced, stat err=%v", err)
	}

	var count int64
	if err := db.Model(&domains.VersionRelease{}).Count(&count).Error; err != nil {
		t.Fatalf("count releases: %v", err)
	}
	if count != 1 {
		t.Fatalf("unexpected release count: got %d, want 1", count)
	}

	latest, err := service.LatestPublished(map[string]string{
		"appName":      "nav-rain-grid",
		"platform":     "linux",
		"architecture": "amd64",
	})
	if err != nil {
		t.Fatalf("query latest release: %v", err)
	}
	if latest.Guid != second.Guid {
		t.Fatalf("unexpected latest release: got %s, want %s", latest.Guid, second.Guid)
	}
}

func testVersionReleaseFileHeader(t *testing.T, fileName string, data []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/", &body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}
	files := req.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("unexpected multipart file count: %d", len(files))
	}
	return files[0]
}

func assertVersionReleaseFile(t *testing.T, versionDir string, release domains.VersionRelease, data []byte) {
	t.Helper()
	if release.Status != domains.VersionReleaseStatusPublished {
		t.Fatalf("unexpected release status: %d", release.Status)
	}
	if release.FileSize != int64(len(data)) {
		t.Fatalf("unexpected file size: got %d, want %d", release.FileSize, len(data))
	}
	sum := sha256.Sum256(data)
	if release.Checksum != hex.EncodeToString(sum[:]) {
		t.Fatalf("unexpected checksum: %s", release.Checksum)
	}
	rel, err := filepath.Rel(versionDir, release.FilePath)
	if err != nil {
		t.Fatalf("resolve release path: %v", err)
	}
	if rel == "." || rel == ".." || filepath.IsAbs(rel) || len(rel) >= 2 && rel[:2] == ".." {
		t.Fatalf("release file should be stored under configured path: %s", release.FilePath)
	}
	written, err := os.ReadFile(release.FilePath)
	if err != nil {
		t.Fatalf("read release file: %v", err)
	}
	if !bytes.Equal(written, data) {
		t.Fatalf("unexpected release file content: %q", string(written))
	}
}
