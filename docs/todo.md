# `todo` 待办命令

管理你的个人待办事项，支持添加、列出、标记完成与删除。数据以 JSON 形式存储在本地（`~/.philokun/todo.json`），不依赖任何云服务。

## 子命令

### `philokun todo add <内容...>`

追加一条待办，支持带空格的多词内容（多个参数自动拼接）。

```bash
philokun todo add 完成 Git 教程仓库
philokun todo add 买牛奶 写代码      # 合并为「买牛奶 写代码」一条
```

### `philokun todo list`

列出全部待办，按 ID 编号并显示完成状态（`[x]` 表示已完成）。

```bash
philokun todo list
# 1. [x] 完成 Git 教程仓库
# 2. [ ] 给 philokun 加完成状态
```

### `philokun todo done <ID>`

将指定 ID 的待办标记为已完成。

```bash
philokun todo done 1
# 已将 ID 1 标记为完成。
```

### `philokun todo rm <ID>`

删除指定 ID 的待办。

```bash
philokun todo rm 2
# 已删除 ID 2。
```

## 数据存储

```
~/.philokun/todo.json
```

```json
{
  "seq": 2,
  "todos": [
    { "id": 1, "text": "完成 Git 教程仓库", "done": true },
    { "id": 2, "text": "给 philokun 加完成状态", "done": false }
  ]
}
```

- `seq` 为自增 ID 计数器，保证每条待办 ID 唯一且稳定。
- 文件不存在时自动创建；读取失败（如文件损坏）会返回错误。

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/todo.go` | `add` / `list` / `done` / `rm` 子命令与参数校验 |
| 存储 | `internal/store/todo.go` | `Todo` 结构体与 `AddTodo` / `ListTodos` / `DoneTodo` / `RmTodo` |

`done` / `rm` 通过 ID 定位（`findTodo`），ID 不存在时返回明确错误。
