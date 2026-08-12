# timer — 倒计时提醒

倒计时结束后提醒你。结束时会触发终端响铃（`\a`）并尽量弹出系统通知（macOS 用 `osascript`、Linux 用 `notify-send`）。可与番茄钟 `pomodoro` 互补：番茄钟专注工作，timer 用于任意短提醒。

## 用法

```bash
philokun timer 25m            # 25 分钟
philokun timer 1h30m          # 1 小时 30 分
philokun timer 90             # 90 秒（纯数字按秒）
philokun timer 10m 泡茶        # 带提醒文字
```

## 时长格式

- 纯数字：按秒（如 `90` = 90 秒）
- 组合单位：支持 `h`/`m`/`s` 任意组合，如 `1h30m`、`25m`、`90s`、`2h`

## 示例

```bash
$ philokun timer 2s 测试
⏳ 倒计时 2s，结束于 19:37:52

🔔 测试
```

## 说明

- 倒计时按秒刷新显示剩余时间（终端实时更新从开始时刻计算，避免漂移）。
- 提醒文字会同时出现在终端和系统通知标题中。
- 系统通知为尽力而为：macOS/Linux 桌面环境有对应命令时弹出，否则仅终端响铃 + 文字，不影响功能。
- 中途用 `Ctrl+C` 可取消。
