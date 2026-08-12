package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
)

// dnsCmd 解析域名的各种记录（A / AAAA / CNAME / MX / NS）。
var dnsCmd = &cobra.Command{
	Use:   "dns <域名>",
	Short: "DNS 解析（A / AAAA / CNAME / MX / NS）",
	Long: `查询域名的 DNS 记录，默认返回 A + AAAA，可用 -t 指定类型。

示例:
  philokun dns example.com
  philokun dns example.com -t mx
  philokun dns example.com -t all`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		typ, _ := cmd.Flags().GetString("type")
		typ = strings.ToLower(typ)

		switch typ {
		case "a":
			printIPs(domain, "A", net.LookupIP)
		case "aaaa":
			printIPs(domain, "AAAA", net.LookupIP)
		case "cname":
			cname, err := net.LookupCNAME(domain)
			if err != nil {
				return err
			}
			fmt.Printf("CNAME  %s\n", cname)
		case "mx":
			mxs, err := net.LookupMX(domain)
			if err != nil {
				return err
			}
			for _, mx := range mxs {
				fmt.Printf("MX  %d  %s\n", mx.Pref, mx.Host)
			}
		case "ns":
			nss, err := net.LookupNS(domain)
			if err != nil {
				return err
			}
			for _, ns := range nss {
				fmt.Printf("NS  %s\n", ns.Host)
			}
		case "all":
			for _, t := range []string{"a", "aaaa", "cname", "mx", "ns"} {
				if err := runDNSQuery(domain, t); err != nil {
					fmt.Printf("（%s 查询失败: %v）\n", strings.ToUpper(t), err)
				}
			}
		default:
			// 默认 A + AAAA
			ips, err := net.LookupIP(domain)
			if err != nil {
				return err
			}
			for _, ip := range ips {
				if v4 := ip.To4(); v4 != nil {
					fmt.Printf("A    %s\n", v4)
				} else {
					fmt.Printf("AAAA %s\n", ip)
				}
			}
		}
		return nil
	},
}

// printIPs 输出某类型的 IP 解析结果（A/AAAA 共用 LookupIP）。
func printIPs(domain, label string, lookup func(string) ([]net.IP, error)) {
	ips, err := lookup(domain)
	if err != nil {
		fmt.Printf("（%s 查询失败: %v）\n", label, err)
		return
	}
	for _, ip := range ips {
		if label == "A" && ip.To4() == nil {
			continue
		}
		if label == "AAAA" && ip.To4() != nil {
			continue
		}
		fmt.Printf("%s    %s\n", label, ip.String())
	}
}

// runDNSQuery 给 -t all 复用单个类型查询。
func runDNSQuery(domain, typ string) error {
	switch typ {
	case "a", "aaaa":
		printIPs(domain, strings.ToUpper(typ), net.LookupIP)
	case "cname":
		cname, err := net.LookupCNAME(domain)
		if err != nil {
			return err
		}
		fmt.Printf("CNAME  %s\n", cname)
	case "mx":
		mxs, err := net.LookupMX(domain)
		if err != nil {
			return err
		}
		for _, mx := range mxs {
			fmt.Printf("MX  %d  %s\n", mx.Pref, mx.Host)
		}
	case "ns":
		nss, err := net.LookupNS(domain)
		if err != nil {
			return err
		}
		for _, ns := range nss {
			fmt.Printf("NS  %s\n", ns.Host)
		}
	}
	return nil
}

func init() {
	dnsCmd.Flags().StringP("type", "t", "default", "记录类型: a / aaaa / cname / mx / ns / all（默认 A+AAAA）")
	rootCmd.AddCommand(dnsCmd)
}
