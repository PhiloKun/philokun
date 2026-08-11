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

// 清空后列表应返回 0 条，undo 可完整恢复。
func TestNoteClearAndUndo(t *testing.T) {
	setupTempHome(t)

	AddNote("one")
	AddNote("two")
	AddNote("three")

	n, err := ClearNotes()
	if err != nil {
		t.Fatalf("ClearNotes: %v", err)
	}
	if n != 3 {
		t.Fatalf("期望清空 3 条，实际 %d", n)
	}
	notes, _ := ListNotes()
	if len(notes) != 0 {
		t.Fatalf("清空后期望 0 条，实际 %d", len(notes))
	}

	// 二次清空应返回 0
	if n2, _ := ClearNotes(); n2 != 0 {
		t.Fatalf("二次清空应返回 0，实际 %d", n2)
	}

	// undo 应完整恢复 3 条
	restored, err := UndoDelete()
	if err != nil {
		t.Fatalf("UndoDelete: %v", err)
	}
	if restored != 3 {
		t.Fatalf("期望恢复 3 条，实际 %d", restored)
	}
	notes, _ = ListNotes()
	for _, n := range notes {
		if n.Deleted {
			t.Fatal("恢复后仍存在已删除标记")
		}
	}
	if len(notes) != 3 {
		t.Fatalf("恢复后期望 3 条，实际 %d", len(notes))
	}
}

// 删除后新增笔记应从 1 开始回收编号，undo 时若冲突则自动换号。
func TestNoteIDRecycling(t *testing.T) {
	setupTempHome(t)

	AddNote("first")  // 1
	AddNote("second") // 2
	AddNote("third")  // 3

	DeleteNotes([]int{1, 2})

	// 删除 1、2 后，新增笔记应占用最小编号 1
	if err := AddNote("after-delete"); err != nil {
		t.Fatalf("AddNote: %v", err)
	}
	notes, _ := ListNotes()
	if len(notes) != 2 {
		t.Fatalf("期望 2 条，实际 %d", len(notes))
	}
	if notes[0].ID != 1 || notes[0].Text != "after-delete" {
		t.Fatalf("新笔记应分配编号 1 且排在首位，实际 %+v", notes)
	}

	// undo 恢复 1、2；原 #1 与新笔记 #1 冲突，应被重新分配，最终无重复 ID。
	n, err := UndoDelete()
	if err != nil {
		t.Fatalf("UndoDelete: %v", err)
	}
	if n != 2 {
		t.Fatalf("期望恢复 2 条，实际 %d", n)
	}
	notes, _ = ListNotes()
	if len(notes) != 4 {
		t.Fatalf("undo 后期望 4 条，实际 %d", len(notes))
	}
	idSet := make(map[int]bool)
	for _, n := range notes {
		if idSet[n.ID] {
			t.Fatalf("存在重复 ID: %d", n.ID)
		}
		idSet[n.ID] = true
	}
	expected := map[int]bool{1: true, 2: true, 3: true, 4: true}
	if len(idSet) != len(expected) {
		t.Fatalf("ID 集合不正确: %v", idSet)
	}
	for id := range expected {
		if !idSet[id] {
			t.Fatalf("缺少期望 ID %d，实际 %v", id, idSet)
		}
	}
}

// 物理删除应从存储中真正移除记录，且不可被 undo 恢复。
func TestNotePurge(t *testing.T) {
	setupTempHome(t)

	AddNote("a") // 1
	AddNote("b") // 2
	AddNote("c") // 3

	// 物理删除 #2
	n, err := PurgeNote(2)
	if err != nil {
		t.Fatalf("PurgeNote: %v", err)
	}
	if n.Text != "b" {
		t.Fatalf("物理删除返回内容错误: %q", n.Text)
	}
	notes, _ := ListNotes()
	if len(notes) != 2 {
		t.Fatalf("期望 2 条，实际 %d", len(notes))
	}
	for _, x := range notes {
		if x.ID == 2 {
			t.Fatal("物理删除后 #2 不应再出现")
		}
	}

	// 物理删除不存在的 ID 应报错
	if _, err := PurgeNote(99); err == nil {
		t.Fatal("物理删除不存在的 ID 应返回错误")
	}

	// 批量物理删除 1 与 3
	if _, err := PurgeNotes([]int{1, 3}); err != nil {
		t.Fatalf("PurgeNotes: %v", err)
	}
	notes, _ = ListNotes()
	if len(notes) != 0 {
		t.Fatalf("批量物理删除后期望 0 条，实际 %d", len(notes))
	}

	// 物理删除不可被 undo 恢复（last_deleted 为空）
	if _, err := UndoDelete(); err == nil {
		t.Fatal("物理删除后 undo 应返回错误（无软删记录）")
	}
}

// 物理清空全部后文件应为空，且不可撤销。
func TestNotePurgeAll(t *testing.T) {
	setupTempHome(t)
	AddNote("x")
	AddNote("y")
	AddNote("z")

	removed, err := PurgeAllNotes()
	if err != nil {
		t.Fatalf("PurgeAllNotes: %v", err)
	}
	if removed != 3 {
		t.Fatalf("期望移除 3 条，实际 %d", removed)
	}
	notes, _ := ListNotes()
	if len(notes) != 0 {
		t.Fatalf("物理清空后期望 0 条，实际 %d", len(notes))
	}
	if _, err := UndoDelete(); err == nil {
		t.Fatal("物理清空后 undo 应返回错误")
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
