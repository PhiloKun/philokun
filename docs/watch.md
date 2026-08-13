# watch — 周期性重复执行命令 / 请求 URL

每隔固定间隔重复执行一个 shell 命令，或请求一个 URL 并打印状态，类似 `watch` 命令。

## 用法

| 命令 | 作用 |
|------|------|
| `philokun watch <命令>` | 每 2 秒执行一次 shell 命令 |
| `philokun watch -n <秒> -c <次数> <命令>` | 自定义间隔与最大次数（`-c 0` 无限） |
| `philokun watch <URL>` | 以 http(s):// 开头时按 URL 处理，打印状态码与响应大小 |

## 示例

```bash
# 每秒刷新时间
philokun watch -n 1 "date"

# 最多执行 10 次，每次间隔 5 秒
philokun watch -n 5 -c 10 "ls -l"

# 监控一个健康接口
philokun watch -n 3 https://example.com/health
# HTTP 200  OK
# 响应大小: 12 字节
```

## 实现说明

- 命令模式：整段参数作为 `sh -c` 执行，stdout/stderr 直接透传。
- URL 模式：用 `net/http` GET 请求，输出状态码、状态行与响应字节数。
- 每次刷新前清屏（ANSI `\033[2J\033[H`），仅保留当前这一次输出。
- 间隔默认 2 秒；`-c` 指定次数（0 表示无限，需手动 `Ctrl-C` 退出）。
