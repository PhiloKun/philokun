# `pomodoro` 番茄钟

启动一个终端倒计时番茄钟，结束时响铃提醒。可用更短的别名 `pomo`。

## 用法

```bash
philokun pomodoro 25      # 或 philokun pomo 25
# 🍅 番茄钟开始，专注 25 分钟，结束时间 14:30:00
# 剩余: 24:59 ...
```

倒计时会每秒刷新一行显示剩余时间（MM:SS），归零后播放提示音并发出系统通知，然后提示休息。

```bash
⏰ 时间到！休息一下吧。
```

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/pomodoro.go` | 用 `time.NewTicker` 每秒刷新剩余时间，归零调用 `alertBell` |

- 参数必须是正整数（分钟数），否则报错。
- 结束提醒 `alertBell` 按平台播放实际提示音并弹出系统通知：
  - macOS：`afplay` 播放系统音 + `osascript` 弹通知。
  - Linux：`paplay`/`aplay` 播放 + `notify-send` 弹通知。
  - Windows：PowerShell 播放系统音 + Toast 通知。
  - 以上都不可用时，回退到终端 BEL（`\a`）。
- 按 `Ctrl+C` 可随时中断。
