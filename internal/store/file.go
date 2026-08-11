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

// saveJSON 把 v 以缩进 JSON 形式原子写入指定数据文件。
// 实现：先写临时文件，再 os.Rename 覆盖目标。Rename 在同级目录下是原子操作，
// 保证“要么写入成功、要么旧文件不变”，避免写一半崩溃导致数据损坏（事务性保证）。
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
	tmp, err := os.CreateTemp(dir, ".tmp-"+filename+"-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// 任何失败路径都清理临时文件，避免遗留垃圾文件。
	defer func() {
		_ = os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, p)
}
