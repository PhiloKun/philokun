package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"philokun/internal/store"
)

var (
	shortenCustom string
	shortenPort   string
	shortenHost   string
)

// shortenCmd 是 URL 短链服务分组命令。
var shortenCmd = &cobra.Command{
	Use:   "shorten",
	Short: "URL 短链服务（创建/查询/重定向）",
	Long: `把长链接压缩成易分享的短码，并提供本地 HTTP 重定向服务。
所有短链以 JSON 存于 ~/.philokun/shorts.json，纯本地、零外部依赖。`,
}

// shortenCreateCmd 创建一条短链。
var shortenCreateCmd = &cobra.Command{
	Use:   "create <长链接>",
	Short: "把长链接生成短链",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code, err := store.CreateShort(args[0], shortenCustom)
		if err != nil {
			return err
		}
		fmt.Printf("短链已创建: %s  ->  %s\n", code, args[0])
		return nil
	},
}

// shortenGetCmd 查询某短码对应的原始链接。
var shortenGetCmd = &cobra.Command{
	Use:   "get <短码>",
	Short: "查询短码对应的原始链接",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := store.GetShort(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("短码: %s\n原始链接: %s\n创建时间: %s\n点击数: %d\n",
			info.Code, info.URL, info.CreatedAt.Format("2006-01-02 15:04"), info.Clicks)
		return nil
	},
}

// shortenListCmd 列出全部短链。
var shortenListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出全部短链",
	RunE: func(cmd *cobra.Command, args []string) error {
		infos, err := store.ListShorts()
		if err != nil {
			return err
		}
		if len(infos) == 0 {
			fmt.Println("暂无短链。用 `philokun shorten create <链接>` 创建第一条吧。")
			return nil
		}
		for _, s := range infos {
			fmt.Printf("%s -> %s  (点击 %d)\n", s.Code, s.URL, s.Clicks)
		}
		return nil
	},
}

// shortenRmCmd 删除一条短链。
var shortenRmCmd = &cobra.Command{
	Use:   "rm <短码>",
	Short: "删除一条短链",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ok, err := store.RmShort(args[0])
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("短码 %q 不存在", args[0])
		}
		fmt.Printf("已删除短码: %s\n", args[0])
		return nil
	},
}

// shortenServeCmd 启动本地 HTTP 重定向服务：访问 /<短码> 即 302 跳转到原始链接。
var shortenServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动本地短链重定向服务",
	Long: `启动一个本地 HTTP 服务：浏览器访问 http://<host>:<port>/<短码> 即跳转到原始链接。
首页 / 返回简单的使用说明。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("%s:%s", shortenHost, shortenPort)
		fmt.Printf("短链重定向服务已启动: http://%s  (Ctrl+C 退出)\n", addr)
		fmt.Println("用法示例: 在浏览器打开 http://" + addr + "/你的短码")
		mux := http.NewServeMux()
		mux.HandleFunc("/", shortenRedirectHandler)
		server := &http.Server{Addr: addr, Handler: mux}
		if err := server.ListenAndServe(); err != nil {
			return fmt.Errorf("启动服务失败: %w", err)
		}
		return nil
	},
}

// shortenRedirectHandler 处理重定向：根路径返回说明，其余按首段路径作为短码解析。
func shortenRedirectHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	path = strings.Trim(path, "/")
	if path == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "philokun 短链服务\n访问 /<短码> 即跳转到原始链接。\n")
		return
	}
	target, err := store.ResolveShort(path)
	if err != nil {
		http.Error(w, "短码不存在: "+path, http.StatusNotFound)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func init() {
	shortenCreateCmd.Flags().StringVarP(&shortenCustom, "code", "c", "", "自定义短码（默认自动生成）")
	shortenServeCmd.Flags().StringVarP(&shortenPort, "port", "p", "9010", "重定向服务监听端口")
	shortenServeCmd.Flags().StringVar(&shortenHost, "host", "127.0.0.1", "重定向服务监听地址")

	shortenCmd.AddCommand(shortenCreateCmd)
	shortenCmd.AddCommand(shortenGetCmd)
	shortenCmd.AddCommand(shortenListCmd)
	shortenCmd.AddCommand(shortenRmCmd)
	shortenCmd.AddCommand(shortenServeCmd)
	rootCmd.AddCommand(shortenCmd)
}
