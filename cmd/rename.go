package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// renameCmd 批量重命名文件（前缀 / 后缀 / 替换 / 序号）。
var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "批量重命名文件",
	Long: `在目录内批量重命名文件，支持四种模式（用 flag 指定）：

  --prefix <s>   给每个文件名加前缀
  --suffix <s>   给每个文件名加后缀（在扩展名前）
  --replace <a>  把文件名中的 a 替换为 b（配合 --with）
  --with <b>     替换目标串（与 --replace 搭配）
  --index <s>    按序号重命名：<s>1<s> <s>2<s> ...（s 为前缀）

目录通过 --dir 指定。默认只改普通文件；-r 递归子目录。
先 --dry-run 预览，确认后再执行。

示例:
  philokun rename --dir ./photos --prefix "trip-"
  philokun rename --dir ./docs --replace " " --with "_"
  philokun rename --dir ./data --index "img_" --dry-run
  philokun rename --dir ./data --suffix "_bak" -r`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		prefix, _ := cmd.Flags().GetString("prefix")
		suffix, _ := cmd.Flags().GetString("suffix")
		replace, _ := cmd.Flags().GetString("replace")
		with, _ := cmd.Flags().GetString("with")
		index, _ := cmd.Flags().GetString("index")
		recursive, _ := cmd.Flags().GetBool("recursive")
		dry, _ := cmd.Flags().GetBool("dry-run")

		if dir == "" {
			return fmt.Errorf("请用 --dir 指定目录")
		}
		modeCount := 0
		if prefix != "" {
			modeCount++
		}
		if suffix != "" {
			modeCount++
		}
		if replace != "" {
			modeCount++
		}
		if index != "" {
			modeCount++
		}
		if modeCount == 0 {
			return fmt.Errorf("至少指定一种重命名模式：--prefix / --suffix / --replace / --index")
		}
		if modeCount > 1 {
			return fmt.Errorf("一次只支持一种重命名模式")
		}
		if replace != "" && with == "" {
			return fmt.Errorf("--replace 需配合 --with 指定替换目标串")
		}

		files, err := collectFiles(dir, recursive)
		if err != nil {
			return err
		}
		seq := 0
		for _, path := range files {
			old := filepath.Base(path)
			var newName string
			switch {
			case prefix != "":
				newName = prefix + old
			case suffix != "":
				ext := filepath.Ext(old)
				base := strings.TrimSuffix(old, ext)
				newName = base + suffix + ext
			case replace != "":
				newName = strings.ReplaceAll(old, replace, with)
			case index != "":
				seq++
				newName = fmt.Sprintf("%s%d", index, seq)
				if ext := filepath.Ext(old); ext != "" {
					newName += ext
				}
			}
			if newName == old {
				continue
			}
			newPath := filepath.Join(filepath.Dir(path), newName)
			if dry {
				fmt.Printf("[预览] %s  ->  %s\n", path, newPath)
				continue
			}
			if _, err := os.Stat(newPath); err == nil {
				fmt.Printf("[跳过] 目标已存在: %s\n", newPath)
				continue
			}
			if err := os.Rename(path, newPath); err != nil {
				fmt.Fprintf(os.Stderr, "重命名失败 %s: %v\n", path, err)
				continue
			}
			fmt.Printf("%s  ->  %s\n", path, newPath)
		}
		return nil
	},
}

// collectFiles 收集目录（可选递归）下的普通文件。
func collectFiles(dir string, recursive bool) ([]string, error) {
	var out []string
	walk := func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p == dir {
				return nil
			}
			if !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		out = append(out, p)
		return nil
	}
	if err := filepath.WalkDir(dir, walk); err != nil {
		return nil, err
	}
	return out, nil
}

func init() {
	renameCmd.Flags().StringP("dir", "d", "", "目标目录（必填）")
	renameCmd.Flags().StringP("prefix", "p", "", "加前缀")
	renameCmd.Flags().StringP("suffix", "s", "", "加后缀（扩展名前）")
	renameCmd.Flags().String("replace", "", "替换源串（配合 --with）")
	renameCmd.Flags().String("with", "", "替换目标串（配合 --replace）")
	renameCmd.Flags().StringP("index", "i", "", "按序号重命名，参数为前缀")
	renameCmd.Flags().BoolP("recursive", "r", false, "递归子目录")
	renameCmd.Flags().Bool("dry-run", false, "仅预览，不真正重命名")
	rootCmd.AddCommand(renameCmd)
}
