package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// captureStdout 临时把 os.Stdout 重定向到缓冲区，返回读取函数与清理函数。
func captureStdout(t *testing.T) func() string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	return func() string {
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = old
		return buf.String()
	}
}

func TestVersionCommand(t *testing.T) {
	// version 命令使用 fmt.Printf 直接打到 stdout，这里捕获 stdout 验证文案。
	get := captureStdout(t)
	versionCmd.Run(versionCmd, []string{})
	got := get()

	want := "philokun version " + version + "\n"
	if got != want {
		t.Fatalf("version 输出不符: 期望 %q，实际 %q", want, got)
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"todo", "version", "qrcode"} {
		if !names[want] {
			t.Fatalf("根命令缺少子命令: %s", want)
		}
	}
}
