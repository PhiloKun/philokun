package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// slugCmd 把任意文本转成 URL 友好的 slug。
var slugCmd = &cobra.Command{
	Use:   "slug <文本>",
	Short: "把文本转成 URL 友好的 slug",
	Long: `将标题、中文或带符号的文本转换为 URL 友好的 slug。

特性：
  - 小写化、去重连字符、去首尾连字符
  - 非字母数字（CJK 除外）转为连字符
  - 可用 -s 指定分隔符（默认 -）

示例:
  philokun slug "Hello, World!"
  philokun slug "如何学习 Go 语言" -s _
  philokun slug "My Post Title" -s -`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sep, _ := cmd.Flags().GetString("sep")
		if sep == "" {
			sep = "-"
		}
		fmt.Println(toSlug(args[0], sep))
		return nil
	},
}

// toSlug 把文本转成 slug。CJK 字符直接保留（按 UTF-8 逐 rune 处理）。
func toSlug(s, sep string) string {
	var b strings.Builder
	prevSep := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevSep = false
		case r >= 0x4e00 && r <= 0x9fff: // CJK 统一表意文字
			b.WriteRune(r)
			prevSep = false
		case r == ' ' || r == '-' || r == '_' || r == '/' || r == '.' || r == ',' || r == '&' || r == '+' || r == '#' || r == '%' || r == '?' || r == '!' || r == '(' || r == ')':
			if !prevSep && b.Len() > 0 {
				b.WriteString(sep)
			}
			prevSep = true
		default:
			if !prevSep && b.Len() > 0 {
				b.WriteString(sep)
			}
			prevSep = true
		}
	}
	out := b.String()
	out = strings.Trim(out, sep)
	// 合并连续分隔符
	out = regexp.MustCompile(regexp.QuoteMeta(sep)+"+").ReplaceAllString(out, sep)
	return out
}

func init() {
	slugCmd.Flags().StringP("sep", "s", "-", "分隔符（默认 -）")
	rootCmd.AddCommand(slugCmd)
}
