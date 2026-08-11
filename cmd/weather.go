package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// weatherCmd 查询指定城市的当天天气。
var weatherCmd = &cobra.Command{
	Use:   "weather <城市>",
	Short: "查询城市当天天气（Open-Meteo 免 key 接口）",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		city := strings.Join(args, " ")

		// 1) 地理编码：城市名 -> 经纬度
		geo, err := geocode(city)
		if err != nil {
			return err
		}
		if geo.Latitude == 0 && geo.Longitude == 0 {
			return fmt.Errorf("找不到城市 %q", city)
		}

		// 2) 当前天气
		w, err := currentWeather(geo.Latitude, geo.Longitude)
		if err != nil {
			return err
		}

		fmt.Printf("城市: %s\n", geo.Name)
		fmt.Printf("温度: %.1f°C\n", w.Temperature)
		fmt.Printf("风速: %.1f km/h\n", w.Windspeed)
		fmt.Printf("天气代码: %d (%s)\n", w.Weathercode, describeWeather(w.Weathercode))
		return nil
	},
}

// geocodingResp 是 Open-Meteo 地理编码接口的（部分）响应结构。
type geocodingResp struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

// geoResult 是解析后的地理编码结果。
type geoResult struct {
	Name      string
	Latitude  float64
	Longitude float64
}

// weatherResp 是 Open-Meteo forecast 接口的（部分）响应结构。
type weatherResp struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
		Weathercode int     `json:"weathercode"`
		Time        string  `json:"time"`
	} `json:"current_weather"`
}

// weatherResult 是解析后的天气结果。
type weatherResult struct {
	Temperature float64
	Windspeed   float64
	Weathercode int
}

const (
	geocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL  = "https://api.open-meteo.com/v1/forecast"
)

// httpGetJSON 发起 GET 请求并解析 JSON 到 out（out 须为指针）。
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

// geocode 把城市名解析为经纬度。
func geocode(city string) (geoResult, error) {
	u := fmt.Sprintf("%s?name=%s&count=1&language=zh", geocodingURL, url.QueryEscape(city))
	var gr geocodingResp
	if err := httpGetJSON(u, &gr); err != nil {
		return geoResult{}, err
	}
	if len(gr.Results) == 0 {
		return geoResult{}, nil
	}
	r := gr.Results[0]
	return geoResult{Name: r.Name, Latitude: r.Latitude, Longitude: r.Longitude}, nil
}

// currentWeather 查询指定经纬度的当前天气。
func currentWeather(lat, lon float64) (weatherResult, error) {
	u := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f&current_weather=true", forecastURL, lat, lon)
	var wr weatherResp
	if err := httpGetJSON(u, &wr); err != nil {
		return weatherResult{}, err
	}
	return weatherResult{
		Temperature: wr.CurrentWeather.Temperature,
		Windspeed:   wr.CurrentWeather.Windspeed,
		Weathercode: wr.CurrentWeather.Weathercode,
	}, nil
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

func init() {
	rootCmd.AddCommand(weatherCmd)
}
