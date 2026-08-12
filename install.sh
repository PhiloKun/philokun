#!/usr/bin/env bash
# philokun 一键安装脚本
# 用法:
#   curl -sSfL https://raw.githubusercontent.com/philokun/philokun/main/install.sh | sh
#   # 或指定版本:
#   curl -sSfL https://raw.githubusercontent.com/philokun/philokun/main/install.sh | sh -s -- v1.0.0
#   # 国内加速（走 Gitee，需显式指定版本）:
#   curl -sSfL https://gitee.com/philokun/philokun/raw/main/install.sh | RELEASE_MIRROR=gitee sh -s -- v1.0.0
#
# 下载源由 RELEASE_MIRROR 控制: github(默认) / gitee(国内加速)。
# 脚本会自动识别操作系统/架构，从对应 Release 下载对应二进制，
# 安装到 ~/.local/bin（或 $INSTALL_DIR 指定的目录），并提示加入 PATH。

set -euo pipefail

REPO="philokun/philokun"
BINARY="philokun"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${1:-latest}"

# 颜色输出
info()  { printf '\033[32m[INFO]\033[0m %s\n' "$1"; }
warn()  { printf '\033[33m[WARN]\033[0m %s\n' "$1"; }
error() { printf '\033[31m[ERROR]\033[0m %s\n' "$1" >&2; exit 1; }

# 检测操作系统
detect_os() {
  case "$(uname -s)" in
    Linux)  echo "linux" ;;
    Darwin) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) error "不支持的操作系统: $(uname -s)" ;;
  esac
}

# 检测 CPU 架构
detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) error "不支持的架构: $(uname -m)" ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
EXT=""
if [ "$OS" = "windows" ]; then EXT=".exe"; fi

# 下载源：默认 github，国内用户可用 RELEASE_MIRROR=gitee 走 Gitee（国内加速）
RELEASE_MIRROR="${RELEASE_MIRROR:-github}"
case "$RELEASE_MIRROR" in
  github)
    GH_BASE="https://github.com/${REPO}/releases"
    if [ "$VERSION" = "latest" ]; then
      BASE="${GH_BASE}/latest/download"
    else
      BASE="${GH_BASE}/download/${VERSION}"
    fi
    ;;
  gitee)
    # Gitee Release 附件直链（国内访问稳定）
    GITEE_BASE="https://gitee.com/${REPO}/releases/download"
    if [ "$VERSION" = "latest" ]; then
      # Gitee 没有 latest 概念，latest 时回退到最新 tag 需用户显式指定，这里直接用 v 前缀提示
      error "Gitee 源不支持 latest，请显式指定版本，例如: RELEASE_MIRROR=gitee sh -s -- v1.0.0"
    else
      BASE="${GITEE_BASE}/${VERSION}"
    fi
    ;;
  *)
    error "未知的 RELEASE_MIRROR: ${RELEASE_MIRROR}（仅支持 github / gitee）"
    ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}${EXT}"
URL="${BASE}/${ASSET}"

info "下载源: ${RELEASE_MIRROR}"
info "检测到平台: ${OS}/${ARCH}"
info "下载地址: ${URL}"

# 创建安装目录
mkdir -p "$INSTALL_DIR"

# 下载（带镜像回退：优先 RELEASE_MIRROR，失败时尝试另一个源）
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

# 构造备用源 URL：github 失败则试 gitee，反之亦然
alt_url() {
  case "$RELEASE_MIRROR" in
    github)
      if [ "$VERSION" = "latest" ]; then
        echo "https://gitee.com/${REPO}/releases/download/${CURRENT_TAG:-v1.0.0}/${ASSET}"
      else
        echo "https://gitee.com/${REPO}/releases/download/${VERSION}/${ASSET}"
      fi
      ;;
    gitee)
      if [ "$VERSION" = "latest" ]; then
        echo "https://github.com/${REPO}/releases/latest/download/${ASSET}"
      else
        echo "https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
      fi
      ;;
  esac
}

download() {
  local u="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -sSLf "$u" -o "$TMP"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$TMP" "$u"
  else
    return 2
  fi
}

if ! download "$URL"; then
  FALLBACK="$(alt_url)"
  warn "主源下载失败，尝试备用源: ${FALLBACK}"
  if ! download "$FALLBACK"; then
    error "下载失败，请确认版本 ${VERSION} 已发布且包含资产 ${ASSET}"
  fi
  URL="$FALLBACK"
  info "已通过备用源下载"
fi

DEST="${INSTALL_DIR}/${BINARY}${EXT}"
install -m 0755 "$TMP" "$DEST"
info "已安装到: ${DEST}"

# 检查 PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    warn "${INSTALL_DIR} 不在 PATH 中，请执行以下命令之一："
    echo "    export PATH=\"\$PATH:${INSTALL_DIR}\""
    echo "    # 或写入 shell 配置: echo 'export PATH=\"\$PATH:${INSTALL_DIR}\"' >> ~/.$(basename "$SHELL")rc"
    ;;
esac

info "安装完成！运行 '${BINARY} version' 验证。"
