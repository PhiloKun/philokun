# `qrcode` 二维码生成器

启动一个本地 Web 服务，打开网页即可把文本、网址、数字等任意内容实时转换成二维码图片，并支持下载（PNG）与复制到剪贴板。界面自适应桌面与移动端。

## 用法

```bash
# 默认监听 http://127.0.0.1:9000
philokun qrcode

# 自定义监听地址与端口
philokun qrcode --host 0.0.0.0 --port 8080
```

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--port` | `-p` | `9000` | Web 服务监听端口 |
| `--host` | — | `127.0.0.1` | Web 服务监听地址 |

启动后终端提示：

```
二维码生成器已启动，请在浏览器打开: http://127.0.0.1:9000
按 Ctrl+C 退出。
```

## 网页功能

- **实时生成**：输入框内输入内容，停止输入约 250ms 后自动刷新二维码（防抖，响应迅速）。
- **任意长度**：底层库自动选择 QR 版本以容纳不同长度的数据。
- **下载 / 复制**：一键下载 PNG，或把二维码图片直接复制到剪贴板。
- **响应式**：在手机与桌面浏览器上都能良好显示。

## 后端接口

二维码通过 HTTP 接口生成，可直接用于脚本或自动化：

```
GET /api/qrcode?text=<内容>
```

- 成功：返回 `image/png` 的二维码图片字节流（HTTP 200）。
- 空内容：返回 `204 No Content`，前端据此隐藏图像。
- 非 GET 方法：返回 `405 Method Not Allowed`。

```bash
# 示例：用 curl 下载二维码
curl "http://127.0.0.1:9000/api/qrcode?text=https://github.com/PhiloKun/philokun" -o qr.png
```

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/qrcode.go` | 启动 Web 服务，挂载静态页面与 `/api/qrcode` 接口 |
| 网页 | `cmd/qrcode-web/index.html` | 嵌入的二维码网页（输入框 / 实时生成 / 下载 / 复制） |
| 存储 | `internal/store/qrcode.go` | `GenerateQRCode` 使用 `go-qrcode` 生成 PNG（Medium 纠错级别，512px） |

网页通过 `//go:embed qrcode-web/*` 编译进二进制，无需额外部署静态资源。

## 测试

`cmd/qrcode_test.go` 与 `internal/store/qrcode_test.go` 覆盖：

- 合法文本 / 中文内容能生成可解码的 PNG 且带正确文件头；
- 空内容返回错误（store 层）与 `204`（HTTP 层）；
- POST 请求被拒绝（`405`）；
- 生成的 PNG 能被标准库 `image/png` 解码。

运行：

```bash
go test ./...
```
