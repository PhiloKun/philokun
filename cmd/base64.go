package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// base64Cmd 对文本做 base64 编码或解码。
var base64Cmd = &cobra.Command{
	Use:   "base64 <encode|decode> <文本>",
	Short: "base64 编码 / 解码",
	Long: `对文本做 base64 编码或解码。

示例:
  philokun base64 encode "hello world"
  philokun base64 decode "aGVsbG8gd29ybGQ="
  echo hello | philokun base64 encode -   # 从 stdin 读取`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := strings.ToLower(args[0])
		if mode != "encode" && mode != "decode" {
			return fmt.Errorf("第一个参数必须是 encode 或 decode")
		}
		var input string
		if len(args) >= 2 {
			input = args[1]
			if input == "-" {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("读取 stdin 失败: %w", err)
				}
				input = strings.TrimRight(string(b), "\n")
			}
		} else {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("读取 stdin 失败: %w", err)
			}
			input = strings.TrimRight(string(b), "\n")
		}

		switch mode {
		case "encode":
			fmt.Println(base64.StdEncoding.EncodeToString([]byte(input)))
		case "decode":
			dec, err := base64.StdEncoding.DecodeString(input)
			if err != nil {
				return fmt.Errorf("解码失败（输入不是合法 base64）: %w", err)
			}
			fmt.Println(string(dec))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(base64Cmd)
}
