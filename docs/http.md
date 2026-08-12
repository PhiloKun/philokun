# `http` HTTP 请求快捷版

发送一个 HTTP 请求，打印状态码、响应头以及响应体摘要。适合在终端里快速探活接口、查看返回结构，纯标准库实现。

## 用法

```bash
# GET 请求
philokun http GET https://example.com

# POST 请求，带 JSON body
philokun http POST https://httpbin.org/post -d '{"a":1}' -c application/json
```

## 参数

| 参数 | 简写 | 说明 |
|------|------|------|
| `<方法>` | — | HTTP 方法，如 GET / POST / PUT / DELETE（大小写不敏感） |
| `<URL>` | — | 目标地址，需含 `http://` 或 `https://` |
| `--data` | `-d` | 请求体（POST/PUT 等使用） |
| `--content-type` | `-c` | 有 `-d` 时设置的 Content-Type，默认 `application/json` |

## 输出

```
状态码: 200 OK
响应头:
  Content-Type: text/html; charset=utf-8
  ...
响应体 (773 字节，最多显示 4096):
<html>...</html>
```

响应体最多读取 4096 字节用于展示，避免大响应刷屏。请求超时 15 秒。

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/http.go` | 解析方法/URL/body，构造请求并打印响应摘要 |
