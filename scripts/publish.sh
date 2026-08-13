#!/usr/bin/env bash
# philokun 一键发布脚本（自动化发版全流程）。
#
# 流程：升版本 -> 交叉编译 -> 提交并同步文档 -> 推两个远端 -> 发 GitHub 与 Gitee Release。
#
# 用法:
#   ./scripts/publish.sh 1.2.0          # 发布指定版本
#   ./scripts/publish.sh                # 使用 cmd/version.go 里的版本号
#   VERSION=1.2.0 ./scripts/publish.sh  # 也可用环境变量指定
#
# 环境变量:
#   GITEE_TOKEN   必填，用于创建 Gitee Release 并上传附件（API v5）
#   GITHUB_TOKEN  可选，gh 已登录时不必设置；未登录时用于 gh auth login 走 token
#   SKIP_BUILD    设为 1 可跳过编译（仅发布已存在的 dist/）
#   NO_PUSH       设为 1 不推送远端、不发 Release（只本地构建 + git commit）
#
# 依赖: git, gh (GitHub CLI), curl, go, sha256sum/shasum
set -euo pipefail

cd "$(dirname "$0")/.."   # 切到项目根目录

# ---------- 颜色 ----------
info()  { printf '\033[32m[INFO]\033[0m %s\n' "$1"; }
warn()  { printf '\033[33m[WARN]\033[0m %s\n' "$1"; }
error() { printf '\033[31m[ERROR]\033[0m %s\n' "$1" >&2; exit 1; }

# ---------- 解析版本号 ----------
VERSION="${1:-${VERSION:-}}"
if [ -z "$VERSION" ]; then
  VERSION="$(grep -oE 'var version = "[^"]+"' cmd/version.go | head -1 | sed -E 's/var version = "([^"]+)"/\1/')"
fi
[ -z "$VERSION" ] && error "无法确定版本号，请传入参数或设置 VERSION 环境变量"
TAG="v${VERSION}"

# ---------- 升版本（若与 version.go 不一致） ----------
CUR="$(grep -oE 'var version = "[^"]+"' cmd/version.go | head -1 | sed -E 's/var version = "([^"]+)"/\1/')"
if [ "$CUR" != "$VERSION" ]; then
  info "更新 version.go: ${CUR} -> ${VERSION}"
  sed -i.bak -E "s/var version = \"[^\"]+\"/var version = \"${VERSION}\"/" cmd/version.go
  rm -f cmd/version.go.bak
  # 同步 README 中“当前为 X.Y.Z 版本”的引用
  if grep -qE "当前为 \`[0-9.]+\` 版本" README.md; then
    sed -i.bak -E "s/当前为 \`[0-9.]+\` 版本/当前为 \`${VERSION}\` 版本/" README.md
    rm -f README.md.bak
    info "同步 README.md 版本引用 -> ${VERSION}"
  fi
else
  info "版本号已为 ${VERSION}，无需修改 version.go"
fi

# ---------- 编译 ----------
if [ "${SKIP_BUILD:-0}" != "1" ]; then
  info "交叉编译版本: ${VERSION}"
  ./scripts/release.sh "$VERSION"
else
  warn "SKIP_BUILD=1，跳过编译，假设 dist/ 已就绪"
  [ -d dist ] || error "dist/ 不存在，无法发布"
fi

# ---------- Git 提交 + 推两远端 ----------
if [ "${NO_PUSH:-0}" != "1" ]; then
  info "提交文档/版本变更并推送"
  git add -A
  if git diff --cached --quiet; then
    info "无改动需要提交"
  else
    git commit -m "release: v${VERSION}"
  fi
  git push gitee main
  git push origin main
  # 打 tag 并推送到两个远端（GitHub 的 gh 会自动建 tag，但 gitee 需要显式推送）
  if git rev-parse "$TAG" >/dev/null 2>&1; then
    warn "tag ${TAG} 已存在，跳过打 tag"
  else
    git tag "$TAG"
  fi
  git push gitee "refs/tags/$TAG"
  git push origin "refs/tags/$TAG"
  info "已推送到 gitee / github（含 tag ${TAG}）"
else
  warn "NO_PUSH=1，跳过 git 提交与推送"
fi

# ---------- 发 GitHub Release ----------
if [ "${NO_PUSH:-0}" != "1" ]; then
  info "发布 GitHub Release ${TAG}"
  if command -v gh >/dev/null 2>&1; then
    if gh release view "$TAG" >/dev/null 2>&1; then
      warn "GitHub Release ${TAG} 已存在，补充上传缺失附件"
    else
      gh release create "$TAG" --title "philokun ${TAG}" --notes "Release ${TAG}" || true
    fi
    # 上传所有 dist 资产（已存在则覆盖）
    for f in dist/philokun-* dist/checksums-*.sha256; do
      [ -e "$f" ] || continue
      gh release upload "$TAG" "$f" --clobber 2>/dev/null || gh release upload "$TAG" "$f" || true
    done
    info "GitHub Release ${TAG} 完成"
  else
    warn "未安装 gh，跳过 GitHub Release（请手动在网页发布）"
  fi

  # ---------- 发 Gitee Release ----------
  info "发布 Gitee Release ${TAG}"
  GITEE_TOKEN="${GITEE_TOKEN:-}"
  [ -z "$GITEE_TOKEN" ] && error "缺少 GITEE_TOKEN 环境变量，无法发布 Gitee Release"
  GITEE_API="https://gitee.com/api/v5/repos/PhiloKun/philokun/releases"
  GITEE_REPO="PhiloKun/philokun"

  # 1) 按 tag 精确查询是否已有 Release（必须匹配 tag_name，避免误命中其他版本）
  REL_ID="$(curl -sS "${GITEE_API}?access_token=${GITEE_TOKEN}" \
    | python3 -c "import sys,json
try:
    d=json.load(sys.stdin)
except Exception:
    d=[]
if isinstance(d,list):
    for r in d:
        if r.get('tag_name')=='${TAG}':
            print(r.get('id')); break" 2>/dev/null)"

  if [ -z "$REL_ID" ]; then
    # 不存在则创建（必须带 target_commitish，否则 Gitee 返回 400）
    REL_ID="$(curl -sS -X POST "${GITEE_API}" \
      -H "Content-Type: application/json" \
      -d "{\"access_token\":\"${GITEE_TOKEN}\",\"tag_name\":\"${TAG}\",\"name\":\"philokun ${TAG}\",\"body\":\"Release ${TAG}\",\"target_commitish\":\"main\",\"prerelease\":false}" \
      | python3 -c "import sys,json
try:
    r=json.load(sys.stdin); print(r.get('id') or '')
except Exception: print('')" 2>/dev/null)"
  fi
  [ -z "$REL_ID" ] && error "无法创建/获取 Gitee Release（REL_ID 为空）"
  info "Gitee Release id=${REL_ID} (tag=${TAG})"

  # 2) 上传附件（先查已存在，避免重复）
  for f in dist/philokun-* dist/checksums-*.sha256; do
    [ -e "$f" ] || continue
    NAME="$(basename "$f")"
    EXIST="$(curl -sS "${GITEE_API}/${REL_ID}/attach_files?access_token=${GITEE_TOKEN}" \
      | python3 -c "import sys,json
try:
    d=json.load(sys.stdin)
except Exception:
    d=[]
if isinstance(d,list):
    for a in d:
        if a.get('name')=='${NAME}':
            print('exists'); break" 2>/dev/null)"
    if [ "$EXIST" = "exists" ]; then
      warn "Gitee 附件已存在，跳过: ${NAME}"
      continue
    fi
    curl -sSfL -X POST "${GITEE_API}/${REL_ID}/attach_files?access_token=${GITEE_TOKEN}" \
      -F "file=@${f}" >/dev/null
    info "已上传 Gitee 附件: ${NAME}"
  done
  info "Gitee Release ${TAG} 完成"
else
  warn "NO_PUSH=1，跳过 GitHub / Gitee Release 发布"
fi

echo ""
info "🎉 发布完成: ${TAG}"
info "用户安装/升级:"
echo "  curl -sSfL https://gitee.com/PhiloKun/philokun/raw/main/install.sh | RELEASE_MIRROR=gitee sh -s -- ${TAG}"
