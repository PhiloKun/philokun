# `note` 终端速记本

随手记录闪念笔记，数据存在本地 `~/.philokun/notes.json`，零外部依赖。支持添加、列出与本地全文检索。

## 子命令

### `philokun note add <内容...>`

记录一条笔记，多个参数自动拼接成一句，并自动保存时间。

```bash
philokun note add 想到一个绝妙的域名 idea
# 已记录笔记: 想到一个绝妙的域名 idea
```

### `philokun note list`

列出全部笔记（含 ID 与创建时间）。

```bash
philokun note list
# 1. [2026-08-11T10:00:00+08:00] 想到一个绝妙的域名 idea
```

### `philokun note search <关键词...>`

按关键词做**不区分大小写**的本地全文检索。

```bash
philokun note search 域名
# 找到 1 条匹配 "域名" 的笔记：
# 1. [2026-08-11T10:00:00+08:00] 想到一个绝妙的域名 idea
```

## 数据存储

```
~/.philokun/notes.json
```

```json
{
  "seq": 1,
  "notes": [
    { "id": 1, "text": "想到一个绝妙的域名 idea", "at": "2026-08-11T10:00:00+08:00" }
  ]
}
```

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/note.go` | `add` / `list` / `search` 子命令 |
| 存储 | `internal/store/note.go` | `AddNote` / `ListNotes` / `SearchNotes`（检索用 `strings.Contains`） |

`search` 直接对内存中的笔记做 `strings.Contains` 匹配，无需任何数据库或外部索引，轻量且即时。
