#!/usr/bin/env bash
# philokun 一键安装脚本
# 用法:
#   curl -sSfL https://raw.githubusercontent.com/philokun/philokun/main/install.sh | sh
#   # 或指定版本:
#   curl -sSfL https://raw.githubusercontent.com/philokun/philokun/main/install.sh | sh -s -- v1.0.0
#
# 脚本会自动识别操作系统/架构，从 GitHub Release 下载对应二进制，
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

# 解析版本对应的下载 URL
if [ "$VERSION" = "latest" ]; then
  TAG="latest"
  BASE="https://github.com/${REPO}/releases/latest/download"
else
  TAG="$VERSION"
  BASE="https://github.com/${REPO}/releases/download/${VERSION}"
fi

ASSET="${BINARY}-${OS}-${ARCH}${EXT}"
URL="${BASE}/${ASSET}"

info "检测到平台: ${OS}/${ARCH}"
info "下载地址: ${URL}"

# 创建安装目录
mkdir -p "$INSTALL_DIR"

# 下载
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl -sSLf "$URL" -o "$TMP" || error "下载失败，请确认版本 ${VERSION} 已发布且包含资产 ${ASSET}"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP" "$URL" || error "下载失败，请确认版本 ${VERSION} 已发布且包含资产 ${ASSET}"
else
  error "需要 curl 或 wget 才能下载"
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
