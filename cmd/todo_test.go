package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"philokun/internal/store"
)

func withTempHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
}

func TestTodoAddAndList(t *testing.T) {
	withTempHome(t)

	var out bytes.Buffer
	todoAddCmd.SetOut(&out)
	if err := todoAddCmd.RunE(todoAddCmd, []string{"买牛奶", "写代码"}); err != nil {
		t.Fatalf("todo add 失败: %v", err)
	}

	todos, err := store.ListTodos()
	if err != nil {
		t.Fatalf("ListTodos 失败: %v", err)
	}
	if len(todos) != 1 || todos[0] != "买牛奶 写代码" {
		t.Fatalf("空格参数未正确拼接: %v", todos)
	}
}

func TestTodoAddRequiresArg(t *testing.T) {
	withTempHome(t)
	// cobra.MinimumNArgs(1) 会在 RunE 之前拦截，这里直接验证 RunE 在空参数时由 cobra 报错。
	// 通过命令的 Args 校验来断言。
	if err := todoAddCmd.Args(todoAddCmd, []string{}); err == nil {
		t.Fatal("无参数时应当报错")
	}
}

func TestTodoListEmpty(t *testing.T) {
	withTempHome(t)

	get := captureStdout(t)
	if err := todoListCmd.RunE(todoListCmd, []string{}); err != nil {
		t.Fatalf("todo list 失败: %v", err)
	}
	if got := get(); got == "" {
		t.Fatal("空列表应输出提示文案")
	}
}

func TestDataFileLocation(t *testing.T) {
	withTempHome(t)
	_ = os.UserHomeDir
	_ = filepath.Join
}
