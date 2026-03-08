#!/usr/bin/env bash
set -euo pipefail

APP_NAME="aurora-ums-service"
SERVICE_NAME="aurora-ums"
SERVICE_FILE_NAME="aurora-ums.service"
INSTALL_DIR="/opt/aurora/ums"
BIN_PATH="${INSTALL_DIR}/${APP_NAME}"
SYSTEMD_PATH="/etc/systemd/system/${SERVICE_FILE_NAME}"

TLS_CERT_PATH="/etc/aurora/certs/ums.crt"
TLS_KEY_PATH="/etc/aurora/certs/ums.key"
PURGE_CERTS=0

log() {
  printf '[ums-uninstall] %s\n' "$*"
}

die() {
  printf '[ums-uninstall][error] %s\n' "$*" >&2
  exit 1
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

usage() {
  cat <<'EOF'
Usage:
  ./uninstall.sh [options]

Options:
  --purge-certs   Remove /etc/aurora/certs/ums.crt and /etc/aurora/certs/ums.key
  -h, --help      Show help
EOF
}

parse_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --purge-certs)
        PURGE_CERTS=1
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

stop_and_disable_service() {
  if command -v systemctl >/dev/null 2>&1; then
    if systemctl list-unit-files | grep -q "^${SERVICE_FILE_NAME}"; then
      log "stopping service ${SERVICE_NAME}"
      systemctl stop "${SERVICE_NAME}" || true
      log "disabling service ${SERVICE_NAME}"
      systemctl disable "${SERVICE_NAME}" || true
      systemctl reset-failed "${SERVICE_NAME}" || true
    fi
  fi
}

remove_systemd_unit() {
  if [ -f "$SYSTEMD_PATH" ]; then
    log "removing unit ${SYSTEMD_PATH}"
    rm -f "$SYSTEMD_PATH"
  fi
  if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
  fi
}

remove_binaries() {
  if [ -f "$BIN_PATH" ]; then
    log "removing binary ${BIN_PATH}"
    rm -f "$BIN_PATH"
  fi
  if [ -d "$INSTALL_DIR" ]; then
    log "removing install dir ${INSTALL_DIR}"
    rm -rf "$INSTALL_DIR"
  fi
}

remove_tls_if_requested() {
  if [ "$PURGE_CERTS" -ne 1 ]; then
    return
  fi
  log "purging service tls files"
  rm -f "$TLS_CERT_PATH" "$TLS_KEY_PATH"
}

main() {
  parse_args "$@"
  ensure_root "$@"
  stop_and_disable_service
  remove_systemd_unit
  remove_binaries
  remove_tls_if_requested
  log "uninstall completed: ${SERVICE_NAME}"
}

main "$@"
