package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// serveCmd 启动一个临时静态文件服务器，方便分享本地目录。
var serveCmd = &cobra.Command{
	Use:   "serve [目录]",
	Short: "启动临时静态文件服务器",
	Long: `把当前目录（或指定目录）作为一个静态文件服务器跑起来，方便在局域网分享文件。

示例:
  philokun serve                  # 服务当前目录，默认 8080 端口
  philokun serve ./dist -p 9000  # 服务指定目录与端口`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")

		abs, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return fmt.Errorf("目录不存在: %s", dir)
		}
		if !info.IsDir() {
			return fmt.Errorf("不是目录: %s", dir)
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		fmt.Printf("在 http://%s 提供 %s\n（Ctrl+C 停止）\n", addr, abs)
		return http.ListenAndServe(addr, http.FileServer(http.Dir(abs)))
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "监听端口")
	serveCmd.Flags().String("host", "0.0.0.0", "监听地址")
	rootCmd.AddCommand(serveCmd)
}
