#!/bin/sh
set -eu

APP_NAME="nav-rain-grid"
BINARY_NAME="navRainGridApp"
VERSION="latest"
INSTALL_DIR="/opt/nav-rain-grid"
SERVICE_NAME="nav-rain-grid"
API_BASE=""
LATEST_URL=""
DOWNLOAD_URL=""
CONFIG_URL=""
EXTRA_ARGS=""
FORCE_CONFIG="0"

usage() {
  cat <<'EOF'
Nav Rain Grid server installer

Usage:
  sh install-server.sh [options]

Options:
  --api-base URL           Backend API base URL, for example https://host/api
  --latest-url URL         Full latest-version API URL. Overrides --api-base
  --download-url URL       Direct binary download URL. Overrides version API lookup
  --version VERSION        Version number filter for version API, default latest
  --app-name NAME          Version release appName filter, default nav-rain-grid
  --binary-name NAME       Installed binary name, default navRainGridApp
  --install-dir DIR        Install directory, default /opt/nav-rain-grid
  --service-name NAME      systemd service name, default nav-rain-grid
  --config-url URL         Optional config.yaml download URL
  --force-config           Overwrite existing config.yaml when --config-url is set
  --extra-args "ARGS"      Extra server command arguments
  -h, --help               Show help

Examples:
  curl -fsSL https://example.com/install-server.sh | sudo sh -s -- \
    --api-base https://example.com/api

  curl -fsSL https://example.com/install-server.sh | sudo sh -s -- \
    --download-url https://example.com/api/version-release/GUID/download
EOF
}

log() {
  printf '%s\n' "==> $*"
}

die() {
  printf '%s\n' "ERROR: $*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

fetch_url() {
  url="$1"
  curl -fsSL --retry 3 --connect-timeout 15 "$url"
}

download_to() {
  url="$1"
  output="$2"
  curl -fL --retry 3 --connect-timeout 15 -o "$output" "$url"
}

json_string() {
  key="$1"
  input="$2"
  printf '%s' "$input" |
    tr '\n' ' ' |
    sed -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p" |
    sed 's#\\/#/#g; s#\\"#"#g; s#\\\\#\\#g' |
    sed -n '1p'
}

checksum_file() {
  file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    set -- $(sha256sum "$file")
    printf '%s\n' "$1"
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    set -- $(shasum -a 256 "$file")
    printf '%s\n' "$1"
    return 0
  fi
  return 1
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --api-base)
      API_BASE="${2:-}"
      shift 2
      ;;
    --latest-url)
      LATEST_URL="${2:-}"
      shift 2
      ;;
    --download-url)
      DOWNLOAD_URL="${2:-}"
      shift 2
      ;;
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --app-name)
      APP_NAME="${2:-}"
      shift 2
      ;;
    --binary-name)
      BINARY_NAME="${2:-}"
      shift 2
      ;;
    --install-dir)
      INSTALL_DIR="${2:-}"
      shift 2
      ;;
    --service-name)
      SERVICE_NAME="${2:-}"
      shift 2
      ;;
    --config-url)
      CONFIG_URL="${2:-}"
      shift 2
      ;;
    --force-config)
      FORCE_CONFIG="1"
      shift
      ;;
    --extra-args)
      EXTRA_ARGS="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown option: $1"
      ;;
  esac
done

[ "$(id -u)" -eq 0 ] || die "please run as root, for example: curl ... | sudo sh"
[ -n "$BINARY_NAME" ] || die "--binary-name is required"
[ -n "$INSTALL_DIR" ] || die "--install-dir is required"
[ -n "$SERVICE_NAME" ] || die "--service-name is required"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)
    GOOS="linux"
    ;;
  *)
    die "unsupported OS: $OS. This installer currently supports Linux systemd hosts."
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    GOARCH="amd64"
    ;;
  aarch64|arm64)
    GOARCH="arm64"
    ;;
  *)
    die "unsupported architecture: $ARCH"
    ;;
esac

need_cmd curl
need_cmd install
need_cmd systemctl

CHECKSUM=""
if [ -z "$DOWNLOAD_URL" ]; then
  if [ -z "$LATEST_URL" ]; then
    [ -n "$API_BASE" ] || die "set --api-base or --download-url"
    LATEST_URL="${API_BASE%/}/version-release/latest?appName=${APP_NAME}&platform=${GOOS}&architecture=${GOARCH}"
    if [ -n "$VERSION" ] && [ "$VERSION" != "latest" ]; then
      LATEST_URL="${LATEST_URL}&version=${VERSION}"
    fi
  fi

  log "Resolving latest version from ${LATEST_URL}"
  LATEST_JSON="$(fetch_url "$LATEST_URL")" || die "latest version lookup failed"
  DOWNLOAD_URL="$(json_string "downloadUrl" "$LATEST_JSON")"
  CHECKSUM="$(json_string "checksum" "$LATEST_JSON")"
  [ -n "$DOWNLOAD_URL" ] || die "downloadUrl not found in latest version response"
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

log "Downloading ${BINARY_NAME} for ${GOOS}/${GOARCH}"
download_to "$DOWNLOAD_URL" "${TMP_DIR}/${BINARY_NAME}" || die "download failed: ${DOWNLOAD_URL}"
[ -s "${TMP_DIR}/${BINARY_NAME}" ] || die "downloaded binary is empty"

if [ -n "$CHECKSUM" ]; then
  if CALCULATED_CHECKSUM="$(checksum_file "${TMP_DIR}/${BINARY_NAME}")"; then
    EXPECTED="$(printf '%s' "$CHECKSUM" | tr '[:upper:]' '[:lower:]')"
    CALCULATED="$(printf '%s' "$CALCULATED_CHECKSUM" | tr '[:upper:]' '[:lower:]')"
    [ "$EXPECTED" = "$CALCULATED" ] || die "checksum mismatch: expected ${EXPECTED}, got ${CALCULATED}"
    log "Checksum verified"
  else
    log "Checksum is provided but sha256sum/shasum is unavailable; skipping verification"
  fi
fi

log "Installing ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
install -d -m 0755 "$INSTALL_DIR"
install -d -m 0755 "${INSTALL_DIR}/data"
install -d -m 0755 "${INSTALL_DIR}/logback"

if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
  log "Stopping existing ${SERVICE_NAME} service"
  systemctl stop "$SERVICE_NAME"
fi

install -m 0755 "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
ln -sf "${INSTALL_DIR}/${BINARY_NAME}" "/usr/local/bin/${BINARY_NAME}"

if [ -n "$CONFIG_URL" ]; then
  CONFIG_TARGET="${INSTALL_DIR}/config.yaml"
  if [ ! -f "$CONFIG_TARGET" ] || [ "$FORCE_CONFIG" = "1" ]; then
    log "Downloading config.yaml"
    download_to "$CONFIG_URL" "${TMP_DIR}/config.yaml" || die "config download failed: ${CONFIG_URL}"
    [ -s "${TMP_DIR}/config.yaml" ] || die "downloaded config is empty"
    install -m 0644 "${TMP_DIR}/config.yaml" "$CONFIG_TARGET"
  else
    log "Keeping existing ${CONFIG_TARGET}"
  fi
fi

SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
log "Writing systemd service ${SERVICE_FILE}"
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Nav Rain Grid Server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${BINARY_NAME} ${EXTRA_ARGS}
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

log "Starting ${SERVICE_NAME} service"
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl restart "$SERVICE_NAME"

log "Installed successfully"
systemctl --no-pager --full status "$SERVICE_NAME" || true
