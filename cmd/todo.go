package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"philokun/internal/store"
)

// todoCmd 是“分组命令”：它本身不做事，只是 add / list 的父级。
var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "管理待办事项",
}

// todoAddCmd 演示：带参数 + 业务逻辑 + 错误处理。
// Args: cobra.MinimumNArgs(1) 表示至少需要 1 个参数，否则 Cobra 自动报错。
var todoAddCmd = &cobra.Command{
	Use:   "add [待办内容...]",
	Short: "添加一条待办",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 用户可能输入带空格的内容，用空格把多个参数拼回一句。
		text := strings.Join(args, " ")
		if err := store.AddTodo(text); err != nil {
			return err
		}
		fmt.Printf("已添加待办: %s\n", text)
		return nil
	},
}

// todoListCmd 演示：无参数命令 + 读数据并展示。
var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有待办",
	RunE: func(cmd *cobra.Command, args []string) error {
		todos, err := store.ListTodos()
		if err != nil {
			return err
		}
		if len(todos) == 0 {
			fmt.Println("暂无待办。用 `philokun todo add 内容` 添加第一条吧。")
			return nil
		}
		for _, t := range todos {
			mark := "[ ]"
			if t.Done {
				mark = "[x]"
			}
			fmt.Printf("%d. %s %s\n", t.ID, mark, t.Text)
		}
		return nil
	},
}

// todoDoneCmd 把指定 ID 的待办标记为完成。
var todoDoneCmd = &cobra.Command{
	Use:   "done <ID>",
	Short: "标记待办为已完成",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ID 必须是数字: %w", err)
		}
		if err := store.DoneTodo(id); err != nil {
			return err
		}
		fmt.Printf("已将 ID %d 标记为完成。\n", id)
		return nil
	},
}

// todoRmCmd 删除指定 ID 的待办。
var todoRmCmd = &cobra.Command{
	Use:   "rm <ID>",
	Short: "删除一条待办",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ID 必须是数字: %w", err)
		}
		if err := store.RmTodo(id); err != nil {
			return err
		}
		fmt.Printf("已删除 ID %d。\n", id)
		return nil
	},
}

func init() {
	// 子命令先挂到 todoCmd，再把 todoCmd 挂到 rootCmd。
	todoCmd.AddCommand(todoAddCmd)
	todoCmd.AddCommand(todoListCmd)
	todoCmd.AddCommand(todoDoneCmd)
	todoCmd.AddCommand(todoRmCmd)
	rootCmd.AddCommand(todoCmd)
}
