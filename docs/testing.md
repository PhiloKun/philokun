# 测试说明

本项目使用 Go 标准库 `testing` 编写单元测试，覆盖命令层（`cmd`）与存储层（`internal/store`）。

## 运行方式

```bash
go test ./...          # 运行全部测试
go test ./... -v       # 显示每个测试用例的详细输出
go test ./... -run QR  # 只运行名称匹配的子测试（如二维码相关）
```

## 测试用例一览

| 测试文件 | 测试函数 | 覆盖点 |
|----------|----------|--------|
| `internal/store/todo_test.go` | `TestAddAndListTodos` | 添加后列表数量与顺序 |
| | `TestListTodosEmpty` | 空列表不报错 |
| | `TestDataFileCreated` | 数据文件生成在 `~/.philokun/todo.json` |
| `internal/store/qrcode_test.go` | `TestGenerateQRCodeValid` | 生成合法可解码 PNG |
| | `TestGenerateQRCodeEmpty` | 空内容返回错误 |
| | `TestGenerateQRCodeChinese` | 中文内容可编码 |
| `cmd/version_test.go` | `TestVersionCommand` | 版本输出格式正确 |
| | `TestRootCommandHasSubcommands` | 根命令注册了 todo/version/qrcode |
| `cmd/todo_test.go` | `TestTodoAddAndList` | 多参数空格拼接 |
| | `TestTodoAddRequiresArg` | 无参数时 Cobra 报错 |
| | `TestTodoListEmpty` | 空列表输出提示 |
| `cmd/qrcode_test.go` | `TestQrcodeHandlerValid` | 接口返回合法 PNG |
| | `TestQrcodeHandlerEmpty` | 空内容返回 204 |
| | `TestQrcodeHandlerPostRejected` | POST 被拒绝 405 |

## 测试设计要点

- **不污染真实数据**：测试通过 `t.Setenv("HOME", tmp)` 把数据目录重定向到 `t.TempDir()`，不会触碰你真实的 `~/.philokun`。
- **Web 服务不占端口**：`qrcode` 的 HTTP 接口用 `net/http/httptest` 在内存中模拟请求，避免真正监听端口。
- **输出捕获**：命令使用 `fmt.Printf`/`fmt.Println` 直接打到 stdout 时，测试用管道重定向 `os.Stdout` 来断言输出内容。

## 持续集成建议

每次提交前运行 `go test ./...`，全部通过后再推送，可保证核心功能不被破坏。
