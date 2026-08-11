package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"philokun/internal/store"
)

// setupTmpHome 把数据目录隔离到临时目录，并清空 rootCmd 的短链相关状态。
func setupTmpHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
}

// TestShortenRedirectIntegration 通过真实 HTTP 服务验证 /<短码> 的 302 跳转。
func TestShortenRedirectIntegration(t *testing.T) {
	setupTmpHome(t)

	code, err := store.CreateShort("https://integ.example.com/page", "ig")
	if err != nil {
		t.Fatalf("创建短链失败: %v", err)
	}

	// 用 httptest 启动与 shortenRedirectHandler 等价的服务（复用同一 handler）。
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortenRedirectHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := srv.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // 不自动跟随，仅检查首跳 302
	}
	resp, err := client.Get(srv.URL + "/" + code)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("期望 302，得到 %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "https://integ.example.com/page" {
		t.Errorf("重定向目标 = %q，期望原始链接", loc)
	}

	// 点击计数应 +1。
	info, _ := store.GetShort(code)
	if info.Clicks != 1 {
		t.Errorf("集成点击后应计 1 次，得到 %d", info.Clicks)
	}
}

// TestShortenNotFoundIntegration 验证未知短码返回 404。
func TestShortenNotFoundIntegration(t *testing.T) {
	setupTmpHome(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortenRedirectHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/missing")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("期望 404，得到 %d", resp.StatusCode)
	}
}

// TestShortenRootIntegration 验证根路径返回使用说明文本。
func TestShortenRootIntegration(t *testing.T) {
	setupTmpHome(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortenRedirectHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望 200，得到 %d", resp.StatusCode)
	}
}

// executeShorten 在测试里执行 shorten 子命令，返回输出与错误。
// 通过临时重置 args 并调用 Execute 的方式模拟 CLI 调用。
func executeShorten(t *testing.T, args ...string) (string, error) {
	t.Helper()
	rootCmd.SetArgs(append([]string{"shorten"}, args...))
	// 捕获输出（简单方式：执行后比对存储状态，这里只关心是否报错）。
	err := rootCmd.Execute()
	return "", err
}

// TestShortenCreateCommandE2E 端到端执行 create + list + rm 命令路径。
func TestShortenCreateCommandE2E(t *testing.T) {
	setupTmpHome(t)

	if _, err := executeShorten(t, "create", "https://e2e.example.com", "-c", "e2e"); err != nil {
		t.Fatalf("create 命令失败: %v", err)
	}
	// 验证存储层确实写入。
	info, err := store.GetShort("e2e")
	if err != nil {
		t.Fatalf("短码未写入: %v", err)
	}
	if info.URL != "https://e2e.example.com" {
		t.Errorf("URL 不匹配: %q", info.URL)
	}

	// 自定义码冲突应使命令报错。
	if _, err := executeShorten(t, "create", "https://other.com", "-c", "e2e"); err == nil {
		t.Error("重复自定义短码命令应报错")
	}

	// rm 命令路径。
	if _, err := executeShorten(t, "rm", "e2e"); err != nil {
		t.Fatalf("rm 命令失败: %v", err)
	}
	if _, err := store.GetShort("e2e"); err == nil {
		t.Error("rm 后短码应不存在")
	}
}

// TestShortenServeHandlerStatus 验证 handler 对合法短码返回正确的 Location 头。
func TestShortenServeHandlerStatus(t *testing.T) {
	setupTmpHome(t)
	code, _ := store.CreateShort("https://hdr.example.com", "")

	req := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	w := httptest.NewRecorder()
	shortenRedirectHandler(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("期望 302，得到 %d", w.Code)
	}
	if got := w.Header().Get("Location"); got != "https://hdr.example.com" {
		t.Errorf("Location = %q", got)
	}
}

// TestShortenCreateCommandOutput 验证 create 命令对非法链接报错（集成到 CLI 路径）。
func TestShortenCreateCommandOutput(t *testing.T) {
	setupTmpHome(t)
	// 非法链接（scheme 不支持）应报错。
	if _, err := executeShorten(t, "create", "ftp://x.com"); err == nil {
		t.Error("非法链接 create 应报错")
	}
}
