package store

import (
	"testing"
)

// withTmpHome 把数据目录重定向到临时目录，避免污染真实 ~/.philokun。
func withTmpHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
}

func TestEncodeBase62(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{35, "z"}, // a-z 末尾
		{36, "A"}, // A-Z 开头
		{61, "Z"}, // A-Z 末尾
		{62, "10"},
		{3844, "100"}, // 62^2
	}
	for _, c := range cases {
		if got := encodeBase62(c.n); got != c.want {
			t.Errorf("encodeBase62(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"example.com", "https://example.com", false},
		{"http://a.com/x", "http://a.com/x", false},
		{"  https://b.com  ", "https://b.com", false},
		{"", "", true},
		{"ftp://x.com", "", true},
		{"not a url", "", true},
	}
	for _, c := range cases {
		got, err := normalizeURL(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("normalizeURL(%q) 期望错误，却得到 %q", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeURL(%q) 意外错误: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("normalizeURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestValidCode(t *testing.T) {
	valid := []string{"abc", "A1-b", "x_y", "123"}
	invalid := []string{"", "a b", "a/b", "a.b", "a@b"}
	for _, v := range valid {
		if !validCode(v) {
			t.Errorf("validCode(%q) 应为合法", v)
		}
	}
	for _, v := range invalid {
		if validCode(v) {
			t.Errorf("validCode(%q) 应为非法", v)
		}
	}
}

func TestCreateAndGetShort(t *testing.T) {
	withTmpHome(t)
	code, err := CreateShort("https://example.com/very/long/path", "")
	if err != nil {
		t.Fatalf("CreateShort 失败: %v", err)
	}
	if code == "" {
		t.Fatal("返回的短码为空")
	}
	info, err := GetShort(code)
	if err != nil {
		t.Fatalf("GetShort 失败: %v", err)
	}
	if info.URL != "https://example.com/very/long/path" {
		t.Errorf("URL 不匹配: %q", info.URL)
	}
	if info.Clicks != 0 {
		t.Errorf("新建短链点击数应为 0，得到 %d", info.Clicks)
	}
}

func TestCreateShortAutoPrefix(t *testing.T) {
	withTmpHome(t)
	// 连续创建会自动递增短码，且互不冲突。
	seen := map[string]bool{}
	for i := 0; i < 5; i++ {
		code, err := CreateShort("https://example.com", "")
		if err != nil {
			t.Fatalf("第 %d 次 CreateShort 失败: %v", i, err)
		}
		if seen[code] {
			t.Fatalf("短码冲突: %q", code)
		}
		seen[code] = true
	}
}

func TestCreateShortCustom(t *testing.T) {
	withTmpHome(t)
	code, err := CreateShort("https://my.site", "mysite")
	if err != nil {
		t.Fatalf("自定义短码创建失败: %v", err)
	}
	if code != "mysite" {
		t.Errorf("期望自定义短码 mysite，得到 %q", code)
	}
	// 重复自定义短码应报错。
	if _, err := CreateShort("https://other.site", "mysite"); err == nil {
		t.Error("重复自定义短码应返回错误")
	}
	// 非法短码应报错。
	if _, err := CreateShort("https://x.com", "bad code!"); err == nil {
		t.Error("非法短码应返回错误")
	}
}

func TestResolveShortIncrementsClicks(t *testing.T) {
	withTmpHome(t)
	code, _ := CreateShort("https://count.me", "")
	for i := 1; i <= 3; i++ {
		u, err := ResolveShort(code)
		if err != nil {
			t.Fatalf("ResolveShort 失败: %v", err)
		}
		if u != "https://count.me" {
			t.Errorf("ResolveShort 返回 %q", u)
		}
		info, _ := GetShort(code)
		if info.Clicks != i {
			t.Errorf("点击数应为 %d，得到 %d", i, info.Clicks)
		}
	}
}

func TestGetShortNotFound(t *testing.T) {
	withTmpHome(t)
	if _, err := GetShort("nope"); err == nil {
		t.Error("查询不存在的短码应返回错误")
	}
}

func TestRmShort(t *testing.T) {
	withTmpHome(t)
	code, _ := CreateShort("https://del.me", "")
	ok, err := RmShort(code)
	if err != nil || !ok {
		t.Fatalf("RmShort 失败: ok=%v err=%v", ok, err)
	}
	// 删除后再次查询应失败。
	if _, err := GetShort(code); err == nil {
		t.Error("删除后查询应失败")
	}
	// 删除不存在的短码返回 (false, nil)。
	ok, err = RmShort("ghost")
	if err != nil || ok {
		t.Errorf("删除不存在短码应返回 (false, nil)，得到 (ok=%v, err=%v)", ok, err)
	}
}

func TestListShorts(t *testing.T) {
	withTmpHome(t)
	if _, err := CreateShort("https://a.com", "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateShort("https://b.com", "b"); err != nil {
		t.Fatal(err)
	}
	infos, err := ListShorts()
	if err != nil {
		t.Fatalf("ListShorts 失败: %v", err)
	}
	if len(infos) != 2 {
		t.Errorf("期望 2 条短链，得到 %d", len(infos))
	}
}
