# tree — 目录树展示

以树状结构递归打印目录内容，默认从当前目录开始。零云依赖。

## 用法

| 命令 | 作用 |
|------|------|
| `philokun tree [目录]` | 打印目录树（默认当前目录） |
| `philokun tree -L <深度>` | 限制最大递归深度（`0` 表示不限制） |
| `philokun tree -a` | 包含隐藏文件 / 目录 |
| `philokun tree -d` | 只看目录 |

## 示例

```bash
philokun tree
philokun tree ./cmd
philokun tree -L 2
philokun tree -a -d /etc
```

## 实现说明

- 使用 `os.ReadDir` 递归遍历，输出 `├──` / `└──` 形式的树。
- `-a` 默认跳过以 `.` 开头的隐藏项；`-d` 仅展示目录节点。
- 目录节点末尾带 `/` 以区分文件。
