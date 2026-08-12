package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
)

// randCmd 生成随机字符串或随机数字。
var randCmd = &cobra.Command{
	Use:   "rand",
	Short: "生成随机字符串 / 数字",
	Long: `生成随机字符串或指定范围内的随机整数，使用 crypto/rand（密码学安全）。

示例:
  philokun rand -l 16           # 16 位随机字符串（默认字符集）
  philokun rand -l 32 -c hex    # 32 位十六进制串
  philokun rand -n 1 100        # [1,100] 之间的随机整数
  philokun rand -i 5 -n 1 100   # 批量 5 个随机整数`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		length, _ := cmd.Flags().GetInt("length")
		charset, _ := cmd.Flags().GetString("charset")
		count, _ := cmd.Flags().GetInt("count")

		if len(args) == 2 {
			// 随机整数模式: rand -n MIN MAX
			min, err := parseIntArg(args[0])
			if err != nil {
				return fmt.Errorf("无效的最小值: %w", err)
			}
			max, err := parseIntArg(args[1])
			if err != nil {
				return fmt.Errorf("无效的最大值: %w", err)
			}
			if max < min {
				return fmt.Errorf("最大值必须 >= 最小值")
			}
			if count <= 0 {
				count = 1
			}
			for i := 0; i < count; i++ {
				v, err := randIntRange(min, max)
				if err != nil {
					return err
				}
				fmt.Println(v)
			}
			return nil
		}

		// 随机字符串模式
		if length <= 0 {
			length = 16
		}
		cs := charsetFor(charset)
		if count <= 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			s, err := randString(length, cs)
			if err != nil {
				return err
			}
			fmt.Println(s)
		}
		return nil
	},
}

// charsetFor 根据名称返回字符集。
func charsetFor(name string) string {
	switch name {
	case "hex":
		return "0123456789abcdef"
	case "alpha":
		return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "num":
		return "0123456789"
	case "alnum":
		return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	default: // 默认：去易混淆字符的可读字符集
		return "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	}
}

// randString 生成指定长度的随机字符串。
func randString(n int, charset string) (string, error) {
	out := make([]byte, n)
	max := big.NewInt(int64(len(charset)))
	for i := range out {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = charset[idx.Int64()]
	}
	return string(out), nil
}

// randIntRange 生成 [min, max] 闭区间内的随机整数。
func randIntRange(min, max int64) (int64, error) {
	span := max - min + 1
	v, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		return 0, err
	}
	return min + v.Int64(), nil
}

func parseIntArg(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func init() {
	randCmd.Flags().IntP("length", "l", 16, "随机字符串长度")
	randCmd.Flags().StringP("charset", "c", "default", "字符集: default / hex / alpha / num / alnum")
	randCmd.Flags().IntP("count", "n", 1, "生成数量")
	rootCmd.AddCommand(randCmd)
}
