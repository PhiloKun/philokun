package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/philokun/internal/store"
)

// cryptCmd 文件级加解密（AES-256-GCM，复用保险箱的 scrypt 密钥派生）。
var cryptCmd = &cobra.Command{
	Use:   "crypt",
	Short: "文件加密 / 解密（AES-256-GCM）",
	Long: `对文件做对称加密 / 解密，使用 scrypt 派生密钥 + AES-256-GCM。
密文格式为 JSON 信封（含随机 salt/nonce），与保险箱同一套加密算法。

示例:
  philokun crypt enc secret.txt            # 加密，输出 secret.txt.crypt
  philokun crypt enc secret.txt -o x.bin   # 指定输出文件
  philokun crypt dec secret.txt.crypt      # 解密，输出 secret.txt`,
}

// cryptEncCmd 加密文件。
var cryptEncCmd = &cobra.Command{
	Use:   "enc <输入文件>",
	Short: "加密一个文件",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := args[0]
		password, err := readPassword("请输入加密密码（不会存储）: ")
		if err != nil {
			return err
		}
		if password == "" {
			return fmt.Errorf("密码不能为空")
		}
		data, err := os.ReadFile(in)
		if err != nil {
			return err
		}
		out, _ := cmd.Flags().GetString("output")
		if out == "" {
			out = in + ".crypt"
		}
		if err := store.EncryptFile(data, password, out); err != nil {
			return err
		}
		fmt.Printf("已加密 -> %s\n", out)
		return nil
	},
}

// cryptDecCmd 解密文件。
var cryptDecCmd = &cobra.Command{
	Use:   "dec <输入文件>",
	Short: "解密一个文件",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := args[0]
		password, err := readPassword("请输入解密密码: ")
		if err != nil {
			return err
		}
		plain, err := store.DecryptFile(in, password)
		if err != nil {
			return err
		}
		out, _ := cmd.Flags().GetString("output")
		if out == "" {
			out = stripCryptExt(in)
		}
		if err := os.WriteFile(out, plain, 0o600); err != nil {
			return err
		}
		fmt.Printf("已解密 -> %s\n", out)
		return nil
	},
}

// stripCryptExt 去掉 .crypt 后缀作为默认输出名。
func stripCryptExt(in string) string {
	if filepath.Ext(in) == ".crypt" {
		return in[:len(in)-len(".crypt")]
	}
	return in + ".dec"
}

func init() {
	cryptEncCmd.Flags().StringP("output", "o", "", "输出文件路径（默认 <输入>.crypt）")
	cryptDecCmd.Flags().StringP("output", "o", "", "输出文件路径（默认去掉 .crypt 后缀）")
	cryptCmd.AddCommand(cryptEncCmd)
	cryptCmd.AddCommand(cryptDecCmd)
	rootCmd.AddCommand(cryptCmd)
}
