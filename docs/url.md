# `url` 链接管理

把常用链接以「别名」形式持久化在本地，方便在终端里集中管理与一键打开。数据存在 `~/.philokun/urls.json`，零外部依赖。

## 子命令

### `philokun url add <别名> <链接>`

新增（或覆盖）一条别名链接。

```bash
philokun url add gh https://github.com
# 已保存 "gh" -> https://github.com
```

### `philokun url list`

列出全部链接书签（带数字 ID 标识，便于辨识）。

```bash
philokun url list
# 1. gh -> https://github.com
# 2. doc -> https://pkg.go.dev
```

### `philokun url open <别名>`

在默认浏览器中打开别名对应的链接。macOS 使用 `open`，Windows 使用 `cmd /c start`，其余平台使用 `xdg-open`。

```bash
philokun url open gh
# 正在打开: https://github.com
```

## 数据存储

```
~/.philokun/urls.json
```

```json
{
  "seq": 2,
  "urls": [
    { "id": 1, "alias": "gh", "link": "https://github.com" },
    { "id": 2, "alias": "doc", "link": "https://pkg.go.dev" }
  ]
}
```

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/url.go` | `add` / `list` / `open` 子命令，跨平台打开命令 |
| 存储 | `internal/store/url.go` | `AddURL`（别名唯一，重复则覆盖，自增 ID）/ `ListURLs` / `GetURL` |

别名作为唯一键，重复 `add` 同一别名会更新其链接而非新增。`id` 为自增数字标识，旧数据（无 `id`）在首次列出时自动补充分配并写回。
