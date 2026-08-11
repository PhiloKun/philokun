package store

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// noteFile 是笔记 JSON 文件的内存结构。
type noteFile struct {
	Seq         int    `json:"seq"`                   // 已废弃：保留以保证旧数据兼容，新增笔记不再使用
	Notes       []Note `json:"notes"`                 // 所有笔记（含已软删）
	LastDeleted []int  `json:"last_deleted,omitempty"` // 最近一次删除批次的 ID，供 undo 恢复
}

// Note 是单条闪念笔记。
type Note struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	At        string `json:"at"`                            // 创建时间（RFC3339）
	Deleted   bool   `json:"deleted,omitempty"`             // 是否已软删除
	DeletedAt string `json:"deleted_at,omitempty"`          // 软删除时间
}

const notesStoreFile = "notes.json"

// loadNotes 读取笔记文件到内存；文件不存在时返回空结构。
func loadNotes() (*noteFile, error) {
	var nf noteFile
	if _, err := loadJSON(notesStoreFile, &nf); err != nil {
		return nil, err
	}
	return &nf, nil
}

// saveNotes 持久化笔记文件。
func saveNotes(nf *noteFile) error {
	return saveJSON(notesStoreFile, nf)
}

// AddNote 追加一条笔记并保存，自动记录时间并分配最小可用 ID。
// 删除笔记后腾出的编号会被回收，保证用户看到的编号从 1 开始连续。
func AddNote(text string) error {
	nf, err := loadNotes()
	if err != nil {
		return err
	}
	nf.Notes = append(nf.Notes, Note{
		ID:   nextNoteID(nf.Notes),
		Text: text,
		At:   time.Now().Format(time.RFC3339),
	})
	return saveNotes(nf)
}

// nextNoteID 返回当前未删除笔记中尚未使用的最小正整数 ID。
func nextNoteID(notes []Note) int {
	used := make(map[int]bool)
	for _, n := range notes {
		if !n.Deleted {
			used[n.ID] = true
		}
	}
	id := 1
	for used[id] {
		id++
	}
	return id
}

// isActiveID 判断当前是否存在未删除且使用指定 ID 的笔记。
func isActiveID(notes []Note, id int) bool {
	for _, n := range notes {
		if n.ID == id && !n.Deleted {
			return true
		}
	}
	return false
}

// ListNotes 返回全部未删除的笔记（按 ID 升序排列）。
func ListNotes() ([]Note, error) {
	nf, err := loadNotes()
	if err != nil {
		return nil, err
	}
	var out []Note
	for _, n := range nf.Notes {
		if !n.Deleted {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// SearchNotes 按关键词做本地不区分大小写全文检索。
func SearchNotes(keyword string) ([]Note, error) {
	notes, err := ListNotes()
	if err != nil {
		return nil, err
	}
	k := strings.ToLower(keyword)
	var out []Note
	for _, n := range notes {
		if strings.Contains(strings.ToLower(n.Text), k) {
			out = append(out, n)
		}
	}
	return out, nil
}

// GetNote 按 ID 取一条未删除的笔记。
func GetNote(id int) (*Note, error) {
	nf, err := loadNotes()
	if err != nil {
		return nil, err
	}
	for i := range nf.Notes {
		if nf.Notes[i].ID == id && !nf.Notes[i].Deleted {
			return &nf.Notes[i], nil
		}
	}
	return nil, errNotFound(fmt.Sprintf("笔记 #%d 不存在", id))
}

// DeleteNote 软删除单条笔记，并记录到最近删除批次。
func DeleteNote(id int) (*Note, error) {
	nf, err := loadNotes()
	if err != nil {
		return nil, err
	}
	for i := range nf.Notes {
		if nf.Notes[i].ID == id && !nf.Notes[i].Deleted {
			nf.Notes[i].Deleted = true
			nf.Notes[i].DeletedAt = time.Now().Format(time.RFC3339)
			nf.LastDeleted = appendUnique(nf.LastDeleted, id)
			if err := saveNotes(nf); err != nil {
				return nil, err
			}
			n := nf.Notes[i]
			return &n, nil
		}
	}
	return nil, errNotFound(fmt.Sprintf("笔记 #%d 不存在", id))
}

// DeleteNotes 软删除多条笔记（事务性：任一 ID 不存在则整体失败，不持久化部分删除）。
// 返回被删除的笔记列表。
func DeleteNotes(ids []int) ([]Note, error) {
	nf, err := loadNotes()
	if err != nil {
		return nil, err
	}
	deleted := make([]Note, 0, len(ids))
	for _, id := range ids {
		found := false
		for i := range nf.Notes {
			if nf.Notes[i].ID == id && !nf.Notes[i].Deleted {
				nf.Notes[i].Deleted = true
				nf.Notes[i].DeletedAt = time.Now().Format(time.RFC3339)
				nf.LastDeleted = appendUnique(nf.LastDeleted, id)
				deleted = append(deleted, nf.Notes[i])
				found = true
				break
			}
		}
		if !found {
			return nil, errNotFound(fmt.Sprintf("笔记 #%d 不存在", id))
		}
	}
	if err := saveNotes(nf); err != nil {
		return nil, err
	}
	return deleted, nil
}

// UpdateNote 修改一条未删除笔记的文本，保留创建时间 At。
func UpdateNote(id int, text string) (*Note, error) {
	nf, err := loadNotes()
	if err != nil {
		return nil, err
	}
	for i := range nf.Notes {
		if nf.Notes[i].ID == id && !nf.Notes[i].Deleted {
			nf.Notes[i].Text = text
			if err := saveNotes(nf); err != nil {
				return nil, err
			}
			n := nf.Notes[i]
			return &n, nil
		}
	}
	return nil, errNotFound(fmt.Sprintf("笔记 #%d 不存在", id))
}

// ClearNotes 清空全部笔记（逻辑删除标记后持久化），并返回被清空的条数。
// 若原本就没有任何笔记，返回 0 且不视为错误。
func ClearNotes() (int, error) {
	nf, err := loadNotes()
	if err != nil {
		return 0, err
	}
	cleared := 0
	for i := range nf.Notes {
		if !nf.Notes[i].Deleted {
			nf.Notes[i].Deleted = true
			nf.Notes[i].DeletedAt = time.Now().Format(time.RFC3339)
			nf.LastDeleted = appendUnique(nf.LastDeleted, nf.Notes[i].ID)
			cleared++
		}
	}
	if err := saveNotes(nf); err != nil {
		return 0, err
	}
	return cleared, nil
}

// UndoDelete 恢复最近一次删除批次中的笔记，返回恢复数量。
// 若被恢复笔记的原始 ID 已被新笔记占用，则自动为其分配新的最小可用 ID。
func UndoDelete() (int, error) {
	nf, err := loadNotes()
	if err != nil {
		return 0, err
	}
	if len(nf.LastDeleted) == 0 {
		return 0, errNotFound("没有可撤销的删除操作")
	}
	count := 0
	for _, id := range nf.LastDeleted {
		for i := range nf.Notes {
			if nf.Notes[i].ID == id && nf.Notes[i].Deleted {
				if isActiveID(nf.Notes, id) {
					nf.Notes[i].ID = nextNoteID(nf.Notes)
				}
				nf.Notes[i].Deleted = false
				nf.Notes[i].DeletedAt = ""
				count++
				break
			}
		}
	}
	nf.LastDeleted = nil
	if err := saveNotes(nf); err != nil {
		return 0, err
	}
	return count, nil
}

// appendUnique 在切片尾部追加 v，若已存在则忽略。
func appendUnique(s []int, v int) []int {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}
