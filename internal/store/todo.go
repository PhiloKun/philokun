// Package store 负责待办数据的持久化（存到本地 JSON 文件）。
// 把“数据怎么存”和“命令怎么跑”分开，是很好的工程习惯：
// 以后你想换成数据库或云同步，只需要改这个包，命令代码一行都不用动。
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// todoFile 是 JSON 文件在内存里的结构。
// 待办目前用最简单的字符串表示，以后可以扩展成带 ID / 完成状态的结构体。
type todoFile struct {
	Todos []string `json:"todos"`
}

// dataPath 返回数据文件完整路径：~/.philokun/todo.json
// 放在用户主目录下，不污染项目目录，重装命令也不会丢数据。

func dataPath() (string, error) {
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

// load 读取并解析 JSON 文件；文件不存在时返回空列表而不是报错。
func load() (todoFile, error) {
	p, err := dataPath()
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
	if err := json.Unmarshal(data, &tf); err != nil {
		return todoFile{}, err
	}
	return tf, nil
}

// save 把内存结构写回 JSON 文件（带缩进，方便人读）。
func save(tf todoFile) error {
	p, err := dataPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// AddTodo 在末尾追加一条待办并保存。
func AddTodo(text string) error {
	tf, err := load()
	if err != nil {
		return err
	}
	tf.Todos = append(tf.Todos, text)
	return save(tf)
}

// ListTodos 返回当前所有待办。
func ListTodos() ([]string, error) {
	tf, err := load()
	if err != nil {
		return nil, err
	}
	return tf.Todos, nil
}
