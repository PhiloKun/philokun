package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// pomodoroCmd 启动一个终端番茄钟/倒计时，结束时响铃提醒。
var pomodoroCmd = &cobra.Command{
	Use:     "pomodoro <分钟数>",
	Aliases: []string{"pomo"},
	Short:   "启动番茄钟倒计时，结束响铃提醒",
	Args:    cobra.ExactArgs(1),
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

		// 倒计时归零：播放提示音并发出系统通知。
		alertBell()
		fmt.Println("⏰ 时间到！休息一下吧。")
		return nil
	},
}

// alertBell 在倒计时结束时发出提醒。
// 优先使用系统自带能力播放提示音 + 桌面通知；若都不可用，则回退到终端 BEL(\a)。
func alertBell() {
	msg := "番茄钟时间到，休息一下吧！"

	switch runtime.GOOS {
	case "darwin":
		// 播放系统提示音，并弹出桌面通知。
		_ = exec.Command("afplay", "/System/Library/Sounds/Ping.aiff").Run()
		_ = exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification "%s" with title "🍅 番茄钟"`, msg)).Run()
	case "linux":
		// 优先用 paplay/aplay 播放，再用 notify-send 弹通知。
		for _, bin := range []string{"paplay", "aplay"} {
			if _, err := exec.LookPath(bin); err == nil {
				_ = exec.Command(bin, "/usr/share/sounds/freedesktop/stereo/complete.oga").Run()
				break
			}
		}
		if _, err := exec.LookPath("notify-send"); err == nil {
			_ = exec.Command("notify-send", "🍅 番茄钟", msg).Run()
		}
	case "windows":
		// 通过 PowerShell 播放系统提示音并弹出 Toast 通知。
		ps := fmt.Sprintf(`[System.Media.SystemSounds]::Beep.Play(); `+
			`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType=WindowsRuntime] > $null`)
		_ = exec.Command("powershell", "-NoProfile", "-Command", ps).Run()
	}

	// 兜底：终端 BEL，部分终端会发出蜂鸣。
	fmt.Print("\a")
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
