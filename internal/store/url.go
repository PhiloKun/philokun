package store

// urlFile 是 URL 书签 JSON 文件的内存结构。
type urlFile struct {
	URLs []URL `json:"urls"`
}

// URL 是一条常用链接书签，别名作为唯一键。
type URL struct {
	Alias string `json:"alias"` // 别名（唯一）
	Link  string `json:"link"`  // 链接地址
}

const urlStoreFile = "urls.json"

// AddURL 新增或覆盖一条别名链接并保存。
func AddURL(alias, link string) error {
	var uf urlFile
	if _, err := loadJSON(urlStoreFile, &uf); err != nil {
		return err
	}
	for i, u := range uf.URLs {
		if u.Alias == alias {
			uf.URLs[i].Link = link
			return saveJSON(urlStoreFile, uf)
		}
	}
	uf.URLs = append(uf.URLs, URL{Alias: alias, Link: link})
	return saveJSON(urlStoreFile, uf)
}

// ListURLs 返回全部链接书签。
func ListURLs() ([]URL, error) {
	var uf urlFile
	if _, err := loadJSON(urlStoreFile, &uf); err != nil {
		return nil, err
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
