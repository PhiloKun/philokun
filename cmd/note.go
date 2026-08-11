package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"philokun/internal/store"
)

// noteCmd 是笔记分组命令，本身不做事，只是 add / list / search 的父级。
var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "终端速记本：随手记闪念笔记",
}

// noteAddCmd 追加一条笔记。
var noteAddCmd = &cobra.Command{
	Use:   "add [内容...]",
	Short: "记一条闪念笔记",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		if err := store.AddNote(text); err != nil {
			return err
		}
		fmt.Printf("已记录笔记: %s\n", text)
		return nil
	},
}

// noteListCmd 列出全部笔记。
var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出全部笔记",
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, err := store.ListNotes()
		if err != nil {
			return err
		}
		if len(notes) == 0 {
			fmt.Println("暂无笔记。用 `philokun note add 内容` 记下第一条吧。")
			return nil
		}
		for _, n := range notes {
			fmt.Printf("%d. [%s] %s\n", n.ID, n.At, n.Text)
		}
		return nil
	},
}

// noteSearchCmd 按关键词本地全文检索。
var noteSearchCmd = &cobra.Command{
	Use:   "search <关键词>",
	Short: "按关键词搜索笔记",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := strings.Join(args, " ")
		notes, err := store.SearchNotes(keyword)
		if err != nil {
			return err
		}
		if len(notes) == 0 {
			fmt.Printf("没有匹配 %q 的笔记。\n", keyword)
			return nil
		}
		fmt.Printf("找到 %d 条匹配 %q 的笔记：\n", len(notes), keyword)
		for _, n := range notes {
			fmt.Printf("%d. [%s] %s\n", n.ID, n.At, n.Text)
		}
		return nil
	},
}

func init() {
	noteCmd.AddCommand(noteAddCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteSearchCmd)
	rootCmd.AddCommand(noteCmd)
}
