package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// unitCmd 常用单位换算（长度 / 重量 / 温度）。
var unitCmd = &cobra.Command{
	Use:   "unit",
	Short: "单位换算（长度 / 重量 / 温度）",
	Long: `在同类单位之间换算。

类别:
  length  长度:  m / cm / km / mm / in / ft / mi / yd
  weight  重量:  kg / g / t / lb / oz
  temp    温度:  c / f / k

示例:
  philokun unit length 1 m cm
  philokun unit weight 1 kg lb
  philokun unit temp 100 c f`,
	Args: cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		cat := strings.ToLower(args[0])
		val, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("非法数值: %q", args[1])
		}
		from := strings.ToLower(args[2])
		to := strings.ToLower(args[3])

		out, err := convertUnit(cat, val, from, to)
		if err != nil {
			return err
		}
		// 整数不显示小数
		if out == float64(int64(out)) {
			fmt.Printf("%d %s\n", int64(out), to)
		} else {
			fmt.Printf("%.6g %s\n", out, to)
		}
		return nil
	},
}

// convertUnit 按类别换算单位。
func convertUnit(cat string, val float64, from, to string) (float64, error) {
	switch cat {
	case "length":
		meters, ok := lengthToMeter(val, from)
		if !ok {
			return 0, fmt.Errorf("未知长度单位: %s", from)
		}
		out, ok := meterTo(lengthToMap, to)
		if !ok {
			return 0, fmt.Errorf("未知长度单位: %s", to)
		}
		return meters * out, nil
	case "weight":
		kilos, ok := weightToKg(val, from)
		if !ok {
			return 0, fmt.Errorf("未知重量单位: %s", from)
		}
		out, ok := kgTo(weightFromMap, to)
		if !ok {
			return 0, fmt.Errorf("未知重量单位: %s", to)
		}
		return kilos * out, nil
	case "temp":
		return convertTemp(val, from, to)
	default:
		return 0, fmt.Errorf("未知类别: %s（支持 length / weight / temp）", cat)
	}
}

// 长度基准：1 单位 = ? 米
var lengthToMap = map[string]float64{
	"m":   1, "km": 1000, "cm": 0.01, "mm": 0.001,
	"in": 0.0254, "ft": 0.3048, "mi": 1609.344, "yd": 0.9144,
}

func lengthToMeter(v float64, u string) (float64, bool) {
	f, ok := lengthToMap[u]
	return v * f, ok
}

func meterTo(m map[string]float64, u string) (float64, bool) {
	f, ok := m[u]
	return f, ok
}

// 重量基准：1 单位 = ? 千克
var weightFromMap = map[string]float64{
	"kg": 1, "g": 0.001, "t": 1000, "lb": 0.45359237, "oz": 0.028349523125,
}

func weightToKg(v float64, u string) (float64, bool) {
	f, ok := weightFromMap[u]
	return v * f, ok
}

func kgTo(m map[string]float64, u string) (float64, bool) {
	f, ok := m[u]
	return f, ok
}

// convertTemp 温度换算（先转摄氏度再转目标）。
func convertTemp(v float64, from, to string) (float64, error) {
	var c float64
	switch from {
	case "c":
		c = v
	case "f":
		c = (v - 32) * 5 / 9
	case "k":
		c = v - 273.15
	default:
		return 0, fmt.Errorf("未知温度单位: %s", from)
	}
	switch to {
	case "c":
		return c, nil
	case "f":
		return c*9/5 + 32, nil
	case "k":
		return c + 273.15, nil
	}
	return 0, fmt.Errorf("未知温度单位: %s", to)
}

func init() {
	rootCmd.AddCommand(unitCmd)
}
