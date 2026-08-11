package store

// urlFile 是 URL 书签 JSON 文件的内存结构。
type urlFile struct {
	Seq  int   `json:"seq"`  // 自增 ID 计数器
	URLs []URL `json:"urls"`
}

// URL 是一条常用链接书签，别名作为唯一键，ID 为展示用的稳定数字标识。
type URL struct {
	ID    int    `json:"id"`    // 自增数字 ID
	Alias string `json:"alias"` // 别名（唯一）
	Link  string `json:"link"`  // 链接地址
}

const urlStoreFile = "urls.json"

// AddURL 新增或覆盖一条别名链接并保存；新增时分配自增 ID。
func AddURL(alias, link string) error {
	var uf urlFile
	if _, err := loadJSON(urlStoreFile, &uf); err != nil {
		return err
	}
	for i, u := range uf.URLs {
		if u.Alias == alias {
			uf.URLs[i].Link = link // 覆盖链接，保留原 ID
			return saveJSON(urlStoreFile, uf)
		}
	}
	uf.Seq++
	uf.URLs = append(uf.URLs, URL{ID: uf.Seq, Alias: alias, Link: link})
	return saveJSON(urlStoreFile, uf)
}

// ListURLs 返回全部链接书签。兼容旧数据：缺少 ID 时按当前顺序自动分配并写回固化。
func ListURLs() ([]URL, error) {
	var uf urlFile
	if _, err := loadJSON(urlStoreFile, &uf); err != nil {
		return nil, err
	}
	needFix := false
	maxID := uf.Seq
	for _, u := range uf.URLs {
		if u.ID > maxID {
			maxID = u.ID
		}
	}
	for i := range uf.URLs {
		if uf.URLs[i].ID == 0 {
			maxID++
			uf.URLs[i].ID = maxID
			needFix = true
		}
	}
	if needFix {
		uf.Seq = maxID
		_ = saveJSON(urlStoreFile, uf)
	}
	return uf.URLs, nil
}

// GetURL 按别名查找链接，找不到时返回错误。
func GetURL(alias string) (URL, error) {
	urls, err := ListURLs()
	if err != nil {
		return URL{}, err
	}
	for _, u := range urls {
		if u.Alias == alias {
			return u, nil
		}
	}
	return URL{}, errNotFound("未找到别名 " + alias + " 的链接")
}
