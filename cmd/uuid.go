package cmd

import (
	"crypto/rand"
	"fmt"

	"github.com/spf13/cobra"
)

// uuidCmd 批量生成 UUID v4。
var uuidCmd = &cobra.Command{
	Use:   "uuid",
	Short: "生成 UUID v4",
	Long: `生成指定数量的随机 UUID v4（RFC 4122）。

示例:
  philokun uuid          # 生成 1 个
  philokun uuid -n 5     # 批量生成 5 个`,
	RunE: func(cmd *cobra.Command, args []string) error {
		n, _ := cmd.Flags().GetInt("count")
		if n <= 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			u, err := newUUIDv4()
			if err != nil {
				return err
			}
			fmt.Println(u)
		}
		return nil
	},
}

// newUUIDv4 用 crypto/rand 生成 RFC 4122 v4 UUID。
func newUUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // 版本 4
	b[8] = (b[8] & 0x3f) | 0x80 // 变体位
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func init() {
	uuidCmd.Flags().IntP("count", "n", 1, "生成数量")
	rootCmd.AddCommand(uuidCmd)
}
