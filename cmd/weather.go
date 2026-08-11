package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ===================== 数据层 =====================

// weatherCmd 查询指定城市的天气（增强版，含当前/小时/7天/空气质量）。
var weatherCmd = &cobra.Command{
	Use:   "weather <城市...>",
	Short: "查询城市天气（当前/小时/7天/空气质量，免 key）",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for i, city := range args {
			if i > 0 {
				fmt.Println(strings.Repeat("─", 40))
			}
			if err := printCityWeather(city); err != nil {
				fmt.Printf("⚠️ %v\n", err)
			}
		}
		return nil
	},
}

// weatherWebCmd 启动本地网页，以毛玻璃卡片 UI 展示天气。
var weatherWebCmd = &cobra.Command{
	Use:   "web <城市...>",
	Short: "启动天气网页（毛玻璃卡片 UI，支持多城对比）",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetString("port")
		if port == "" {
			port = "8080"
		}
		return serveWeatherWeb(args, port)
	},
}

// geoResult 是解析后的地理编码结果。
type geoResult struct {
	Name      string
	Country   string
	Latitude  float64
	Longitude float64
}

// weatherData 聚合一次查询的全部天气数据。
type weatherData struct {
	City      string
	Country   string
	Current   currentBlock
	Hourly    []hourBlock
	Daily     []dayBlock
	Air       int // 空气质量指数（0 表示无数据）
	AirDesc   string
	Alerts    []string // 极端天气预警文案
}

type currentBlock struct {
	Temp      float64
	Humidity  float64
	Wind      float64
	Code      int
	Time      string
}

type hourBlock struct {
	Time string
	Temp float64
	Code int
}

type dayBlock struct {
	Date       string
	Code       int
	TempMax    float64
	TempMin    float64
	WindMax    float64
	PrecipProb float64
}

// ---- Open-Meteo 响应结构 ----

type geocodingResp struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

type forecastResp struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Humidity    float64 `json:"relativehumidity_2m"`
		Windspeed   float64 `json:"windspeed"`
		Weathercode int     `json:"weathercode"`
		Time        string  `json:"time"`
	} `json:"current_weather"`
	Hourly struct {
		Time          []string  `json:"time"`
		Temperature2m []float64 `json:"temperature_2m"`
		Humidity2m    []float64 `json:"relativehumidity_2m"`
		Weathercode   []int     `json:"weathercode"`
	} `json:"hourly"`
	Daily struct {
		Time        []string  `json:"time"`
		Weathercode []int     `json:"weathercode"`
		TempMax     []float64 `json:"temperature_2m_max"`
		TempMin     []float64 `json:"temperature_2m_min"`
		WindMax     []float64 `json:"windspeed_10m_max"`
		PrecipProb  []float64 `json:"precipitation_probability_max"`
	} `json:"daily"`
}

type airResp struct {
	Hourly struct {
		AQI []int `json:"european_aqi"`
	} `json:"hourly"`
}

const (
	geocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL  = "https://api.open-meteo.com/v1/forecast"
	airURL       = "https://air-quality-api.open-meteo.com/v1/air-quality"
)

func httpGetJSON(endpoint string, out any) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("接口返回状态码 %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func geocode(city string) (geoResult, error) {
	u := fmt.Sprintf("%s?name=%s&count=1&language=zh", geocodingURL, url.QueryEscape(city))
	var gr geocodingResp
	if err := httpGetJSON(u, &gr); err != nil {
		return geoResult{}, err
	}
	if len(gr.Results) == 0 {
		return geoResult{}, fmt.Errorf("找不到城市 %q", city)
	}
	r := gr.Results[0]
	return geoResult{Name: r.Name, Country: r.Country, Latitude: r.Latitude, Longitude: r.Longitude}, nil
}

// fetchWeather 拉取某城市的当前/小时/7天天气，并尝试获取空气质量。
func fetchWeather(city string) (weatherData, error) {
	geo, err := geocode(city)
	if err != nil {
		return weatherData{}, err
	}
	u := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f"+
		"&current_weather=true"+
		"&hourly=temperature_2m,relativehumidity_2m,weathercode"+
		"&daily=weathercode,temperature_2m_max,temperature_2m_min,windspeed_10m_max,precipitation_probability_max"+
		"&timezone=auto",
		forecastURL, geo.Latitude, geo.Longitude)
	var fr forecastResp
	if err := httpGetJSON(u, &fr); err != nil {
		return weatherData{}, err
	}

	wd := weatherData{
		City:    geo.Name,
		Country: geo.Country,
		Current: currentBlock{
			Temp:      fr.CurrentWeather.Temperature,
			Humidity:  fr.CurrentWeather.Humidity,
			Wind:      fr.CurrentWeather.Windspeed,
			Code:      fr.CurrentWeather.Weathercode,
			Time:      fr.CurrentWeather.Time,
		},
	}

	// 当前湿度：current_weather 的 humidity 字段有时为空，
	// 优先取最近整点的 hourly 湿度，没有则保持原值。
	if nowIdx := nearestHourIndex(fr.Hourly.Time); nowIdx >= 0 && nowIdx < len(fr.Hourly.Humidity2m) {
		if h := fr.Hourly.Humidity2m[nowIdx]; h > 0 {
			wd.Current.Humidity = h
		}
	}

	// 小时预报：取今天未来的时段（最多 24 条）。
	nowHour := time.Now().Format("2006-01-02T15")
	for i, t := range fr.Hourly.Time {
		if strings.Compare(t[:13], nowHour) < 0 {
			continue
		}
		if i >= len(fr.Hourly.Temperature2m) || i >= len(fr.Hourly.Weathercode) {
			break
		}
		wd.Hourly = append(wd.Hourly, hourBlock{
			Time: t[11:16],
			Temp: fr.Hourly.Temperature2m[i],
			Code: fr.Hourly.Weathercode[i],
		})
		if len(wd.Hourly) >= 24 {
			break
		}
	}

	// 7 天预报。
	for i := range fr.Daily.Time {
		if i >= len(fr.Daily.Weathercode) || i >= len(fr.Daily.TempMax) ||
			i >= len(fr.Daily.TempMin) || i >= len(fr.Daily.WindMax) ||
			i >= len(fr.Daily.PrecipProb) {
			break
		}
		wd.Daily = append(wd.Daily, dayBlock{
			Date:       fr.Daily.Time[i],
			Code:       fr.Daily.Weathercode[i],
			TempMax:    fr.Daily.TempMax[i],
			TempMin:    fr.Daily.TempMin[i],
			WindMax:    fr.Daily.WindMax[i],
			PrecipProb: fr.Daily.PrecipProb[i],
		})
	}

	// 空气质量（best-effort）。
	if aqi, desc, err := fetchAir(geo.Latitude, geo.Longitude); err == nil {
		wd.Air = aqi
		wd.AirDesc = desc
	}

	wd.Alerts = detectAlerts(wd.Current.Code, wd.Current.Wind, wd.Daily)
	return wd, nil
}

func fetchAir(lat, lon float64) (int, string, error) {
	u := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f&hourly=european_aqi", airURL, lat, lon)
	var ar airResp
	if err := httpGetJSON(u, &ar); err != nil {
		return 0, "", err
	}
	if len(ar.Hourly.AQI) == 0 {
		return 0, "", fmt.Errorf("无空气质量数据")
	}
	// 取第一个有效值。
	idx := 0
	for i, v := range ar.Hourly.AQI {
		if v > 0 {
			idx = i
			break
		}
	}
	aqi := ar.Hourly.AQI[idx]
	return aqi, aqiLevel(aqi), nil
}

// ===================== 业务判断 =====================

// detectAlerts 根据天气代码/风力/降水概率识别极端天气，返回预警文案。
func detectAlerts(code int, wind float64, daily []dayBlock) []string {
	var alerts []string
	// 大风：风速 ≥ 40 km/h（约 6 级风以上）。
	if wind >= 40 {
		alerts = append(alerts, fmt.Sprintf("大风预警 · 当前风速 %.0f km/h，减少外出", wind))
	}
	// 暴雨/雷暴：代码 65(大雨)/82(强阵雨)/95/96/99。
	switch code {
	case 65, 82, 95, 96, 99:
		alerts = append(alerts, "暴雨/雷暴预警 · 注意防雷电与积水")
	}
	// 高温：任意一天最高温 ≥ 35°C。
	for _, d := range daily {
		if d.TempMax >= 35 {
			alerts = append(alerts, fmt.Sprintf("高温预警 · %s 最高 %.0f°C", weekday(d.Date), d.TempMax))
			break
		}
	}
	return alerts
}

func aqiLevel(aqi int) string {
	switch {
	case aqi <= 20:
		return "优"
	case aqi <= 40:
		return "良"
	case aqi <= 60:
		return "轻度污染"
	case aqi <= 80:
		return "中度污染"
	case aqi <= 100:
		return "重度污染"
	default:
		return "严重污染"
	}
}

// ===================== 展示：CLI =====================

func printCityWeather(city string) error {
	wd, err := fetchWeather(city)
	if err != nil {
		return err
	}
	// 标签已明确含义、取值范围/单位，便于用户无需查文档即可理解。
	fmt.Printf("城市 (City): %s %s\n", wd.City, flag(wd.Current.Code))
	fmt.Printf("温度 (Temperature): %.1f°C  天气: %s\n", wd.Current.Temp, describeWeather(wd.Current.Code))
	fmt.Printf("湿度 (Humidity): %.0f%%  范围: 0~100%% (相对湿度)\n", wd.Current.Humidity)
	fmt.Printf("风力 (Wind Speed): %.1f km/h  示例: 0~12 轻风, ≥40 大风\n", wd.Current.Wind)
	if wd.Air > 0 {
		fmt.Printf("空气质量 (AQI): %d (%s)  范围: 0~500, 数值越低越好\n", wd.Air, wd.AirDesc)
	} else {
		fmt.Println("空气质量 (AQI): 暂无数据")
	}
	if len(wd.Alerts) > 0 {
		for _, a := range wd.Alerts {
			fmt.Printf("⚠️ 预警 (Alert): %s\n", a)
		}
	}

	if len(wd.Hourly) > 0 {
		fmt.Printf("\n📈 小时预报 (Hourly Forecast, 未来 24h, 单位 °C):\n")
		fmt.Printf("温度趋势: %s\n", asciiSpark(wd.Hourly))
		fmt.Print("关键时段: ")
		for i, h := range wd.Hourly {
			if i%4 == 0 {
				fmt.Printf("[%s %s %.0f°C] ", h.Time, flag(h.Code), h.Temp)
			}
		}
		fmt.Println()
	}

	if len(wd.Daily) > 0 {
		fmt.Printf("\n🗓️ 未来 7 天 (Daily Forecast, °C / km/h / %%):\n")
		for _, d := range wd.Daily {
			fmt.Printf("  [%s] %s %s  %s  最高/最低: %.0f°C/%.0f°C  最大风速: %.0f km/h  降水概率: %.0f%%\n",
				weekday(d.Date), flag(d.Code), describeWeather(d.Code),
				d.Date[5:], d.TempMax, d.TempMin, d.WindMax, d.PrecipProb)
		}
	}
	return nil
}

// asciiSpark 用纯 ASCII 字符画出温度趋势曲线，避免终端字体缺失导致的乱码。
// 使用 8 级斜坡字符 ',.`-~:;=!*#$@' 表示从最低到最高温度。
func asciiSpark(hours []hourBlock) string {
	if len(hours) == 0 {
		return ""
	}
	const width = 40
	n := len(hours)
	step := max((n-1)/(width-1), 1)
	min, max := hours[0].Temp, hours[0].Temp
	for _, h := range hours {
		if h.Temp < min {
			min = h.Temp
		}
		if h.Temp > max {
			max = h.Temp
		}
	}
	span := max - min
	if span < 0.001 {
		span = 1
	}
	const ramp = ",.-~:;=!*#$@"
	var b strings.Builder
	for i := range width {
		idx := i * step
		if idx >= n {
			idx = n - 1
		}
		pos := int((hours[idx].Temp - min) / span * float64(len(ramp)-1))
		b.WriteByte(ramp[pos])
	}
	return b.String()
}

// nearestHourIndex 找到当前时间最接近的 hourly 数据下标。
func nearestHourIndex(times []string) int {
	now := time.Now()
	best, bestDiff := -1, time.Hour
	for i, t := range times {
		tm, err := time.Parse("2006-01-02T15:04", t)
		if err != nil {
			continue
		}
		d := tm.Sub(now)
		if d < 0 {
			d = -d
		}
		if d < bestDiff {
			bestDiff = d
			best = i
		}
	}
	return best
}

// ===================== 展示：Web =====================

type webCtx struct {
	Cities []weatherData
}

func serveWeatherWeb(cities []string, port string) error {
	data := make([]weatherData, 0, len(cities))
	for _, c := range cities {
		wd, err := fetchWeather(c)
		if err != nil {
			log.Printf("跳过 %s: %v", c, err)
			continue
		}
		data = append(data, wd)
	}
	if len(data) == 0 {
		return fmt.Errorf("没有可展示的城市")
	}

	tmpl := template.Must(template.New("w").Funcs(template.FuncMap{
		"desc":       describeWeather,
		"flag":       flag,
		"weekday":    weekday,
		"aqiLevel":   aqiLevel,
		"bgClass":    bgClass,
		"mod":        func(a, b int) int { return a % b },
		"spark":      func(w weatherData) string { return asciiSpark(w.Hourly) },
		"tempAt":     func(w weatherData, i int) float64 { return w.Hourly[i].Temp },
		"hourAt":     func(w weatherData, i int) string { return w.Hourly[i].Time },
		"hourCodeAt": func(w weatherData, i int) int { return w.Hourly[i].Code },
	}).Parse(weatherHTML))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, webCtx{Cities: data}); err != nil {
			log.Printf("模板渲染错误: %v", err)
		}
	})

	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://127.0.0.1:%s", port)
	fmt.Printf("天气网页已启动: %s  (Ctrl+C 退出)\n", url)
	return http.Serve(ln, nil)
}

// bgClass 根据主导天气返回背景渐变类名。
func bgClass(w weatherData) string {
	switch {
	case w.Current.Code == 0:
		return "bg-sunny"
	case w.Current.Code >= 71 && w.Current.Code <= 86:
		return "bg-snow"
	case w.Current.Code >= 51 && w.Current.Code <= 67:
		return "bg-rain"
	case w.Current.Code >= 80 && w.Current.Code <= 82:
		return "bg-rain"
	case w.Current.Code >= 95:
		return "bg-storm"
	case w.Current.Code >= 1 && w.Current.Code <= 3:
		return "bg-cloudy"
	default:
		return "bg-cold"
	}
}

// ===================== 工具函数 =====================

// flag 返回天气代码对应的 emoji 图标。
func flag(code int) string {
	switch code {
	case 0:
		return "☀️"
	case 1, 2:
		return "🌤️"
	case 3:
		return "☁️"
	case 45, 48:
		return "🌫️"
	case 51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82:
		return "🌧️"
	case 71, 73, 75, 77, 85, 86:
		return "🌨️"
	case 95, 96, 99:
		return "⛈️"
	default:
		return "🌡️"
	}
}

// describeWeather 把 WMO 天气代码翻译成中文描述。
func describeWeather(code int) string {
	switch code {
	case 0:
		return "晴"
	case 1, 2, 3:
		return "多云"
	case 45, 48:
		return "雾"
	case 51, 53, 55, 56, 57:
		return "毛毛雨"
	case 61, 63, 65, 66, 67:
		return "雨"
	case 71, 73, 75, 77:
		return "雪"
	case 80, 81, 82:
		return "阵雨"
	case 85, 86:
		return "阵雪"
	case 95:
		return "雷阵雨"
	case 96, 99:
		return "雷阵雨伴冰雹"
	default:
		return "未知"
	}
}

func weekday(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}[int(t.Weekday())]
}

// ===================== 网页模板 =====================

const weatherHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Philokun 天气</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: "PingFang SC", "Microsoft YaHei", system-ui, sans-serif;
         min-height: 100vh; color: #fff; padding: 28px; }
  .bg-sunny { background: linear-gradient(135deg,#ff9a44,#ffd194); }
  .bg-cloudy { background: linear-gradient(135deg,#5b7a99,#9bb4c9); }
  .bg-rain { background: linear-gradient(135deg,#2c3e50,#4b6584); }
  .bg-storm { background: linear-gradient(135deg,#1a1a2e,#3a1c4a); }
  .bg-snow { background: linear-gradient(135deg,#5d8aa8,#cfe8f3); }
  .bg-cold { background: linear-gradient(135deg,#3a4a5a,#6b8299); }
  h1 { font-size: 22px; margin-bottom: 14px; font-weight: 700; }
  /* 城市横滑栏 */
  .citybar { display: flex; gap: 12px; overflow-x: auto; padding: 6px 2px 14px;
             scroll-snap-type: x mandatory; }
  .citybar::-webkit-scrollbar { height: 6px; }
  .citybar::-webkit-scrollbar-thumb { background: rgba(255,255,255,.4); border-radius: 3px; }
  .chip { flex: 0 0 auto; scroll-snap-align: start; padding: 8px 16px; border-radius: 999px;
          background: rgba(255,255,255,.18); backdrop-filter: blur(8px);
          font-weight: 600; cursor: pointer; border: 1px solid rgba(255,255,255,.25); }
  .chip.active { background: rgba(255,255,255,.4); box-shadow: 0 0 0 2px #fff; }
  /* 毛玻璃卡片 */
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px,1fr)); gap: 18px; margin-top: 6px; }
  .card { background: rgba(255,255,255,.16); backdrop-filter: blur(14px);
          border: 1px solid rgba(255,255,255,.25); border-radius: 20px; padding: 18px;
          box-shadow: 0 8px 32px rgba(0,0,0,.18); }
  .city-name { font-size: 18px; font-weight: 700; }
  .temp-main { font-size: 46px; font-weight: 800; line-height: 1.1; margin: 6px 0; }
  .cond { font-size: 15px; opacity: .95; }
  .meta { display: grid; grid-template-columns: 1fr 1fr; gap: 8px 14px; margin-top: 12px; font-size: 13px; }
  .meta b { font-weight: 600; opacity: .85; }
  .alert { margin-top: 12px; padding: 10px 12px; border-radius: 12px;
           background: rgba(255,59,48,.85); font-weight: 700; font-size: 13px;
           animation: pulse 1.2s infinite; }
  @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.6} }
  .spark { margin-top: 14px; font-size: 18px; letter-spacing: 2px; white-space: nowrap; overflow-x:auto; }
  .hours { display: flex; gap: 6px; overflow-x: auto; margin-top: 8px; padding-bottom: 4px; }
  .hours div { flex: 0 0 auto; text-align: center; font-size: 12px; }
  .days { margin-top: 14px; display: flex; flex-direction: column; gap: 8px; }
  .day { display: flex; justify-content: space-between; align-items: center;
         background: rgba(255,255,255,.12); border-radius: 12px; padding: 10px 12px; font-size: 13px; }
  .day .big { font-size: 16px; font-weight: 700; }
</style>
</head>
<body class="{{ if .Cities }}{{ bgClass (index .Cities 0) }}{{ end }}">
  <h1>🌤️ Philokun 天气</h1>
  <div class="citybar">
    {{ range $i, $c := .Cities }}
    <div class="chip{{ if eq $i 0 }} active{{ end }}">{{ $c.City }} {{ flag $c.Current.Code }}</div>
    {{ end }}
  </div>
  <div class="grid">
  {{ range .Cities }}
    <div class="card">
      <div class="city-name">{{ .City }}{{ if .Country }} · {{ .Country }}{{ end }}</div>
      <div class="temp-main">{{ printf "%.1f" .Current.Temp }}°C</div>
      <div class="cond">{{ flag .Current.Code }} {{ desc .Current.Code }}</div>
      <div class="meta">
        <div><b>湿度</b><br>{{ printf "%.0f" .Current.Humidity }}%</div>
        <div><b>风力</b><br>{{ printf "%.1f" .Current.Wind }} km/h</div>
        <div><b>空气质量</b><br>{{ if gt .Air 0 }}{{ .Air }} ({{ .AirDesc }}){{ else }}—{{ end }}</div>
        <div><b>更新</b><br>{{ .Current.Time }}</div>
      </div>
      {{ if .Alerts }}
        {{ range .Alerts }}<div class="alert">⚠️ {{ . }}</div>{{ end }}
      {{ end }}
      <div class="spark">{{ spark . }}</div>
      <div class="hours">
        {{ $w := . }}{{ range $i, $h := .Hourly }}{{ if eq (mod $i 3) 0 }}
          <div>{{ hourAt $w $i }}<br>{{ flag (hourCodeAt $w $i) }}<br>{{ printf "%.0f" (tempAt $w $i) }}°</div>
        {{ end }}{{ end }}
      </div>
      <div class="days">
        {{ range .Daily }}
          <div class="day">
            <span>{{ weekday .Date }} {{ .Date }}</span>
            <span>{{ flag .Code }} {{ desc .Code }}</span>
            <span class="big">{{ printf "%.0f" .TempMax }}°/{{ printf "%.0f" .TempMin }}°</span>
            <span>风{{ printf "%.0f" .WindMax }} · 降水{{ printf "%.0f" .PrecipProb }}%</span>
          </div>
        {{ end }}
      </div>
    </div>
  {{ end }}
  </div>
</body>
</html>`

func init() {
	weatherWebCmd.Flags().StringP("port", "p", "8080", "网页监听端口")
	weatherCmd.AddCommand(weatherWebCmd)
	rootCmd.AddCommand(weatherCmd)
}
