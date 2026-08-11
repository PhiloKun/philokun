package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// pomodoroCmd 启动一个终端番茄钟/倒计时，结束时响铃。
var pomodoroCmd = &cobra.Command{
	Use:   "pomodoro <分钟数>",
	Short: "启动番茄钟倒计时，结束响铃提醒",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mins, err := strconv.Atoi(args[0])
		if err != nil || mins <= 0 {
			return fmt.Errorf("分钟数必须是正整数")
		}
		duration := time.Duration(mins) * time.Minute
		end := time.Now().Add(duration)

		fmt.Printf("🍅 番茄钟开始，专注 %d 分钟，结束时间 %s\n", mins, end.Format("15:04:05"))

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		// 用每秒刷新一行的方式显示剩余时间。
		for remaining := duration; ; remaining = time.Until(end) {
			fmt.Printf("\r剩余: %s ", formatDuration(remaining))
			if remaining <= 0 {
				break
			}
			<-ticker.C
		}
		fmt.Print("\r")
		// 响铃：\a 让终端发出提示音。
		fmt.Println("\a⏰ 时间到！休息一下吧。")
		return nil
	},
}

// formatDuration 把时长格式化为 MM:SS。
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func init() {
	rootCmd.AddCommand(pomodoroCmd)
}
