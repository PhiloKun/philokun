package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"philokun/internal/store"
)

// urlCmd 是常用链接管理分组命令。
var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "本地 URL 书签管理",
}

// urlAddCmd 新增一条别名链接。
var urlAddCmd = &cobra.Command{
	Use:   "add <别名> <链接>",
	Short: "添加一条别名链接",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias, link := args[0], args[1]
		if err := store.AddURL(alias, link); err != nil {
			return err
		}
		fmt.Printf("已保存 %q -> %s\n", alias, link)
		return nil
	},
}

// urlListCmd 列出全部链接。
var urlListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出全部链接",
	RunE: func(cmd *cobra.Command, args []string) error {
		urls, err := store.ListURLs()
		if err != nil {
			return err
		}
		if len(urls) == 0 {
			fmt.Println("暂无链接。用 `philokun url add 别名 链接` 添加第一条吧。")
			return nil
		}
		for _, u := range urls {
			fmt.Printf("%d. %s -> %s\n", u.ID, u.Alias, u.Link)
		}
		return nil
	},
}

// urlOpenCmd 在浏览器中打开指定别名链接（macOS 用 open）。
var urlOpenCmd = &cobra.Command{
	Use:   "open <别名>",
	Short: "在浏览器中打开别名对应的链接",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		u, err := store.GetURL(args[0])
		if err != nil {
			return err
		}
		// 跨平台打开命令：macOS=open, Windows=cmd start, 其余=xdg-open。
		var cmdOpen *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmdOpen = exec.Command("open", u.Link)
		case "windows":
			cmdOpen = exec.Command("cmd", "/c", "start", u.Link)
		default:
			cmdOpen = exec.Command("xdg-open", u.Link)
		}
		if err := cmdOpen.Start(); err != nil {
			return fmt.Errorf("打开链接失败: %w", err)
		}
		fmt.Printf("正在打开: %s\n", u.Link)
		return nil
	},
}

func init() {
	urlCmd.AddCommand(urlAddCmd)
	urlCmd.AddCommand(urlListCmd)
	urlCmd.AddCommand(urlOpenCmd)
	rootCmd.AddCommand(urlCmd)
}
