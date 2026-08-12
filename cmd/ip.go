package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ipCmd 查询本机公网 IP（调用公共 API，纯标准库实现）。
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "查询本机公网 IP 地址",
	Long: `查询当前网络的公网 IPv4 地址。

依次尝试多个公共接口（ipify / ipapi / 淘宝），自动选用第一个可用的，仅读取 IP 文本。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := &http.Client{Timeout: 8 * time.Second}
		// 多个源：优先境外 ipify，失败回退国内可访问的源。
		sources := []string{
			"https://api.ipify.org?format=json",
			"https://ipapi.co/ip/",
			"https://api.ip.sb/ip",
			"http://ip.taobao.com/outGetIpInfo?ip=myip",
		}
		var lastErr error
		for _, u := range sources {
			resp, err := client.Get(u)
			if err != nil {
				lastErr = err
				continue
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = err
				continue
			}
			ip := extractIP(string(body))
			if ip != "" {
				fmt.Printf("公网 IP: %s\n", ip)
				return nil
			}
			lastErr = fmt.Errorf("响应中未找到 IP: %s", string(body))
		}
		return fmt.Errorf("查询公网 IP 失败: %w", lastErr)
	},
}

// extractIP 从多种返回格式中提取 IP（JSON 的 ip 字段 / 纯文本 / 淘宝嵌套 JSON）。
func extractIP(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// 纯文本（如 ipapi.co/ip 直接返回 IP）
	if net.ParseIP(s) != nil {
		return s
	}
	// 尝试解析 JSON 中的 ip 字段
	var r struct {
		IP  string `json:"ip"`
		Data struct {
			IP string `json:"ip"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(s), &r); err == nil {
		if r.IP != "" {
			return r.IP
		}
		if r.Data.IP != "" {
			return r.Data.IP
		}
	}
	return ""
}

func init() {
	rootCmd.AddCommand(ipCmd)
}
