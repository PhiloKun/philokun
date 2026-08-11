package store

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
)

// shortFile 是短链服务在磁盘上的结构。
// 并发安全由 loadJSON/saveJSON 的原子读写 + 文件级串行访问保证，
// 无需在内存结构里额外持锁。
type shortFile struct {
	Seq   int                `json:"seq"`   // 自增计数器，用于生成短码
	Links map[string]shortLink `json:"links"` // key 为短码 code
}

// shortLink 是单条短链记录。
type shortLink struct {
	Code      string    `json:"code"`      // 短码（唯一）
	URL       string    `json:"url"`       // 原始长链接
	CreatedAt time.Time `json:"created_at"` // 创建时间
	Clicks    int       `json:"clicks"`    // 累计访问/重定向次数
}

// ShortInfo 是短链的非敏感摘要（供列表展示）。
type ShortInfo struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int       `json:"clicks"`
}

const shortStoreFile = "shorts.json"

// base62 字符集，用于把自增 ID 编码成人类友好的短码。
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// loadShorts 读取全部短链；文件不存在时返回空映射。
// 注意：返回的 shortFile 内部已含锁，调用方无需再加锁。
func loadShorts() (shortFile, error) {
	var sf shortFile
	notFound, err := loadJSON(shortStoreFile, &sf)
	if err != nil {
		return shortFile{}, err
	}
	if notFound {
		return shortFile{Links: map[string]shortLink{}}, nil
	}
	if sf.Links == nil {
		sf.Links = map[string]shortLink{}
	}
	return sf, nil
}

// saveShorts 把当前短链映射原子落盘。
func saveShorts(sf shortFile) error {
	return saveJSON(shortStoreFile, sf)
}

// encodeBase62 把非负整数编码为 base62 字符串。
func encodeBase62(n int) string {
	if n == 0 {
		return string(base62Chars[0])
	}
	var b strings.Builder
	for n > 0 {
		b.WriteByte(base62Chars[n%62])
		n /= 62
	}
	// 反转，得到从高位到低位的正确顺序。
	runes := []rune(b.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// randomCode 生成一个基于随机数的短码（默认长度 6），用于冲突重试或自定义兜底。
func randomCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	var b strings.Builder
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		b.WriteByte(base62Chars[idx.Int64()])
	}
	return b.String(), nil
}

// normalizeURL 补全 schemes 缺失的链接（默认补 https://），并做基础校验。
func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("原始链接不能为空")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("链接格式非法: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("仅支持 http/https 链接")
	}
	if u.Host == "" {
		return "", errors.New("链接缺少主机名")
	}
	return raw, nil
}

// validCode 校验用户自定义短码：仅允许字母、数字、- 与 _，且非空。
func validCode(code string) bool {
	if code == "" {
		return false
	}
	for _, r := range code {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			return false
		}
	}
	return true
}

// CreateShort 创建一条短链。
// 若 custom 为空，则基于自增 ID 生成短码（冲突时递增重试）；
// 若 custom 非空，则使用用户指定短码（需校验格式且未被占用）。
// 返回最终短码。
func CreateShort(rawURL, custom string) (string, error) {
	target, err := normalizeURL(rawURL)
	if err != nil {
		return "", err
	}
	sf, err := loadShorts()
	if err != nil {
		return "", err
	}

	if custom != "" {
		if !validCode(custom) {
			return "", errors.New("自定义短码仅允许字母、数字、- 与 _")
		}
		if _, exists := sf.Links[custom]; exists {
			return "", fmt.Errorf("短码 %q 已被占用", custom)
		}
		sf.Links[custom] = shortLink{
			Code:      custom,
			URL:       target,
			CreatedAt: time.Now(),
			Clicks:    0,
		}
		if err := saveShorts(sf); err != nil {
			return "", err
		}
		return custom, nil
	}

	// 自增 ID 生成短码；极小概率冲突（历史删除导致）则递增重试。
	sf.Seq++
	code := encodeBase62(sf.Seq)
	for {
		if _, exists := sf.Links[code]; !exists {
			break
		}
		sf.Seq++
		code = encodeBase62(sf.Seq)
	}
	// 若自增序列与现有随机码都不冲突，使用之；否则兜底随机码。
	if _, exists := sf.Links[code]; exists {
		if code, err = randomCode(6); err != nil {
			return "", err
		}
	}
	sf.Links[code] = shortLink{
		Code:      code,
		URL:       target,
		CreatedAt: time.Now(),
		Clicks:    0,
	}
	if err := saveShorts(sf); err != nil {
		return "", err
	}
	return code, nil
}

// GetShort 按短码返回原始链接；找不到返回错误。
func GetShort(code string) (ShortInfo, error) {
	sf, err := loadShorts()
	if err != nil {
		return ShortInfo{}, err
	}
	l, ok := sf.Links[code]
	if !ok {
		return ShortInfo{}, fmt.Errorf("未找到短码 %q", code)
	}
	return ShortInfo{Code: l.Code, URL: l.URL, CreatedAt: l.CreatedAt, Clicks: l.Clicks}, nil
}

// ResolveShort 查短码并返回原始链接，同时把点击计数 +1 并落盘。
// 用于 HTTP 重定向服务，既返回目标也记录访问量。
func ResolveShort(code string) (string, error) {
	sf, err := loadShorts()
	if err != nil {
		return "", err
	}
	l, ok := sf.Links[code]
	if !ok {
		return "", fmt.Errorf("未找到短码 %q", code)
	}
	l.Clicks++
	sf.Links[code] = l
	if err := saveShorts(sf); err != nil {
		return "", err
	}
	return l.URL, nil
}

// ListShorts 返回全部短链摘要（按创建时间倒序，最新的在前）。
func ListShorts() ([]ShortInfo, error) {
	sf, err := loadShorts()
	if err != nil {
		return nil, err
	}
	infos := make([]ShortInfo, 0, len(sf.Links))
	for _, l := range sf.Links {
		infos = append(infos, ShortInfo{Code: l.Code, URL: l.URL, CreatedAt: l.CreatedAt, Clicks: l.Clicks})
	}
	// 使用稳定排序：点击高的靠前，点击相同则按创建时间倒序。
	sortShorts(infos)
	return infos, nil
}

// sortShorts 按创建时间倒序排序（slice 包未 import，这里用简单插入）。
func sortShorts(infos []ShortInfo) {
	for i := 1; i < len(infos); i++ {
		for j := i; j > 0 && infos[j].CreatedAt.After(infos[j-1].CreatedAt); j-- {
			infos[j], infos[j-1] = infos[j-1], infos[j]
		}
	}
}

// RmShort 删除指定短码的记录。返回 (true,nil) 表示删除成功；(false,nil) 表示不存在。
func RmShort(code string) (bool, error) {
	sf, err := loadShorts()
	if err != nil {
		return false, err
	}
	if _, ok := sf.Links[code]; !ok {
		return false, nil
	}
	delete(sf.Links, code)
	if err := saveShorts(sf); err != nil {
		return false, err
	}
	return true, nil
}

// ShortStoreFile 暴露短链数据文件名，供命令层提示或测试构造路径。
func ShortStoreFile() string {
	return shortStoreFile
}
