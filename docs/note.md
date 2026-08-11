# `note` 终端速记本

随手记录闪念笔记，数据存在本地 `~/.philokun/notes.json`，零外部依赖。支持添加、列出、检索、修改、删除与撤销删除。

> 时间显示格式为 `YYYY-MM-DD HH:mm:ss`（存储内部仍是 RFC3339，兼容旧数据）。

## 子命令

### `philokun note add <内容...>`

记录一条笔记，多个参数自动拼接成一句，并自动保存时间。

```bash
philokun note add 想到一个绝妙的域名 idea
# 已记录笔记: 想到一个绝妙的域名 idea
```

### `philokun note list`

列出全部笔记（含 ID 与创建时间），被删除的笔记不会显示。

```bash
philokun note list
# 1. [2026-08-11 10:00:00] 想到一个绝妙的域名 idea
```

### `philokun note search <关键词...>`

按关键词做**不区分大小写**的本地全文检索。

```bash
philokun note search 域名
# 找到 1 条匹配 "域名" 的笔记：
# 1. [2026-08-11 10:00:00] 想到一个绝妙的域名 idea
```

### `philokun note edit <id>`

修改一条笔记的内容。笔记为单一文本（标题与正文统一为整条文本）。

- 交互模式：先打印当前内容，再提示输入新内容，输入后询问是否保存（`[y/N]`），`n` 或回车取消。
- 非交互模式：用 `-m/--message` 直接指定新内容并保存（配合 `-y/--yes` 跳过确认）。

```bash
philokun note edit 1
# 当前内容: 旧内容
# 请输入新内容（直接回车取消）: 新内容
# 保存对笔记 #1 的修改? [y/N]: y
# 已更新笔记 #1。

philokun note edit 1 -m "用命令直接改好的内容"
# 已更新笔记 #1。
```

### `philokun note rm <id>...`

删除一条或多条笔记（批量用空格分隔多个 ID）。删除为**软删除**，操作后给出确认提示 `[y/N]`，可用 `note undo` 恢复；加 `-y/--yes` 跳过确认（脚本友好）。别名：`del` / `delete` / `remove`。

```bash
philokun note rm 1
# 确认删除笔记 #1 "旧内容"? [y/N]: y
# 已删除 1 条笔记。

philokun note rm 1 3 --yes
# 已删除 2 条笔记。
```

> 删除不存在的 ID 会直接报错，不会删除任何笔记。
> 删除/修改成功后都会打印更新后的笔记列表，方便即时核对。
> 删除后腾出的编号会被回收，新建笔记会从当前最小的可用编号开始（通常从 1 起），而不会继续累加旧编号。

### `philokun note undo`

撤销**最近一次**删除操作，恢复被软删的笔记（整批一起恢复）。

```bash
philokun note undo
# 已撤销删除，恢复 2 条笔记。
```

> 只有最近一次删除批次可被撤销；再次删除会覆盖上一批次记录。

### `philokun note clear`

清空**全部**笔记（软删除，可用 `note undo` 恢复）。操作前有确认提示 `[y/N]`，加 `-y/--yes` 跳过确认（脚本友好）。别名：`empty` / `clean`。

```bash
philokun note clear
# 确认清空全部 3 条笔记？此操作可用 `philokun note undo` 撤销 [y/N]: y
# 已清空 3 条笔记。

philokun note clear --yes
# 已清空 3 条笔记。
```

> 清空为空时直接提示无需操作；清空后列表显示为空状态，可用 `note undo` 一键恢复。

### `philokun note purge <id>...`

**物理删除**指定的笔记（单条或批量，空格分隔多个 ID）：直接从存储中**真正移除记录**，不可撤销。操作前有醒目的【物理删除】确认提示 `[y/N]`，加 `-y/--yes` 跳过确认。别名：`erase` / `wipe` / `destroy`。

```bash
philokun note purge 2
# 确认【物理删除】笔记 #2 "旧内容"？（此操作不可撤销） [y/N]: y
# 已物理删除 1 条笔记（不可撤销）。

philokun note purge 1 3 --yes
# 已物理删除 2 条笔记（不可撤销）。
```

> 与 `rm`（软删除，可 `undo`）不同，`purge` 会永久丢弃数据，请谨慎使用。

### `philokun note purge-all`

**物理清空**全部笔记：真正移除所有记录，不可撤销。操作前有【物理清空】确认提示 `[y/N]`，加 `-y/--yes` 跳过。别名：`erase-all` / `wipe-all`。

```bash
philokun note purge-all
# 确认【物理清空】全部 3 条笔记？此操作不可撤销！ [y/N]: y
# 已物理清空 3 条笔记（不可撤销）。
```

> 与 `clear`（软删除，可 `undo`）不同，`purge-all` 不可恢复；删除后文件中的 `notes` 直接清空。

## 软删除 vs 物理删除

| 命令 | 行为 | 可撤销 | 存储表现 |
|------|------|--------|----------|
| `note rm` / `note clear` | 软删除（置 `deleted` 标记） | 是（`note undo`） | 记录保留但被过滤 |
| `note purge` / `note purge-all` | 物理删除（真正移除） | 否 | 记录从文件移除 |

数据通过 `~/.philokun/notes.json` 持久化；所有写入均采用**临时文件 + 原子 rename**，保证写操作要么完整生效、要么原文件不变（等价事务性，避免半写损坏）。

## 数据存储

```
~/.philokun/notes.json
```

```json
{
  "seq": 1,
  "notes": [
    {
      "id": 1,
      "text": "想到一个绝妙的域名 idea",
      "at": "2026-08-11T10:00:00+08:00",
      "deleted": false,
      "deleted_at": ""
    }
  ],
  "last_deleted": [2, 3]
}
```

- `deleted` / `deleted_at`：软删除标记（旧数据无此字段时默认未删除，完全兼容）。
- `last_deleted`：最近一次删除批次的 ID 列表，供 `undo` 恢复；若被恢复 ID 已被新笔记占用，会自动分配新编号。
- `seq` 字段已废弃，新增笔记时不再使用；系统总是分配当前未删除笔记中最小的可用正整数 ID，确保编号紧凑连续。

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/note.go` | `add` / `list` / `search` / `edit` / `rm` / `undo` 子命令 |
| 存储 | `internal/store/note.go` | `AddNote` / `ListNotes` / `SearchNotes` / `GetNote` / `UpdateNote` / `DeleteNote` / `DeleteNotes` / `ClearNotes` / `PurgeNote` / `PurgeNotes` / `PurgeAllNotes` / `UndoDelete` |

- 删除采用**软删除**：仅置 `deleted=true`，`ListNotes`/`SearchNotes`/`GetNote` 自动过滤，避免误删丢失。
- `undo` 通过 `last_deleted` 记录最近删除批次，整批恢复。
- `search` 直接对内存中的笔记做 `strings.Contains` 匹配，无需任何数据库或外部索引，轻量且即时。
- 交互确认与读取共用单一 stdin 缓冲读取器，避免数据错乱。

