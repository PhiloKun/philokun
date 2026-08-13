package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// watchCmd 周期性重复执行命令或请求 URL。
var watchCmd = &cobra.Command{
	Use:   "watch <命令或URL>",
	Short: "周期性重复执行命令 / 请求 URL",
	Long: `每隔固定间隔重复执行一个 shell 命令，或请求一个 URL 打印状态。

默认间隔 2 秒；-n 指定秒数；-c 指定最大次数（0 表示无限）。
命令模式：整段作为 shell -c 执行；若以 http(s):// 开头则按 URL 处理。

示例:
  philokun watch -n 1 "date"
  philokun watch -n 5 -c 10 "ls -l"
  philokun watch -n 3 https://example.com/health`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		interval, _ := cmd.Flags().GetInt("interval")
		count, _ := cmd.Flags().GetInt("count")
		if interval < 1 {
			interval = 2
		}
		isURL := strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")

		round := 0
		for {
			round++
			fmt.Printf("\033[2J\033[H") // 清屏并回到左上角
			fmt.Printf("── 第 %d 次（%s）──\n", round, time.Now().Format("15:04:05"))
			var err error
			if isURL {
				err = doWatchRequest(target)
			} else {
				err = doWatchCommand(target)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "执行出错: %v\n", err)
			}
			if count > 0 && round >= count {
				return nil
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	},
}

// doWatchCommand 以 shell -c 执行命令并透传输出。
func doWatchCommand(command string) error {
	c := exec.Command("sh", "-c", command)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// doWatchRequest 请求 URL 并打印状态码与响应大小。
func doWatchRequest(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("HTTP %d  %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("响应大小: %d 字节\n", len(body))
	return nil
}

func init() {
	watchCmd.Flags().IntP("interval", "n", 2, "刷新间隔（秒）")
	watchCmd.Flags().IntP("count", "c", 0, "最大执行次数（0 表示无限）")
	rootCmd.AddCommand(watchCmd)
}
