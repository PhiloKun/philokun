package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// zipCmd 压缩 / 解压 zip 归档（基于标准库）。
var zipCmd = &cobra.Command{
	Use:   "zip",
	Short: "压缩 / 解压 zip 归档",
	Long: `把文件或目录打包成 zip，或从 zip 解压到目录。

示例:
  philokun zip c out.zip a.txt b.txt dir/   # 压缩
  philokun zip x archive.zip                # 解压到当前目录
  philokun zip x archive.zip -d target/     # 解压到指定目录`,
}

// zipCreateCmd 创建 zip 归档。
var zipCreateCmd = &cobra.Command{
	Use:   "c <输出.zip> <文件/目录...>",
	Short: "创建 zip 归档",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := args[0]
		zf, err := os.Create(out)
		if err != nil {
			return err
		}
		defer zf.Close()
		zw := zip.NewWriter(zf)
		defer zw.Close()

		for _, src := range args[1:] {
			if err := addToZip(zw, src); err != nil {
				return err
			}
		}
		fmt.Printf("已创建 %s\n", out)
		return nil
	},
}

// zipExtractCmd 从 zip 解压。
var zipExtractCmd = &cobra.Command{
	Use:   "x <归档.zip> [-d 目录]",
	Short: "解压 zip 归档",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dest, _ := cmd.Flags().GetString("dir")
		if dest == "" {
			dest = "."
		}
		if err := unzip(args[0], dest); err != nil {
			return err
		}
		fmt.Printf("已解压到 %s\n", dest)
		return nil
	},
}

// addToZip 把单个文件或目录加入 zip writer。
func addToZip(zw *zip.Writer, src string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return filepath.Walk(src, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return nil
			}
			return writeZipFile(zw, src, p)
		})
	}
	return writeZipFile(zw, filepath.Dir(src), src)
}

// writeZipFile 写入单个文件到 zip（去掉基目录前缀）。
func writeZipFile(zw *zip.Writer, base, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	rel, err := filepath.Rel(base, path)
	if err != nil {
		rel = filepath.Base(path)
	}
	w, err := zw.Create(rel)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

// unzip 把归档解压到 dest 目录，并防止路径穿越。
func unzip(archive, dest string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	for _, zf := range zr.File {
		target := filepath.Join(dest, zf.Name)
		// 防路径穿越：清洗后必须仍在 dest 内
		clean := filepath.Clean(target)
		if !strings.HasPrefix(clean, filepath.Clean(dest)) {
			return fmt.Errorf("非法路径（拒绝解压）: %s", zf.Name)
		}
		if zf.FileInfo().IsDir() {
			if err := os.MkdirAll(clean, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
			return err
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(clean, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return err
		}
		out.Close()
		rc.Close()
	}
	return nil
}

func init() {
	zipExtractCmd.Flags().StringP("dir", "d", "", "解压目标目录（默认当前目录）")
	zipCmd.AddCommand(zipCreateCmd)
	zipCmd.AddCommand(zipExtractCmd)
	rootCmd.AddCommand(zipCmd)
}
