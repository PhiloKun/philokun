package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 是 philokun 的根命令，所有子命令都挂在它下面。
// Use    -> 命令名（也是可执行文件名）
// Short  -> 一句话简介，显示在 `philokun --help` 里
// Long   -> 更长的说明，显示在具体命令的帮助里
var rootCmd = &cobra.Command{
	Use:   "philokun",
	Short: "philokun 是一个个人效率命令行工具",
	Long: `philokun 帮你用命令行管理待办、笔记和日志。
轻量、可脚本化，所有数据都存在本地，不依赖任何云服务。`,
}

// Execute 是程序的唯一入口，由 main.go 调用。
// 命令执行出错时把错误打到 stderr 并以非零码退出，符合 Unix 习惯。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
