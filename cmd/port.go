package cmd

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// portCmd 检测本地端口是否被占用。
var portCmd = &cobra.Command{
	Use:   "port <端口> [端口...]",
	Short: "检测本地端口是否被占用",
	Long: `检测一个或多个本地 TCP 端口的占用情况。

支持同时检测多个端口（空格分隔），或一段区间（如 8080-8090）。
默认检测 127.0.0.1；可用 -H 指定其他地址。

示例:
  philokun port 8080
  philokun port 8080 9000 3000
  philokun port 8080-8090
  philokun port -H 0.0.0.0 5432`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		ports, err := expandPorts(args)
		if err != nil {
			return err
		}
		allFree := true
		for _, p := range ports {
			addr := net.JoinHostPort(host, strconv.Itoa(p))
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				fmt.Printf("端口 %d  → 已被占用\n", p)
				allFree = false
				continue
			}
			ln.Close()
			fmt.Printf("端口 %d  → 空闲\n", p)
		}
		if !allFree {
			return fmt.Errorf("存在被占用的端口")
		}
		return nil
	},
}

// expandPorts 把参数展开为端口整型列表，支持 a-b 区间。
func expandPorts(args []string) ([]int, error) {
	var out []int
	for _, a := range args {
		a = strings.TrimSpace(a)
		if strings.Contains(a, "-") {
			parts := strings.SplitN(a, "-", 2)
			lo, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("非法端口区间: %q", a)
			}
			hi, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("非法端口区间: %q", a)
			}
			if lo > hi {
				lo, hi = hi, lo
			}
			if hi > 65535 {
				return nil, fmt.Errorf("端口超出范围: %d", hi)
			}
			for p := lo; p <= hi; p++ {
				out = append(out, p)
			}
			continue
		}
		p, err := strconv.Atoi(a)
		if err != nil || p < 1 || p > 65535 {
			return nil, fmt.Errorf("非法端口: %q", a)
		}
		out = append(out, p)
	}
	return out, nil
}

func init() {
	portCmd.Flags().StringP("host", "H", "127.0.0.1", "检测的目标地址")
	rootCmd.AddCommand(portCmd)
}
