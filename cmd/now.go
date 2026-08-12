package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// nowCmd 显示当前时间，以及时间戳 <-> 日期互转。
var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "显示当前时间 / 时间戳转换",
	Long: `显示当前时间，或在 Unix 时间戳与日期之间互转。

示例:
  philokun now                 # 当前本地时间 + 时间戳
  philokun now -u              # 当前 UTC 时间
  philokun now 1690000000      # 时间戳 -> 日期
  philokun now 1690000000 -u   # 时间戳 -> UTC 日期
  philokun now 2023-07-22T12:00:00   # 日期 -> 时间戳`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		utc, _ := cmd.Flags().GetBool("utc")
		loc := time.Local
		if utc {
			loc = time.UTC
		}

		if len(args) == 0 {
			now := time.Now().In(loc)
			fmt.Println(now.Format("2006-01-02 15:04:05"))
			fmt.Printf("时间戳: %d\n", now.Unix())
			fmt.Printf("毫秒:   %d\n", now.UnixMilli())
			return nil
		}

		arg := args[0]
		// 纯数字 -> 时间戳转日期
		if ts, err := strconv.ParseInt(arg, 10, 64); err == nil {
			var t time.Time
			if len(arg) >= 13 { // 毫秒级
				t = time.UnixMilli(ts).In(loc)
			} else {
				t = time.Unix(ts, 0).In(loc)
			}
			fmt.Println(t.Format("2006-01-02 15:04:05"))
			return nil
		}

		// 日期字符串 -> 时间戳
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
			time.RFC3339,
		}
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, strings.TrimSpace(arg), loc); err == nil {
				fmt.Printf("时间戳: %d\n", t.Unix())
				fmt.Printf("毫秒:   %d\n", t.UnixMilli())
				return nil
			}
		}
		return fmt.Errorf("无法解析为时间戳或日期: %q", arg)
	},
}

func init() {
	nowCmd.Flags().BoolP("utc", "u", false, "使用 UTC 时区")
	rootCmd.AddCommand(nowCmd)
}
