package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/philokun/internal/store"
)

// noteCmd 是笔记分组命令，本身不做事，只是各子命令的父级。
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
		return printNoteList()
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
			fmt.Printf("%d. [%s] %s\n", n.ID, formatNoteTime(n.At), n.Text)
		}
		return nil
	},
}

// noteEditCmd 修改一条笔记的内容（标题与正文统一为整条文本）。
// 支持 -m/--message 直接给新内容（非交互）；否则交互读取新内容并提供保存/取消。
var noteEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "修改一条笔记的内容",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(strings.TrimSpace(args[0]))
		if err != nil {
			return fmt.Errorf("无效的笔记 ID: %q", args[0])
		}
		n, err := store.GetNote(id)
		if err != nil {
			return err
		}
		flagText, _ := cmd.Flags().GetString("message")
		var newText string
		if flagText != "" {
			newText = flagText
		} else {
			fmt.Printf("当前内容: %s\n", n.Text)
			input := readLine("请输入新内容（直接回车取消）: ")
			if input == "" {
				fmt.Println("已取消，未做任何修改。")
				return nil
			}
			newText = input
		}
		if newText == n.Text {
			fmt.Println("内容未变化，已取消。")
			return nil
		}
		// 仅在交互读取时需要保存确认；-m 显式给内容则直接保存。
		if flagText == "" {
			skip, _ := cmd.Flags().GetBool("yes")
			if !skip && !confirmPrompt(fmt.Sprintf("保存对笔记 #%d 的修改?", id)) {
				fmt.Println("已取消修改。")
				return nil
			}
		}
		updated, err := store.UpdateNote(id, newText)
		if err != nil {
			return err
		}
		fmt.Printf("已更新笔记 #%d。\n", updated.ID)
		fmt.Println("更新后的笔记列表：")
		return printNoteList()
	},
}

// noteRmCmd 删除笔记（单条或批量）。带确认提示，--yes 跳过确认。
var noteRmCmd = &cobra.Command{
	Use:     "rm <id>...",
	Aliases: []string{"del", "delete", "remove"},
	Short:   "删除笔记（支持多个 ID 批量删除）",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := make([]int, 0, len(args))
		for _, a := range args {
			id, err := strconv.Atoi(strings.TrimSpace(a))
			if err != nil {
				return fmt.Errorf("无效的笔记 ID: %q", a)
			}
			ids = append(ids, id)
		}
		// 先校验所有 ID 是否存在，给出明确错误。
		notes, err := store.ListNotes()
		if err != nil {
			return err
		}
		byID := make(map[int]store.Note, len(notes))
		for _, n := range notes {
			byID[n.ID] = n
		}
		var missing []int
		for _, id := range ids {
			if _, ok := byID[id]; !ok {
				missing = append(missing, id)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("以下笔记 ID 不存在: %v", missing)
		}
		// 确认提示。
		skip, _ := cmd.Flags().GetBool("yes")
		if !skip {
			var prompt string
			if len(ids) == 1 {
				n := byID[ids[0]]
				prompt = fmt.Sprintf("确认删除笔记 #%d %q?", ids[0], n.Text)
			} else {
				prompt = fmt.Sprintf("确认删除这 %d 条笔记?（可用 note undo 撤销）", len(ids))
			}
			if !confirmPrompt(prompt) {
				fmt.Println("已取消删除。")
				return nil
			}
		}
		deleted, err := store.DeleteNotes(ids)
		if err != nil {
			return err
		}
		fmt.Printf("已删除 %d 条笔记。\n", len(deleted))
		fmt.Println("更新后的笔记列表：")
		return printNoteList()
	},
}

// noteUndoCmd 撤销最近一次删除，恢复被软删的笔记。
var noteUndoCmd = &cobra.Command{
	Use:   "undo",
	Short: "撤销最近一次删除（恢复被删的笔记）",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := store.UndoDelete()
		if err != nil {
			return err
		}
		fmt.Printf("已撤销删除，恢复 %d 条笔记。\n", n)
		fmt.Println("更新后的笔记列表：")
		return printNoteList()
	},
}

// noteClearCmd 清空全部笔记，带确认提示，--yes 跳过确认。
var noteClearCmd = &cobra.Command{
	Use:     "clear",
	Aliases: []string{"empty", "clean"},
	Short:   "清空全部笔记",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		count, err := store.ListNotes()
		if err != nil {
			return err
		}
		if len(count) == 0 {
			fmt.Println("当前没有笔记，无需清空。")
			return nil
		}
		skip, _ := cmd.Flags().GetBool("yes")
		if !skip {
			if !confirmPrompt(fmt.Sprintf("确认清空全部 %d 条笔记？此操作可用 `philokun note undo` 撤销", len(count))) {
				fmt.Println("已取消清空。")
				return nil
			}
		}
		n, err := store.ClearNotes()
		if err != nil {
			return err
		}
		fmt.Printf("已清空 %d 条笔记。\n", n)
		fmt.Println("更新后的笔记列表：")
		return printNoteList()
	},
}

// notePurgeCmd 物理删除笔记（单条或批量），真正从存储中移除，不可撤销。
// 带确认提示，--yes 跳过确认。别名：erase / wipe / destroy。
var notePurgeCmd = &cobra.Command{
	Use:     "purge <id>...",
	Aliases: []string{"erase", "wipe", "destroy"},
	Short:   "物理删除笔记（真正移除，不可撤销）",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := make([]int, 0, len(args))
		for _, a := range args {
			id, err := strconv.Atoi(strings.TrimSpace(a))
			if err != nil {
				return fmt.Errorf("无效的笔记 ID: %q", a)
			}
			ids = append(ids, id)
		}
		notes, err := store.ListNotes()
		if err != nil {
			return err
		}
		byID := make(map[int]store.Note, len(notes))
		for _, n := range notes {
			byID[n.ID] = n
		}
		var missing []int
		for _, id := range ids {
			if _, ok := byID[id]; !ok {
				missing = append(missing, id)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("以下笔记 ID 不存在: %v", missing)
		}
		skip, _ := cmd.Flags().GetBool("yes")
		if !skip {
			var prompt string
			if len(ids) == 1 {
				n := byID[ids[0]]
				prompt = fmt.Sprintf("确认【物理删除】笔记 #%d %q？（此操作不可撤销）", ids[0], n.Text)
			} else {
				prompt = fmt.Sprintf("确认【物理删除】这 %d 条笔记？（此操作不可撤销）", len(ids))
			}
			if !confirmPrompt(prompt) {
				fmt.Println("已取消物理删除。")
				return nil
			}
		}
		deleted, err := store.PurgeNotes(ids)
		if err != nil {
			return err
		}
		fmt.Printf("已物理删除 %d 条笔记（不可撤销）。\n", len(deleted))
		fmt.Println("更新后的笔记列表：")
		return printNoteList()
	},
}

// notePurgeAllCmd 物理清空全部笔记，真正移除所有记录，不可撤销。带确认提示。
var notePurgeAllCmd = &cobra.Command{
	Use:     "purge-all",
	Aliases: []string{"erase-all", "wipe-all"},
	Short:   "物理清空全部笔记（真正移除，不可撤销）",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, err := store.ListNotes()
		if err != nil {
			return err
		}
		if len(notes) == 0 {
			fmt.Println("当前没有笔记，无需清空。")
			return nil
		}
		skip, _ := cmd.Flags().GetBool("yes")
		if !skip {
			if !confirmPrompt(fmt.Sprintf("确认【物理清空】全部 %d 条笔记？此操作不可撤销！", len(notes))) {
				fmt.Println("已取消物理清空。")
				return nil
			}
		}
		n, err := store.PurgeAllNotes()
		if err != nil {
			return err
		}
		fmt.Printf("已物理清空 %d 条笔记（不可撤销）。\n", n)
		fmt.Println("更新后的笔记列表：")
		return printNoteList()
	},
}

// printNoteList 打印当前全部未删除笔记，空时给出友好提示。
func printNoteList() error {
	notes, err := store.ListNotes()
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		fmt.Println("暂无笔记。用 `philokun note add 内容` 记下第一条吧。")
		return nil
	}
	for _, n := range notes {
		fmt.Printf("%d. [%s] %s\n", n.ID, formatNoteTime(n.At), n.Text)
	}
	return nil
}

// formatNoteTime 把存储中的时间字符串解析并格式化为“YYYY-MM-DD HH:mm:ss”。
// 若解析失败（例如旧数据格式异常），则原样返回，避免输出空白。
func formatNoteTime(at string) string {
	t, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return at
	}
	return t.Format("2006-01-02 15:04:05")
}

// stdinReader 是标准输入的唯一缓冲读取器，供交互式读取与确认共用，
// 避免多个 bufio.Reader 竞争同一 fd 导致数据错乱。
var stdinReader = bufio.NewReader(os.Stdin)

// confirmPrompt 从 stdin 读取 y/n 确认，默认否。
func confirmPrompt(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// readLine 从 stdin 读取一行（不含换行符）。
func readLine(prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	line, _ := stdinReader.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func init() {
	noteCmd.AddCommand(noteAddCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteSearchCmd)
	noteCmd.AddCommand(noteEditCmd)
	noteCmd.AddCommand(noteRmCmd)
	noteCmd.AddCommand(noteUndoCmd)
	noteCmd.AddCommand(noteClearCmd)
	noteCmd.AddCommand(notePurgeCmd)
	noteCmd.AddCommand(notePurgeAllCmd)

	noteEditCmd.Flags().StringP("message", "m", "", "直接指定新内容（非交互模式）")
	noteEditCmd.Flags().BoolP("yes", "y", false, "跳过保存确认直接保存")
	noteRmCmd.Flags().BoolP("yes", "y", false, "跳过删除确认")
	noteClearCmd.Flags().BoolP("yes", "y", false, "跳过清空确认")
	notePurgeCmd.Flags().BoolP("yes", "y", false, "跳过物理删除确认")
	notePurgeAllCmd.Flags().BoolP("yes", "y", false, "跳过物理清空确认")

	rootCmd.AddCommand(noteCmd)
}
