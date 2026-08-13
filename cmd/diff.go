package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// diffCmd 对比两个文件（或两段文本）的行级差异。
var diffCmd = &cobra.Command{
	Use:   "diff <文件A> <文件B>",
	Short: "对比两个文件的行级差异",
	Long: `逐行比较两个文件，输出类似 unified diff 的增删标记。

标记说明：
  +  仅在 B 中出现的行（新增）
  -  仅在 A 中出现的行（删除）
  空格 两文件都有（上下文）

示例:
  philokun diff a.txt b.txt
  philokun diff -u a.txt b.txt    # 紧凑模式（仅显示差异行）`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		onlyDiff, _ := cmd.Flags().GetBool("unified")
		a, err := readLines(args[0])
		if err != nil {
			return err
		}
		b, err := readLines(args[1])
		if err != nil {
			return err
		}
		added, removed, changed := diffLines(a, b)
		for _, d := range diffLinesDetailed(a, b) {
			if onlyDiff && d.kind == " " {
				continue
			}
			fmt.Printf("%s%s\n", d.kind, d.text)
		}
		fmt.Printf("\n摘要: 新增 %d 行，删除 %d 行，相同 %d 行\n", added, removed, changed)
		return nil
	},
}

// lineDiff 单项：类型 + 文本。
type lineDiff struct {
	kind string // " " / "+" / "-"
	text string
}

// readLines 读取文件为行切片（去末尾换行）。
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// diffLinesDetailed 用 LCS 求逐行差异序列。
func diffLinesDetailed(a, b []string) []lineDiff {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var out []lineDiff
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			out = append(out, lineDiff{" ", a[i]})
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			out = append(out, lineDiff{"-", a[i]})
			i++
		} else {
			out = append(out, lineDiff{"+", b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, lineDiff{"-", a[i]})
	}
	for ; j < m; j++ {
		out = append(out, lineDiff{"+", b[j]})
	}
	return out
}

// diffLines 统计增删同数量（供摘要）。
func diffLines(a, b []string) (added, removed, same int) {
	for _, d := range diffLinesDetailed(a, b) {
		switch d.kind {
		case "+":
			added++
		case "-":
			removed++
		case " ":
			same++
		}
	}
	return
}

func init() {
	diffCmd.Flags().BoolP("unified", "u", false, "紧凑模式：仅显示差异行")
	rootCmd.AddCommand(diffCmd)
}
