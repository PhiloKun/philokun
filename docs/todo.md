# `todo` 待办命令

管理你的个人待办事项。数据全部以 JSON 形式存储在本地（`~/.philokun/todo.json`），不依赖任何云服务。

## 子命令

### `philokun todo add <内容...>`

追加一条待办。支持带空格的多词内容，多个参数会被自动拼接成一句。

```bash
philokun todo add 完成 Git 教程仓库
philokun todo add 买牛奶 写代码      # 会合并为「买牛奶 写代码」一条
```

输出：

```
已添加待办: 完成 Git 教程仓库
```

### `philokun todo list`

列出当前全部待办，按添加顺序编号。无待办时给出友好提示。

```bash
philokun todo list
```

输出：

```
1. 完成 Git 教程仓库
2. 给 philokun 加完成状态
```

## 数据存储

待办保存在用户主目录下的 JSON 文件：

```
~/.philokun/todo.json
```

结构示例：

```json
{
  "todos": [
    "完成 Git 教程仓库",
    "给 philokun 加完成状态"
  ]
}
```

- 文件不存在时自动创建；读取失败（如文件损坏）会返回错误。
- 路径位于用户主目录，不污染项目目录，命令重装数据不丢失。

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/todo.go` | 定义 `todo` / `add` / `list` 子命令与参数校验 |
| 存储 | `internal/store/todo.go` | `AddTodo` / `ListTodos` 读写本地 JSON |

## 测试

`internal/store/todo_test.go` 与 `cmd/todo_test.go` 覆盖：

- 添加后列表数量与顺序正确；
- 多参数空格拼接；
- 空列表返回提示文案；
- 无参数时 `add` 由 Cobra 拦截报错；
- 数据文件正确生成在 `~/.philokun/` 下。

运行：

```bash
go test ./...
```
