# `shorten` URL 短链服务

把任意长链接压缩成易分享的短码，并提供本地 HTTP 重定向服务。所有短链以 JSON 形式保存在 `~/.philokun/shorts.json`，**纯本地、零外部依赖、不发任何网络请求**。

## 适用场景

- 分享超长链接（文档、问卷、带追踪参数的 URL）时更简洁。
- 本地搭一个短链跳转服务，配合固定端口做成"个人短域名"。
- 统计某条链接被点击的次数。

## 用法

### 创建短链

```bash
philokun shorten create "https://example.com/very/long/path?utm_source=x"
# 短链已创建: AbC12  ->  https://example.com/very/long/path?utm_source=x

# 自定义短码（仅允许字母、数字、- 与 _）
philokun shorten create "https://my.site" -c mysite
# 短链已创建: mysite  ->  https://my.site
```

- 长链接缺 `http(s)://` 时会自动补 `https://`。
- 仅支持 `http` / `https` 链接，`ftp` 等会被拒绝。
- 自动生成的短码基于自增 ID 做 base62 编码（如 `1`、`aZ`、`10`），保证简短且不冲突；若自定义短码与已有冲突会报错。

### 查询短链

```bash
philokun shorten get mysite
# 短码: mysite
# 原始链接: https://my.site
# 创建时间: 2026-08-11 15:04
# 点击数: 0
```

### 列出全部短链

```bash
philokun shorten list
# AbC12 -> https://example.com/...  (点击 3)
# mysite -> https://my.site  (点击 0)
```

### 删除短链

```bash
philokun shorten rm mysite
# 已删除短码: mysite
```

### 启动本地重定向服务

```bash
philokun shorten serve -p 9010
# 短链重定向服务已启动: http://127.0.0.1:9010  (Ctrl+C 退出)
```

- 浏览器访问 `http://127.0.0.1:9010/<短码>` 即 **302 跳转**到原始链接，并把该短码点击数 +1。
- 访问根路径 `/` 返回简短使用说明。
- 短码不存在返回 `404`。
- 监听地址默认 `127.0.0.1`（仅本机），可用 `--host 0.0.0.0` 暴露到局域网。

## 数据存储

短链保存在 `~/.philokun/shorts.json`：

```json
{
  "seq": 2,
  "links": {
    "AbC12": {
      "code": "AbC12",
      "url": "https://example.com/very/long/path",
      "created_at": "2026-08-11T15:04:05.123456+08:00",
      "clicks": 3
    },
    "mysite": {
      "code": "mysite",
      "url": "https://my.site",
      "created_at": "2026-08-11T15:10:00.000000+08:00",
      "clicks": 0
    }
  }
}
```

- 写入采用原子写（先写临时文件再 `rename`），避免写一半崩溃损坏数据。
- 文件不存在或为空时自动创建，读取不到返回空列表而非报错。

## 测试

代码包含单元测试（`internal/store/short_test.go`）与集成测试（`cmd/shorten_test.go`）。

```bash
go test ./...                # 运行全部测试
go test ./... -run Shorten   # 只跑短链相关
go test ./... -v             # 详细输出
```

测试通过 `t.Setenv("HOME", tmp)` 把数据重定向到临时目录，不会污染真实 `~/.philokun`。
集成测试用 `net/http/httptest` 启动真实 HTTP 服务，验证 `/<短码>` 的 302 跳转与点击计数。

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 存储 | `internal/store/short.go` | 短码生成（base62）、CRUD、点击计数、URL 校验与规范化 |
| 命令 | `cmd/shorten.go` | 子命令解析、`create/get/list/rm/serve`、HTTP 重定向 handler |

短码生成策略：`CreateShort` 先尝试基于自增 `seq` 做 base62 编码；若与历史（被删除）短码冲突则递增重试；用户自定义短码时校验格式并检测占用。重定向通过 `http.Redirect(w, r, target, http.StatusFound)` 完成。
