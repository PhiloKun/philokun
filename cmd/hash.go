package cmd

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// hashCmd 计算文本或文件的哈希值。
var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "计算文本 / 文件的哈希（md5 / sha1 / sha256 / sha512）",
	Long: `计算字符串或文件的哈希值，支持 md5 / sha1 / sha256 / sha512。

默认对字符串求值；指定 -f 后对文件求值。可用 -a 选择算法（可多次计算）。

示例:
  philokun hash hello
  philokun hash -a sha256 -a md5 hello
  philokun hash -f -a sha256 ./file.txt`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		algs, _ := cmd.Flags().GetStringSlice("algo")
		if len(algs) == 0 {
			algs = []string{"sha256"}
		}
		fileMode, _ := cmd.Flags().GetBool("file")

		if fileMode {
			if len(args) != 1 {
				return fmt.Errorf("文件模式下需提供 exactly 一个文件路径")
			}
			return hashFile(args[0], algs)
		}

		text := ""
		if len(args) == 1 {
			text = args[0]
		} else {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			text = string(data)
		}
		for _, a := range algs {
			h, err := newHasher(a)
			if err != nil {
				return err
			}
			h.Write([]byte(text))
			fmt.Printf("%-8s %s\n", strings.ToUpper(a), hex.EncodeToString(h.Sum(nil)))
		}
		return nil
	},
}

// hashFile 计算文件的多种哈希并输出。
func hashFile(path string, algs []string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, a := range algs {
		h, err := newHasher(a)
		if err != nil {
			return err
		}
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}
		fmt.Printf("%-8s %s  %s\n", strings.ToUpper(a), hex.EncodeToString(h.Sum(nil)), path)
	}
	return nil
}

// newHasher 根据算法名返回对应 hash.Hash。
func newHasher(name string) (hash.Hash, error) {
	switch strings.ToLower(name) {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("不支持的算法: %s（支持 md5 / sha1 / sha256 / sha512）", name)
	}
}

func init() {
	hashCmd.Flags().StringSliceP("algo", "a", []string{"sha256"}, "算法: md5 / sha1 / sha256 / sha512（可多次）")
	hashCmd.Flags().BoolP("file", "f", false, "对文件求值（参数为文件路径）")
	rootCmd.AddCommand(hashCmd)
}
