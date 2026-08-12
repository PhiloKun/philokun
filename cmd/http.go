package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// httpCmd 是 HTTP 请求快捷版，支持 GET/POST，打印状态码、响应头与 body 摘要。
var httpCmd = &cobra.Command{
	Use:   "http <方法> <URL>",
	Short: "HTTP 请求快捷版（GET/POST 等）",
	Long: `发送一个 HTTP 请求并展示状态码、响应头与响应体摘要。

示例:
  philokun http GET  https://example.com
  philokun http POST https://httpbin.org/post -d '{"a":1}' -c application/json`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := strings.ToUpper(args[0])
		url := args[1]
		body, _ := cmd.Flags().GetString("data")
		ct, _ := cmd.Flags().GetString("content-type")

		var reqBody io.Reader
		if body != "" {
			reqBody = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return fmt.Errorf("构造请求失败: %w", err)
		}
		if body != "" && ct != "" {
			req.Header.Set("Content-Type", ct)
		}

		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}
		defer resp.Body.Close()

		fmt.Printf("状态码: %s\n", resp.Status)
		fmt.Println("响应头:")
		for k, v := range resp.Header {
			fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if err != nil {
			return fmt.Errorf("读取响应体失败: %w", err)
		}
		fmt.Printf("\n响应体 (%d 字节，最多显示 4096):\n%s\n", len(respBody), string(respBody))
		return nil
	},
}

func init() {
	httpCmd.Flags().StringP("data", "d", "", "请求体（POST/PUT 等使用）")
	httpCmd.Flags().StringP("content-type", "c", "application/json", "Content-Type（有 -d 时生效）")
	rootCmd.AddCommand(httpCmd)
}
