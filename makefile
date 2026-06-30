# 输出目录
BIN_DIR = build

# 程序名
APP_NAME = Gnss

# 日期
DATE := $(shell date +%Y%m%d)

# 版本号
VERSION = V1.0.0

# 本机 OS
ifeq ($(OS),Windows_NT)
    HOST_OS := windows
else
    HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
endif

# 支持的平台
PLATFORMS = linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64

# 如果不是 macOS，则去掉 darwin/amd64
ifeq ($(HOST_OS),darwin)
    PLATFORMS = darwin/amd64 darwin/arm64
else ifeq ($(HOST_OS),windows)
	PLATFORMS = windows/amd64
else ifeq ($(HOST_OS),linux)
	PLATFORMS = linux/amd64
else
    PLATFORMS := $(filter-out darwin/%,$(PLATFORMS))
endif

# 输出检查
print-platforms:
	@echo "HOST_OS     = $(HOST_OS)"
	@echo "PLATFORMS   = $(PLATFORMS)"

all: $(PLATFORMS)

$(PLATFORMS):
	@os=$$(echo $@ | cut -d/ -f1); \
	arch=$$(echo $@ | cut -d/ -f2); \
	output="$(BIN_DIR)/$(VERSION)/$(APP_NAME)-$${os}-$${arch}"; \
	if [ "$$os" = "windows" ]; then \
		output="$${output}.exe"; \
	fi; \
	echo "✅Building for $$os/$$arch, output: $$output"; \
	GOOS=$$os GOARCH=$$arch CGO_ENABLED=1 go build -ldflags "-X main.version=$(VERSION)" -o $$output main.go

win:
    output="$(BIN_DIR)/$(VERSION)/$(APP_NAME)-$${os}-$${arch}"; \
    output="$(output).exe"; \
    echo "✅Building for windows/amd64, output: $(output)"; \
    GOOS=$$os GOARCH=$$arch CGO_ENABLED=1 go build -ldflags "-X main.version=$(VERSION)" -o $$output main.go

linux:
    output="$(BIN_DIR)/$(VERSION)/$(APP_NAME)-$${os}-$${arch}"; \
    output="$(output)"; \
    echo "✅Building for linux/amd64, output: $(output)"; \
    GOOS=$$os GOARCH=$$arch CGO_ENABLED=1 go build -ldflags "-X main.version=$(VERSION)" -o $$output main.go

clean:
	rm -rf $(BIN_DIR)

.PHONY: all clean $(PLATFORMS)