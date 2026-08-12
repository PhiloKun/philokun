package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/philokun/internal/store"
)

//go:embed qrcode-web/*
var qrcodeWebFS embed.FS

var (
	qrcodePort string
	qrcodeHost string
)

// qrcodeCmd 启动一个本地 Web 服务，提供“输入即生成”的二维码工具页面。
var qrcodeCmd = &cobra.Command{
	Use:   "qrcode",
	Short: "启动二维码生成器（网页版）",
	Long: `打开一个本地网页，输入文本/网址/数字即可实时生成二维码，
支持下载 PNG 与复制到剪贴板，界面自适应桌面与移动端。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("%s:%s", qrcodeHost, qrcodePort)
		fmt.Printf("二维码生成器已启动，请在浏览器打开: http://%s\n按 Ctrl+C 退出。\n", addr)

		// 把嵌入的 qrcode-web 目录挂到根路径。
		sub, err := fs.Sub(qrcodeWebFS, "qrcode-web")
		if err != nil {
			return err
		}
		fileServer := http.FileServer(http.FS(sub))
		mux := http.NewServeMux()
		mux.Handle("/", fileServer)
		mux.HandleFunc("/api/qrcode", qrcodeHandler)

		server := &http.Server{Addr: addr, Handler: mux}
		if err := server.ListenAndServe(); err != nil {
			return fmt.Errorf("启动服务失败: %w", err)
		}
		return nil
	},
}

// qrcodeHandler 接收 ?text= 参数，返回 PNG 二维码；空内容返回 204 让前端隐藏图像。
func qrcodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET", http.StatusMethodNotAllowed)
		return
	}
	text := strings.TrimSpace(r.URL.Query().Get("text"))
	if text == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	png, err := store.GenerateQRCode(text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(png)
}

func init() {
	qrcodeCmd.Flags().StringVarP(&qrcodePort, "port", "p", "9000", "Web 服务监听端口")
	qrcodeCmd.Flags().StringVar(&qrcodeHost, "host", "127.0.0.1", "Web 服务监听地址")
	rootCmd.AddCommand(qrcodeCmd)
}
