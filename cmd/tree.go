package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// treeCmd 以树状结构展示目录。
var treeCmd = &cobra.Command{
	Use:   "tree [目录]",
	Short: "以树状结构展示目录",
	Long: `递归打印目录树，默认从当前目录开始。

示例:
  philokun tree
  philokun tree ./cmd
  philokun tree -L 2          # 最大深度 2
  philokun tree -a            # 包含隐藏文件
  philokun tree -d            # 只看目录`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		maxDepth, _ := cmd.Flags().GetInt("depth")
		all, _ := cmd.Flags().GetBool("all")
		dirsOnly, _ := cmd.Flags().GetBool("dirs-only")

		info, err := os.Stat(root)
		if err != nil {
			return err
		}
		fmt.Println(info.Name())
		printTree(root, "", maxDepth, all, dirsOnly)
		return nil
	},
}

// printTree 递归打印目录树。
func printTree(dir, prefix string, maxDepth int, all, dirsOnly bool) {
	if maxDepth == 0 {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var nodes []os.DirEntry
	for _, e := range entries {
		if !all && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		nodes = append(nodes, e)
	}
	for i, e := range nodes {
		last := i == len(nodes)-1
		branch := "├── "
		nextPrefix := prefix + "│   "
		if last {
			branch = "└── "
			nextPrefix = prefix + "    "
		}
		name := e.Name()
		if e.IsDir() {
			fmt.Printf("%s%s%s/\n", prefix, branch, name)
			printTree(filepath.Join(dir, name), nextPrefix, maxDepth-1, all, dirsOnly)
		} else if !dirsOnly {
			fmt.Printf("%s%s%s\n", prefix, branch, name)
		}
	}
}

func init() {
	treeCmd.Flags().IntP("depth", "L", 0, "最大递归深度（0 表示不限制）")
	treeCmd.Flags().BoolP("all", "a", false, "包含隐藏文件/目录")
	treeCmd.Flags().BoolP("dirs-only", "d", false, "只看目录")
	rootCmd.AddCommand(treeCmd)
}
