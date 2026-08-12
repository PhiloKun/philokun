package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// timerCmd 倒计时提醒（终端响铃 + 系统通知）。
var timerCmd = &cobra.Command{
	Use:   "timer <时长>",
	Short: "倒计时提醒",
	Long: `倒计时结束后提醒你（终端响铃 + 系统通知）。

时长支持: 秒(s) / 分(m) / 时(h)，也可直接写数字（按秒）。
示例:
  philokun timer 25m        # 25 分钟
  philokun timer 1h30m      # 1 小时 30 分
  philokun timer 90         # 90 秒
  philokun timer 10m 泡茶   # 带提醒文字`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := parseDuration(args[0])
		if err != nil {
			return err
		}
		msg := "时间到！"
		if len(args) > 1 {
			msg = strings.Join(args[1:], " ")
		}

		end := time.Now().Add(d)
		fmt.Printf("⏳ 倒计时 %s，结束于 %s\n", d.Round(time.Second), end.Format("15:04:05"))
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for remaining := d; ; {
			select {
			case <-ticker.C:
				remaining -= time.Second
				if remaining <= 0 {
					notify(msg)
					return nil
				}
			}
		}
	},
}

// parseDuration 解析支持 h/m/s 的时长字符串。
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("时长不能为空")
	}
	// 纯数字 -> 秒
	if n, err := strconv.Atoi(s); err == nil {
		return time.Duration(n) * time.Second, nil
	}
	// 解析组合形式，如 1h30m / 25m / 90s
	var total time.Duration
	var num strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			num.WriteRune(r)
			continue
		}
		if num.Len() == 0 {
			return 0, fmt.Errorf("非法时长: %q", s)
		}
		v, err := strconv.Atoi(num.String())
		if err != nil {
			return 0, fmt.Errorf("非法时长: %q", s)
		}
		num.Reset()
		switch r {
		case 'h', 'H':
			total += time.Duration(v) * time.Hour
		case 'm', 'M':
			total += time.Duration(v) * time.Minute
		case 's', 'S':
			total += time.Duration(v) * time.Second
		default:
			return 0, fmt.Errorf("未知单位: %q", string(r))
		}
	}
	if num.Len() != 0 {
		return 0, fmt.Errorf("时长缺少单位: %q", s)
	}
	if total <= 0 {
		return 0, fmt.Errorf("时长必须大于 0")
	}
	return total, nil
}

// notify 倒计时结束时提醒：响铃 + 尽量弹出系统通知。
func notify(msg string) {
	fmt.Print("\a") // 终端响铃
	fmt.Printf("\n🔔 %s\n", msg)
	// 尝试系统通知（忽略失败，终端响铃仍可用）
	switch {
	case lookPath("osascript") != "": // macOS
		_ = exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification "%s" with title "philokun 提醒"`, msg)).Run()
	case lookPath("notify-send") != "": // Linux
		_ = exec.Command("notify-send", "philokun 提醒", msg).Run()
	}
}

// lookPath 包装 exec.LookPath，找不到时返回空串。
func lookPath(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}

func init() {
	rootCmd.AddCommand(timerCmd)
}
