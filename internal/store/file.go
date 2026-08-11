package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// errNotFound 构造一个“未找到”语义的错误，供各 store 复用。
func errNotFound(msg string) error {
	return errors.New(msg)
}

// dirPath 返回数据目录 ~/.philokun，并确保其存在。
func dirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".philokun")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// loadJSON 读取指定数据文件并解析到 v（须为指针）。
// 文件不存在时返回 false（notFound=true）且不报错，调用方据此使用零值。
// 这样可以复用同一套“文件不存在即空”的语义，避免每个 store 重复实现。
func loadJSON(filename string, v any) (notFound bool, err error) {
	dir, err := dirPath()
	if err != nil {
		return false, err
	}
	p := filepath.Join(dir, filename)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, err
	}
	return false, nil
}

// saveJSON 把 v 以缩进 JSON 形式写入指定数据文件。
func saveJSON(filename string, v any) error {
	dir, err := dirPath()
	if err != nil {
		return err
	}
	p := filepath.Join(dir, filename)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
