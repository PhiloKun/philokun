package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version 是当前的版本号，以后发布新版本时改这里即可。
// 进阶做法：用 -ldflags "-X github.com/philokun/cmd.version=1.2.3" 在编译时注入，
// 这样不用改源码就能打不同版本的二进制。
var version = "1.2.0"

// versionCmd 演示“最简单的子命令”：只打印信息，不需要参数。
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印 philokun 的版本号",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("philokun version %s\n", version)
	},
}

// init 在包加载时自动执行，把 versionCmd 挂到 rootCmd 下。
// 每个命令文件都用自己的 init() 来“注册”自己，互不干扰。
func init() {
	rootCmd.AddCommand(versionCmd)
}
