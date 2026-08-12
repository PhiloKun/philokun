# `version` 命令

打印当前 `philokun` 的版本号。

```bash
philokun version
```

输出：

```
philokun version 1.0.0
```

## 版本号说明

版本号定义在 `cmd/version.go` 的 `version` 变量中，发布新版本时改这里即可。

进阶做法：用编译注入在构建时指定版本，免去修改源码：

```bash
go build -ldflags "-X github.com/philokun/cmd.version=1.2.3" -o philokun .
```

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/version.go` | 定义 `version` 子命令，输出 `philokun version <版本号>` |

该命令是最简单的子命令示例：只打印信息、不需要参数、不读写任何数据。

## 测试

`cmd/version_test.go` 覆盖：

- `version` 命令输出格式正确（`philokun version <版本>`）；
- 根命令已正确注册 `todo` / `version` / `qrcode` 三个子命令。

运行：

```bash
go test ./...
```
