package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/philokun/internal/store"
)

// 口令生成可用字符集（去掉易混淆的 0/O、1/l/I）。
const pwCharset = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%^&*"

// passCmd 是密码工具分组命令。
var passCmd = &cobra.Command{
	Use:   "pass",
	Short: "本地密码生成与加密保险箱",
}

// passGenCmd 生成高强度随机口令。
var passGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "生成高强度随机口令",
	RunE: func(cmd *cobra.Command, args []string) error {
		length, _ := cmd.Flags().GetInt("length")
		if length <= 0 {
			return fmt.Errorf("长度必须为正整数")
		}
		pw, err := generatePassword(length)
		if err != nil {
			return err
		}
		fmt.Println(pw)
		return nil
	},
}

// passSetCmd 将账号口令加密存入保险箱。
var passSetCmd = &cobra.Command{
	Use:   "set <名称>",
	Short: "加密保存一条口令到本地保险箱",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		value, _ := cmd.Flags().GetString("value")
		if value == "" {
			v, err := readPassword("请输入要保存的口令: ")
			if err != nil {
				return err
			}
			value = v
		}
		password, err := readPassword("请输入保险箱主密码（用于加密，不会存储）: ")
		if err != nil {
			return err
		}
		if err := store.SetSecret(name, value, password); err != nil {
			return err
		}
		fmt.Printf("已加密保存 %q。\n", name)
		return nil
	},
}

// passGetCmd 解密取出口令。
var passGetCmd = &cobra.Command{
	Use:   "get <名称>",
	Short: "从保险箱解密取出口令",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		password, err := readPassword("请输入保险箱主密码: ")
		if err != nil {
			return err
		}
		value, err := store.GetSecret(name, password)
		if err != nil {
			return err
		}

		useJSON, _ := cmd.Flags().GetBool("json")
		if !useJSON {
			fmt.Println(value)
			return nil
		}

		// 结构化输出：密码明文、提取时间戳、保险箱标识符。
		out := struct {
			Password    string `json:"password"`
			RetrievedAt string `json:"retrieved_at"`
			Vault       string `json:"vault"`
		}{
			Password:    value,
			RetrievedAt: time.Now().UTC().Format(time.RFC3339),
			Vault:       store.VaultIdentifier(),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	},
}

// passRmCmd 从保险箱中删除指定名称的记录。
var passRmCmd = &cobra.Command{
	Use:   "rm <名称>",
	Short: "删除保险箱中的记录",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("确认要删除保险箱中的记录 %q 吗？此操作不可恢复 (y/N): ", name)
			var answer string
			if _, err := fmt.Scanln(&answer); err != nil && err.Error() != "unexpected newline" {
				return err
			}
			if answer != "y" && answer != "Y" {
				fmt.Println("已取消删除。")
				return nil
			}
		}
		ok, err := store.RmSecret(name)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Printf("保险箱中没有名为 %q 的记录，无需删除。\n", name)
			return nil
		}
		fmt.Printf("已删除保险箱中的记录 %q。\n", name)
		return nil
	},
}

// passListCmd 列出保险箱里所有记录的 ID 与名称（不解密内容）。
var passListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出保险箱中所有记录（含 ID）",
	RunE: func(cmd *cobra.Command, args []string) error {
		secrets, err := store.ListSecrets()
		if err != nil {
			return err
		}
		if len(secrets) == 0 {
			fmt.Println("保险箱为空。用 `philokun pass set <名称>` 添加第一条吧。")
			return nil
		}
		fmt.Println("保险箱中的记录：")
		for _, s := range secrets {
			fmt.Printf("%d. %s\n", s.ID, s.Name)
		}
		return nil
	},
}

// generatePassword 用 crypto/rand 生成指定长度的随机口令。
func generatePassword(length int) (string, error) {
	result := make([]byte, length)
	max := big.NewInt(int64(len(pwCharset)))
	for i := range result {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = pwCharset[idx.Int64()]
	}
	return string(result), nil
}

// readPassword 从终端读取密码且不回显；非终端环境（如管道）下降级为普通输入。
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	var s string
	_, err := fmt.Scanln(&s)
	return s, err
}

func init() {
	passGenCmd.Flags().IntP("length", "l", 16, "口令长度")
	passSetCmd.Flags().StringP("value", "v", "", "要保存的口令（省略则在终端输入）")

	passGetCmd.Flags().BoolP("json", "j", false, "以 JSON 格式输出（包含 password / retrieved_at / vault）")
	passRmCmd.Flags().BoolP("force", "f", false, "跳过确认直接删除")

	passCmd.AddCommand(passGenCmd)
	passCmd.AddCommand(passSetCmd)
	passCmd.AddCommand(passGetCmd)
	passCmd.AddCommand(passRmCmd)
	passCmd.AddCommand(passListCmd)
	rootCmd.AddCommand(passCmd)
}
