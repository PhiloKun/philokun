// Package store 负责待办数据的持久化（存到本地 JSON 文件）。
// 把“数据怎么存”和“命令怎么跑”分开，是很好的工程习惯：
// 以后你想换成数据库或云同步，只需要改这个包，命令代码一行都不用动。
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Todo 是单条待办的结构体，带唯一 ID、内容与完成状态。
type Todo struct {
	ID   int    `json:"id"`   // 列表中的序号，从 1 开始
	Text string `json:"text"` // 待办内容
	Done bool   `json:"done"` // 是否已完成
}

// todoFile 是 JSON 文件在内存里的结构。
type todoFile struct {
	Seq   int    `json:"seq"`   // 自增 ID 计数器，保证 ID 不重复
	Todos []Todo `json:"todos"` // 所有待办
}

// todoPath 返回数据文件完整路径：~/.philokun/todo.json
// 放在用户主目录下，不污染项目目录，重装命令也不会丢数据。
func todoPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".philokun")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "todo.json"), nil
}

// loadTodos 读取并解析 JSON 文件；文件不存在时返回空列表而不是报错。
// 如果文件是旧的字符串数组格式，会自动迁移为新的 Todo 对象格式；
// 如果是无法识别的损坏格式，则备份原文件后重置为空列表，避免程序崩溃。
func loadTodos() (todoFile, error) {
	p, err := todoPath()
	if err != nil {
		return todoFile{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return todoFile{}, nil
		}
		return todoFile{}, err
	}

	var tf todoFile
	if err := json.Unmarshal(data, &tf); err == nil {
		return tf, nil
	}

	// 兼容旧格式/损坏文件：todos 是字符串数组（如 ["测试"]）。
	// 注意：前面失败的 json.Unmarshal 可能已向 tf.Todos 写入零值，
	// 所以这里使用全新的变量，避免残留脏数据。
	var legacy struct {
		Seq   int      `json:"seq"`
		Todos []string `json:"todos"`
	}
	if err := json.Unmarshal(data, &legacy); err == nil && len(legacy.Todos) > 0 {
		var migrated todoFile
		migrated.Seq = max(legacy.Seq, len(legacy.Todos))
		for i, text := range legacy.Todos {
			migrated.Todos = append(migrated.Todos, Todo{ID: i + 1, Text: text})
		}
		return migrated, nil
	}

	// 无法识别的格式：备份后重置，避免后续操作继续失败。
	backup := p + ".bak." + strconv.FormatInt(time.Now().Unix(), 10)
	_ = os.Rename(p, backup)
	return todoFile{}, nil
}

// saveTodos 把内存结构写回 JSON 文件（带缩进，方便人读）。
func saveTodos(tf todoFile) error {
	p, err := todoPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// AddTodo 在末尾追加一条待办并保存，自动分配自增 ID。
func AddTodo(text string) error {
	tf, err := loadTodos()
	if err != nil {
		return err
	}
	tf.Seq++
	tf.Todos = append(tf.Todos, Todo{ID: tf.Seq, Text: text})
	return saveTodos(tf)
}

// ListTodos 返回当前所有待办。
func ListTodos() ([]Todo, error) {
	tf, err := loadTodos()
	if err != nil {
		return nil, err
	}
	return tf.Todos, nil
}

// findTodo 按 ID 查找待办，返回其在切片中的下标。
func findTodo(todos []Todo, id int) (int, bool) {
	for i, t := range todos {
		if t.ID == id {
			return i, true
		}
	}
	return -1, false
}

// DoneTodo 把指定 ID 的待办标记为已完成。
func DoneTodo(id int) error {
	tf, err := loadTodos()
	if err != nil {
		return err
	}
	idx, ok := findTodo(tf.Todos, id)
	if !ok {
		return fmt.Errorf("未找到 ID 为 %d 的待办", id)
	}
	tf.Todos[idx].Done = true
	return saveTodos(tf)
}

// UndoTodo 把指定 ID 的已完成待办退回为未完成（撤销完成）。
func UndoTodo(id int) error {
	tf, err := loadTodos()
	if err != nil {
		return err
	}
	idx, ok := findTodo(tf.Todos, id)
	if !ok {
		return fmt.Errorf("未找到 ID 为 %d 的待办", id)
	}
	if !tf.Todos[idx].Done {
		return fmt.Errorf("ID 为 %d 的待办当前就是未完成状态，无需撤销", id)
	}
	tf.Todos[idx].Done = false
	return saveTodos(tf)
}

// RmTodo 删除指定 ID 的待办。
func RmTodo(id int) error {
	tf, err := loadTodos()
	if err != nil {
		return err
	}
	idx, ok := findTodo(tf.Todos, id)
	if !ok {
		return fmt.Errorf("未找到 ID 为 %d 的待办", id)
	}
	tf.Todos = append(tf.Todos[:idx], tf.Todos[idx+1:]...)
	return saveTodos(tf)
}
