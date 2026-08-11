# philokun

> 一个用 Go 编写的**个人效率命令行工具**：用一行命令管理你的待办，数据全部存在本地，不依赖任何云服务。

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev/)
[![Cobra](https://img.shields.io/badge/Cli-Cobra-326CE5)](https://github.com/spf13/cobra)
[![GitHub](https://img.shields.io/badge/GitHub-PhiloKun/philokun-blue?logo=github)](https://github.com/PhiloKun/philokun)
[![Gitee](https://img.shields.io/badge/Gitee-PhiloKun/philokun-red?logo=git)](https://gitee.com/PhiloKun/philokun)

---

## 📌 项目简介

`philokun` 是一个轻量、可脚本化的命令行工具，目标是在终端里快速管理个人待办。

- **零依赖云服务**：所有数据以 JSON 形式存在 `~/.philokun/todo.json`，重装命令也不丢数据。
- **基于 Cobra**：命令结构清晰，新增子命令只需新增一个文件并在 `init()` 里注册。
- **符合 Unix 习惯**：错误打到 stderr 并以非零退出码结束，方便在脚本里组合使用。

> 当前为 `0.1.0` 早期版本，已实现待办增、查与版本打印，后续可扩展笔记、日志、完成状态等。

---

## ✨ 功能特性

| 命令 | 作用 |
|------|------|
| `philokun todo add <内容>` | 追加一条待办 |
| `philokun todo list` | 列出全部待办 |
| `philokun version` | 打印当前版本号 |
| `philokun qrcode` | 启动本地网页版二维码生成器 |

---

## 🔳 二维码生成器

`philokun qrcode` 会启动一个本地 Web 服务，打开网页即可把文本、网址、数字等任意内容实时转换成二维码图片，并支持下载（PNG）与复制到剪贴板。界面自适应桌面与移动端。

```bash
# 默认监听 http://127.0.0.1:8080
./philokun qrcode

# 自定义监听地址与端口
./philokun qrcode --host 0.0.0.0 --port 9000
```

在浏览器打开后：

- **实时生成**：输入框内输入内容，停止输入约 250ms 后自动刷新二维码（防抖，响应迅速）。
- **任意长度**：底层库会自动选择 QR 版本以容纳不同长度的数据。
- **下载 / 复制**：一键下载 PNG，或把二维码图片直接复制到剪贴板。
- **响应式**：在手机与桌面浏览器上都能良好显示。

> 二维码通过 `/api/qrcode?text=<内容>` 接口生成，返回标准 PNG，可单独用于脚本或自动化。

---

## 📂 目录结构

```
philokun/
├── main.go                 # 程序入口，调用 cmd.Execute()
├── go.mod / go.sum         # Go 模块与依赖（cobra 等）
├── cmd/                    # 命令定义层（Cobra 命令）
│   ├── root.go             # 根命令 philokun
│   ├── todo.go             # todo 分组命令 + add / list 子命令
│   ├── version.go          # version 子命令
│   ├── qrcode.go           # qrcode 子命令（本地 Web 服务 + 二维码 API）
│   └── qrcode-web/         # 嵌入的二维码网页（输入框/实时生成/下载/复制）
└── internal/
    └── store/
        └── todo.go         # 数据持久化层（读写本地 JSON）
```

设计上把 **“命令怎么跑”**（cmd）与 **“数据怎么存”**（internal/store）分离：
以后想把存储换成数据库或云同步，只改 `store` 包，命令代码一行都不用动。

---

## 🚀 快速开始

### 方式一：本地编译运行

```bash
# 克隆仓库
git clone https://github.com/PhiloKun/philokun.git
cd philokun

# 编译出可执行文件（生成 ./philokun）
go build -o philokun .

# 体验
./philokun version
./philokun todo add 写一份 philokun 的 README
./philokun todo list
```

### 方式二：go install（需 Go 1.22+）

```bash
go install philokun@latest      # 二进制会装到 $GOPATH/bin，记得加入 PATH
philokun version
```

---

## 💡 使用示例

```bash
$ philokun todo add 完成 Git 教程仓库
已添加待办: 完成 Git 教程仓库

$ philokun todo list
1. 完成 Git 教程仓库

$ philokun version
philokun version 0.1.0
```

---

## 🗄️ 数据存储

待办保存在用户主目录下的 JSON 文件：

```
~/.philokun/todo.json
```

结构示例：

```json
{
  "todos": [
    "完成 Git 教程仓库",
    "给 philokun 加完成状态"
  ]
}
```

- 文件不存在时自动创建，读取不到时返回空列表而非报错。
- 该路径在用户主目录下，不污染项目目录，命令重装数据不丢。

---

## 🛠️ 开发指南

**依赖**

```bash
go mod tidy     # 整理依赖
go build ./...  # 编译全部包
go test ./...   # 运行测试
```

**新增一个子命令**（以 `note` 为例）

1. 在 `cmd/` 下新建 `note.go`，定义 `noteCmd` 并在 `init()` 中 `rootCmd.AddCommand(noteCmd)`；
2. 业务逻辑所需的读写放进 `internal/store`，保持分层清晰；
3. `go build -o philokun .` 后试用。

**版本号发布**

`cmd/version.go` 中的 `version` 变量即当前版本。进阶做法是用编译注入，免改源码：

```bash
go build -ldflags "-X philokun/cmd.version=1.2.3" -o philokun .
```

---

## 🔗 相关链接

- GitHub 仓库：<https://github.com/PhiloKun/philokun>
- Gitee 仓库：<https://gitee.com/PhiloKun/philokun>

---

## 📝 许可证

本项目采用 **MIT 许可证** 发布。

---

*最后更新：2026-08-10*
