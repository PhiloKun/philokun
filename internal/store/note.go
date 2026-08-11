package store

import (
	"strings"
	"time"
)

// noteFile 是笔记 JSON 文件的内存结构。
type noteFile struct {
	Seq   int    `json:"seq"`   // 自增 ID 计数器
	Notes []Note `json:"notes"` // 所有笔记
}

// Note 是单条闪念笔记。
type Note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	At   string `json:"at"` // 创建时间（RFC3339）
}

const notesStoreFile = "notes.json"

// AddNote 追加一条笔记并保存，自动记录时间与自增 ID。
func AddNote(text string) error {
	var nf noteFile
	if _, err := loadJSON(notesStoreFile, &nf); err != nil {
		return err
	}
	nf.Seq++
	nf.Notes = append(nf.Notes, Note{
		ID:   nf.Seq,
		Text: text,
		At:   time.Now().Format(time.RFC3339),
	})
	return saveJSON(notesStoreFile, nf)
}

// ListNotes 返回全部笔记（按添加顺序）。
func ListNotes() ([]Note, error) {
	var nf noteFile
	if _, err := loadJSON(notesStoreFile, &nf); err != nil {
		return nil, err
	}
	return nf.Notes, nil
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
