package store

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempHome 把 store 的数据目录临时指向 t.TempDir()，避免污染 ~/.philokun。
// 通过覆盖 dataPath 依赖的 home 目录实现（dataPath 读取 os.UserHomeDir）。
// 这里直接把 HOME 环境变量指向临时目录。
func withTempHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp) // Windows 兼容
	_ = filepath.Join // 确保 import 被使用
}

func TestAddAndListTodos(t *testing.T) {
	withTempHome(t)

	if err := AddTodo("写测试"); err != nil {
		t.Fatalf("AddTodo 失败: %v", err)
	}
	if err := AddTodo("补文档"); err != nil {
		t.Fatalf("AddTodo 失败: %v", err)
	}

	todos, err := ListTodos()
	if err != nil {
		t.Fatalf("ListTodos 失败: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("期望 2 条待办，实际 %d 条: %v", len(todos), todos)
	}
	if todos[0] != "写测试" || todos[1] != "补文档" {
		t.Fatalf("待办顺序或内容不符: %v", todos)
	}
}

func TestListTodosEmpty(t *testing.T) {
	withTempHome(t)

	todos, err := ListTodos()
	if err != nil {
		t.Fatalf("空列表 ListTodos 不应报错: %v", err)
	}
	if len(todos) != 0 {
		t.Fatalf("期望空列表，实际: %v", todos)
	}
}

func TestDataFileCreated(t *testing.T) {
	withTempHome(t)

	if err := AddTodo("检查文件是否生成"); err != nil {
		t.Fatalf("AddTodo 失败: %v", err)
	}

	home, _ := os.UserHomeDir()
	p := filepath.Join(home, ".philokun", "todo.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("数据文件未创建: %v", err)
	}
}
