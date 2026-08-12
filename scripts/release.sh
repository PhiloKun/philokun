#!/usr/bin/env bash
# 交叉编译脚本：生成各平台二进制，便于上传到 GitHub/Gitee Release。
# 用法:
#   ./scripts/release.sh            # 使用 cmd/version.go 里的版本号
#   VERSION=1.0.0 ./scripts/release.sh
#   ./scripts/release.sh v1.2.3     # 命令行指定版本（覆盖 VERSION 环境变量）
#
# 产物输出到 dist/ 目录，文件名形如 philokun-darwin-arm64 等。

set -euo pipefail

cd "$(dirname "$0")/.."   # 切到项目根目录

# 版本号优先级: 命令行参数 > 环境变量 VERSION > 从 version.go 提取
VERSION="${1:-${VERSION:-}}"
if [ -z "$VERSION" ]; then
  VERSION="$(grep -oE 'var version = "[^"]+"' cmd/version.go | head -1 | sed -E 's/var version = "([^"]+)"/\1/')"
fi
if [ -z "$VERSION" ]; then
  echo "无法确定版本号，请传入参数或设置 VERSION 环境变量" >&2
  exit 1
fi

BINARY="philokun"
DIST="dist"
LDFLAGS="-X github.com/philokun/cmd.version=${VERSION}"

# 支持的平台: OS/ARCH
PLATFORMS=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

mkdir -p "$DIST"
echo "构建版本: ${VERSION}"
echo "输出目录: ${DIST}/"

for platform in "${PLATFORMS[@]}"; do
  OS="${platform%/*}"
  ARCH="${platform#*/}"
  OUT="${DIST}/${BINARY}-${OS}-${ARCH}"
  if [ "$OS" = "windows" ]; then OUT="${OUT}.exe"; fi

  echo "  -> ${OS}/${ARCH}"
  GOOS="$OS" GOARCH="$ARCH" CGO_ENABLED=0 \
    go build -ldflags "$LDFLAGS" -o "$OUT" .
done

# 生成校验和
cd "$DIST"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "${BINARY}"-* > "checksums-${VERSION}.sha256"
elif command -v shasum >/dev/null 2>&1; then
  shasum -a 256 "${BINARY}"-* > "checksums-${VERSION}.sha256"
fi
echo "已生成校验和: ${DIST}/checksums-${VERSION}.sha256"

echo ""
echo "下一步：把这些文件上传到 GitHub/Gitee Release (tag: v${VERSION})"
echo "  - ${DIST}/${BINARY}-*"
echo "  - ${DIST}/checksums-${VERSION}.sha256"
