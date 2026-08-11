package store

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTempHome 把数据目录指到临时目录，避免污染真实 ~/.philokun，并在测试后清理。
func setupTempHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
}

func TestNoteCRUDAndUndo(t *testing.T) {
	setupTempHome(t)

	if err := AddNote("alpha"); err != nil {
		t.Fatalf("AddNote: %v", err)
	}
	if err := AddNote("beta"); err != nil {
		t.Fatalf("AddNote: %v", err)
	}
	if err := AddNote("gamma"); err != nil {
		t.Fatalf("AddNote: %v", err)
	}

	notes, err := ListNotes()
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) != 3 {
		t.Fatalf("期望 3 条，实际 %d", len(notes))
	}

	// 修改
	if _, err := UpdateNote(2, "BETA-EDITED"); err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}
	edited, err := GetNote(2)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if edited.Text != "BETA-EDITED" {
		t.Fatalf("修改未生效，得到 %q", edited.Text)
	}

	// 批量删除 1 与 3
	if _, err := DeleteNotes([]int{1, 3}); err != nil {
		t.Fatalf("DeleteNotes: %v", err)
	}
	notes, _ = ListNotes()
	if len(notes) != 1 || notes[0].ID != 2 {
		t.Fatalf("删除后期望只剩 #2，实际 %+v", notes)
	}

	// 删除不存在的 ID 应报错
	if _, err := DeleteNote(99); err == nil {
		t.Fatal("删除不存在的 ID 应返回错误")
	}

	// 撤销最近一次删除，恢复 1 与 3
	n, err := UndoDelete()
	if err != nil {
		t.Fatalf("UndoDelete: %v", err)
	}
	if n != 2 {
		t.Fatalf("期望恢复 2 条，实际 %d", n)
	}
	notes, _ = ListNotes()
	if len(notes) != 3 {
		t.Fatalf("撤销后期望 3 条，实际 %d", len(notes))
	}

	// 再删除一条后，undo 应报错（无删除记录）
	if _, err := DeleteNote(1); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}
	if _, err := UndoDelete(); err != nil {
		t.Fatalf("第一次 undo 应成功: %v", err)
	}
	if _, err := UndoDelete(); err == nil {
		t.Fatal("无删除记录时 undo 应返回错误")
	}
}

func TestNoteDeletedNotListed(t *testing.T) {
	setupTempHome(t)
	AddNote("x")
	AddNote("y")
	// 直接软删后，ListNotes/SearchNotes 不应包含它
	if _, err := DeleteNote(1); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}
	notes, _ := ListNotes()
	for _, n := range notes {
		if n.ID == 1 {
			t.Fatal("已删除笔记不应出现在 ListNotes")
		}
	}
	got, _ := GetNote(1)
	if got != nil {
		t.Fatal("已删除笔记不应被 GetNote 取到")
	}
}

// 确保数据确实写入了 ~/.philokun/notes.json（持久化）。
func TestNotePersistedToFile(t *testing.T) {
	setupTempHome(t)
	AddNote("persist-me")
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".philokun", notesStoreFile))
	if err != nil {
		t.Fatalf("读取数据文件失败: %v", err)
	}
	if !contains(string(data), "persist-me") {
		t.Fatalf("数据文件未包含笔记内容: %s", data)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
