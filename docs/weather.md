# `weather` 天气查询

查询城市天气，使用 [Open-Meteo](https://open-meteo.com/) 免费、免 API Key 的接口（天气 + 空气质量）。联网请求通过标准库 `net/http` + `encoding/json` 实现。

支持两种形态：

- `weather <城市...>`：终端文本输出，含当前天气、小时趋势、7 天预报、极端天气预警。
- `weather web <城市...>`：启动本地毛玻璃卡片网页，支持多城横滑对比、动态渐变背景、极端天气醒目提示。

## 用法

### 终端文本

```bash
philokun weather 北京
# 城市: 北京 🌧️
# 温度: 30.7°C   毛毛雨
# 湿度: 62%
# 风力: 12.3 km/h
# 空气质量: 82 (重度污染)
# ⚠️ 预警: 暴雨/雷暴预警 · 注意防雷电与积水
#
# 📈 小时预报（温度趋势）:
# ▁▂▃▅▇█...
# 17:00 🌧️ 32°C  21:00 ☁️ 28°C ...
#
# 🗓️ 未来 7 天:
#   [周二] 🌧️ 毛毛雨  08-11 33°C/26°C  风14 降水92%
#   ...
```

多城市直接空格分隔，依次展示并对比：

```bash
philokun weather 北京 上海 Tokyo
```

### 网页 UI

```bash
philokun weather web 北京 上海 -p 8090
# 天气网页已启动: http://127.0.0.1:8090  (Ctrl+C 退出)
```

打开浏览器后可见：

- 顶部城市横滑栏，选中城市高亮（`active`）。
- 毛玻璃卡片（`backdrop-filter: blur`），展示城市名、大字号温度、天气图标与文字。
- 当前：温度 / 湿度 / 风力 / 空气质量（AQI + 等级）/ 更新时间。
- 小时预报：ASCII 温度趋势曲线 + 关键时段（图标/温度）。
- 7 天预报：卡片列表，每日天气摘要、最高/最低温、风力、降水概率。
- 极端天气（暴雨、高温 ≥35°C、大风 ≥40km/h 等）以红色脉冲卡片醒目提示，并显示预警等级文案。
- 页面背景根据主导天气动态渐变：晴天暖色、阴天冷色、雨天暗调、雷暴深紫、雪天冷蓝。

## 实现说明

查询分三步，均返回 JSON：

1. **地理编码**：`geocoding-api.open-meteo.com/v1/search?name=<城市>&count=1&language=zh`
2. **天气预报**：`api.open-meteo.com/v1/forecast` 拉取 `current_weather` + `hourly`（温度/天气代码）+ `daily`（7 天最高/最低/风力/降水概率），`timezone=auto`。
3. **空气质量（best-effort）**：`air-quality-api.open-meteo.com/v1/air-quality` 取欧洲 AQI。

`detectAlerts` 根据天气代码、风力、各日最高温识别极端天气；`bgClass` 决定网页背景渐变；`asciiSpark` 生成终端温度趋势字符画。

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/weather.go` | 解析城市、三次 HTTP 请求、CLI 文本输出、启动 `weather web` 本地服务与 HTML 模板 |

网络请求设置 10 秒超时，接口异常或非 200 状态会返回明确错误；空气质量缺失时不影响主流程。
