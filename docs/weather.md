# `weather` 天气查询

查询指定城市的当天天气，使用 [Open-Meteo](https://open-meteo.com/) 免费、免 API Key 的接口。联网请求通过标准库 `net/http` + `encoding/json` 实现。

## 用法

```bash
philokun weather 北京
# 城市: 北京
# 温度: 26.3°C
# 风速: 12.5 km/h
# 天气代码: 1 (多云)
```

城市名支持中文或英文：

```bash
philokun weather Shanghai
philokun weather Tokyo
```

## 实现说明

查询分两步，均返回 JSON：

1. **地理编码**：`https://geocoding-api.open-meteo.com/v1/search?name=<城市>&count=1&language=zh`
   把城市名解析为经纬度。
2. **当前天气**：`https://api.open-meteo.com/v1/forecast?latitude=..&longitude=..&current_weather=true`
   取 `temperature` / `windspeed` / `weathercode`。

`weathercode` 按 WMO 标准映射为中文描述（晴 / 多云 / 雨 / 雪 / 雷阵雨 等），见 `cmd/weather.go` 的 `describeWeather`。

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/weather.go` | 解析城市、两次 HTTP 请求、格式化输出 |

网络请求设置了 10 秒超时，接口异常或非 200 状态会返回明确错误。
