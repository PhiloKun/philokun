package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// jsonCmd JSON 格式化 / 压缩 / 校验工具。
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "JSON 格式化、压缩与校验",
	Long: `处理 JSON 文本：格式化（缩进）、压缩成一行、或仅做语法校验。

示例:
  philokun json fmt                    # 从 stdin 读，格式化输出
  philokun json fmt data.json          # 格式化文件
  philokun json min data.json          # 压缩成一行
  philokun json check data.json        # 仅校验语法，合法则输出 OK`,
}

// jsonFmtCmd 把 JSON 美化（带缩进）。
var jsonFmtCmd = &cobra.Command{
	Use:   "fmt [文件]",
	Short: "格式化（美化）JSON，带缩进",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readJSONInput(args)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return fmt.Errorf("JSON 格式错误: %w", err)
		}
		fmt.Println(buf.String())
		return nil
	},
}

// jsonMinCmd 把 JSON 压缩成一行。
var jsonMinCmd = &cobra.Command{
	Use:   "min [文件]",
	Short: "压缩 JSON 成一行（去除空白）",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readJSONInput(args)
		if err != nil {
			return err
		}
		compacted, err := compactJSON(data)
		if err != nil {
			return fmt.Errorf("JSON 格式错误: %w", err)
		}
		fmt.Println(string(compacted))
		return nil
	},
}

// jsonCheckCmd 仅做语法校验。
var jsonCheckCmd = &cobra.Command{
	Use:   "check [文件]",
	Short: "仅校验 JSON 语法是否合法",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readJSONInput(args)
		if err != nil {
			return err
		}
		if !json.Valid(data) {
			return fmt.Errorf("JSON 语法不合法")
		}
		fmt.Println("OK")
		return nil
	},
}

// readJSONInput 从文件或 stdin 读取 JSON 内容。
func readJSONInput(args []string) ([]byte, error) {
	if len(args) == 1 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}

// compactJSON 去掉 JSON 中的空白字符。
func compactJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func init() {
	jsonCmd.AddCommand(jsonFmtCmd)
	jsonCmd.AddCommand(jsonMinCmd)
	jsonCmd.AddCommand(jsonCheckCmd)
	rootCmd.AddCommand(jsonCmd)
}
