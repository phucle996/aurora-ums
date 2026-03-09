#!/usr/bin/env bash
set -euo pipefail

APP_NAME="aurora-ums-service"
SERVICE_NAME="aurora-ums"
SERVICE_FILE_NAME="aurora-ums.service"
INSTALL_DIR="/opt/aurora/ums"
BIN_PATH="${INSTALL_DIR}/${APP_NAME}"
SYSTEMD_PATH="/etc/systemd/system/${SERVICE_FILE_NAME}"

REPO_SLUG="phucle996/aurora-ums"
RELEASE_VERSION="${RELEASE_VERSION:-}"
NO_START=0
APP_HOSTNAME="${APP_HOSTNAME:-ums.aurora.local}"
BACKEND_PORT="${AURORA_UMS_BACKEND_PORT:-3005}"
ADMIN_RPC_ENDPOINT="${ADMIN_RPC_ENDPOINT:-}"
TLS_CERT_PATH="/etc/aurora/certs/ums.crt"
TLS_KEY_PATH="/etc/aurora/certs/ums.key"
TLS_CA_PATH="/etc/aurora/certs/ca.crt"
NGINX_CONF_PATH="/etc/nginx/conf.d/aurora-ums.conf"
UMS_ENV_FILE="/etc/aurora/ums.env"

log() {
  printf '[ums-install] %s\n' "$*"
}

die() {
  printf '[ums-install][error] %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

fetch_url() {
  local url="$1"
  local out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$out"
    return
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -qO "$out" "$url"
    return
  fi
  die "curl/wget is required"
}

ensure_root() {
  if [ "$(id -u)" -eq 0 ]; then
    return
  fi
  if command -v sudo >/dev/null 2>&1; then
    log "re-running with sudo"
    exec sudo -E bash "$0" "$@"
  fi
  die "must run as root (or have sudo)"
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) die "unsupported architecture: $arch" ;;
  esac
}

resolve_latest_version() {
  local tmp
  tmp="$(mktemp)"
  fetch_url "https://api.github.com/repos/${REPO_SLUG}/releases/latest" "$tmp"
  local tag
  tag="$(grep -m1 '"tag_name"' "$tmp" | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/')"
  rm -f "$tmp"
  [ -n "$tag" ] || die "cannot resolve latest release tag from ${REPO_SLUG}"
  printf '%s' "$tag"
}

ensure_user() {
  if id -u aurora >/dev/null 2>&1; then
    return
  fi
  log "creating linux user aurora"
  useradd --system --home /home/aurora --create-home --shell /usr/sbin/nologin aurora
}

ensure_tls_materials() {
  [ -f "$TLS_CERT_PATH" ] || die "tls cert not found: $TLS_CERT_PATH"
  [ -f "$TLS_KEY_PATH" ] || die "tls key not found: $TLS_KEY_PATH"
  [ -f "$TLS_CA_PATH" ] || die "tls ca not found: $TLS_CA_PATH"
}

install_binary() {
  local arch version tar_name download_url tmp_dir tar_path
  arch="$(detect_arch)"
  version="$RELEASE_VERSION"
  if [ -z "$version" ]; then
    version="$(resolve_latest_version)"
  fi

  tar_name="${APP_NAME}_linux_${arch}.tar.gz"
  download_url="https://github.com/${REPO_SLUG}/releases/download/${version}/${tar_name}"
  log "downloading release ${version} (${arch})"

  tmp_dir="$(mktemp -d)"
  tar_path="${tmp_dir}/${tar_name}"
  fetch_url "$download_url" "$tar_path"

  mkdir -p "$INSTALL_DIR"
  tar -xzf "$tar_path" -C "$tmp_dir"
  install -m 0755 "${tmp_dir}/${APP_NAME}_linux_${arch}" "$BIN_PATH"
  chown -R aurora:aurora "$INSTALL_DIR"
  rm -rf "$tmp_dir"
}

install_systemd_unit() {
  local tmp_unit unit_url
  tmp_unit="$(mktemp)"
  unit_url="https://raw.githubusercontent.com/${REPO_SLUG}/main/install/${SERVICE_FILE_NAME}"
  fetch_url "$unit_url" "$tmp_unit"
  install -m 0644 "$tmp_unit" "$SYSTEMD_PATH"
  rm -f "$tmp_unit"
}

write_runtime_env() {
  [ -n "$ADMIN_RPC_ENDPOINT" ] || die "admin rpc endpoint is required (--admin-rpc-endpoint)"
  mkdir -p "$(dirname "$UMS_ENV_FILE")"
  cat >"$UMS_ENV_FILE" <<EOF
ADMIN_RPC_ENDPOINT=${ADMIN_RPC_ENDPOINT}
EOF
  chmod 0600 "$UMS_ENV_FILE"
}

ensure_nginx() {
  if command -v nginx >/dev/null 2>&1; then
    return
  fi
  if command -v apt-get >/dev/null 2>&1; then
    apt-get install -y nginx
    return
  fi
  if command -v dnf >/dev/null 2>&1; then
    dnf install -y nginx
    return
  fi
  if command -v yum >/dev/null 2>&1; then
    yum install -y nginx
    return
  fi
  if command -v apk >/dev/null 2>&1; then
    apk add --no-cache nginx
    return
  fi
  die "nginx is required but cannot be installed automatically"
}

install_nginx_proxy() {
  ensure_nginx
  local tmp_conf template_url
  tmp_conf="$(mktemp)"
  template_url="https://raw.githubusercontent.com/${REPO_SLUG}/main/install/nginx.conf"
  fetch_url "$template_url" "$tmp_conf"
  sed -i \
    -e "s|__SERVER_NAME__|${APP_HOSTNAME}|g" \
    -e "s|__TLS_CERT_FILE__|${TLS_CERT_PATH}|g" \
    -e "s|__TLS_KEY_FILE__|${TLS_KEY_PATH}|g" \
    -e "s|__TLS_CA_FILE__|${TLS_CA_PATH}|g" \
    -e "s|__BACKEND_PORT__|${BACKEND_PORT}|g" \
    "$tmp_conf"
  install -m 0644 "$tmp_conf" "$NGINX_CONF_PATH"
  rm -f "$tmp_conf"

  nginx -t
  systemctl enable nginx
  systemctl restart nginx
}

restart_service() {
  systemctl daemon-reload
  systemctl enable "$SERVICE_NAME"
  if [ "$NO_START" -eq 0 ]; then
    systemctl restart "$SERVICE_NAME"
    systemctl --no-pager --full status "$SERVICE_NAME" || true
  fi
}

usage() {
  cat <<'EOF'
Usage:
  ./install.sh [options]

Options:
  -v <version>             Release tag (default: latest)
  -r <repo>                GitHub repo slug (default: phucle996/aurora-ums)
  --app-host <hostname>    Public hostname used by nginx (default: ums.aurora.local)
  --admin-rpc-endpoint <host:port>  Admin gRPC endpoint for bootstrap (required)
  --no-start               Do not restart service after install
  -h, --help               Show help
EOF
}

parse_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      -v)
        [ "$#" -ge 2 ] || die "missing value for -v"
        RELEASE_VERSION="$2"
        shift 2
        ;;
      -r)
        [ "$#" -ge 2 ] || die "missing value for -r"
        REPO_SLUG="$2"
        shift 2
        ;;
      --app-host)
        [ "$#" -ge 2 ] || die "missing value for --app-host"
        APP_HOSTNAME="$2"
        shift 2
        ;;
      --admin-rpc-endpoint)
        [ "$#" -ge 2 ] || die "missing value for --admin-rpc-endpoint"
        ADMIN_RPC_ENDPOINT="$2"
        shift 2
        ;;
      --no-start)
        NO_START=1
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        die "unknown argument: $1"
        ;;
    esac
  done
}

main() {
  parse_args "$@"
  ensure_root "$@"
  need_cmd tar
  need_cmd systemctl
  ensure_user
  install_binary
  install_systemd_unit
  write_runtime_env
  ensure_tls_materials
  install_nginx_proxy
  restart_service
  log "install completed: ${SERVICE_NAME}"
}

main "$@"
