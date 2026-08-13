package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// ageCmd 计算年龄或两个日期之间的差。
var ageCmd = &cobra.Command{
	Use:   "age <出生日期> [结束日期]",
	Short: "计算年龄 / 两个日期之间的差",
	Long: `计算从某日期到今天（或到指定结束日期）的年龄 / 时长。

日期支持 2006-01-02、2006/01/02 或 20060102。
省略结束日期则默认到今天。

示例:
  philokun age 1990-05-20
  philokun age 1990-05-20 2025-01-01
  philokun age 2024-01-01 2025-06-15   # 两个日期差`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, err := parseDateFlex(args[0])
		if err != nil {
			return err
		}
		to := time.Now()
		if len(args) == 2 {
			to, err = parseDateFlex(args[1])
			if err != nil {
				return err
			}
		}
		if to.Before(from) {
			return fmt.Errorf("结束日期不能早于开始日期")
		}
		printDateDiff(from, to)
		return nil
	},
}

// parseDateFlex 解析多种格式的日期。
func parseDateFlex(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", "2006/01/02", "20060102", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析日期: %q（支持 2006-01-02 / 2006/01/02 / 20060102）", s)
}

// printDateDiff 打印年/月/日差以及总天数。
func printDateDiff(from, to time.Time) {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	days := to.Day() - from.Day()
	if days < 0 {
		months--
		days += daysInMonth(from.Year(), from.Month())
	}
	if months < 0 {
		years--
		months += 12
	}
	total := int(to.Sub(from).Hours() / 24)
	fmt.Printf("从 %s 到 %s：\n", from.Format("2006-01-02"), to.Format("2006-01-02"))
	if years > 0 {
		fmt.Printf("  %d 年 %d 个月 %d 天\n", years, months, days)
	} else if months > 0 {
		fmt.Printf("  %d 个月 %d 天\n", months, days)
	} else {
		fmt.Printf("  %d 天\n", days)
	}
	fmt.Printf("  合计约 %d 天\n", total)
}

// daysInMonth 返回某年某月的天数。
func daysInMonth(year int, month time.Month) int {
	switch month {
	case time.February:
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			return 29
		}
		return 28
	case time.April, time.June, time.September, time.November:
		return 30
	default:
		return 31
	}
}

func init() {
	rootCmd.AddCommand(ageCmd)
}
