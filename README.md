# philokun

> 一个用 Go 编写的**个人效率命令行工具**：用一行命令管理你的待办，数据全部存在本地，不依赖任何云服务。

[![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go)](https://go.dev/)
[![Cobra](https://img.shields.io/badge/Cli-Cobra-326CE5)](https://github.com/spf13/cobra)
[![GitHub](https://img.shields.io/badge/GitHub-PhiloKun/philokun-blue?logo=github)](https://github.com/PhiloKun/philokun)
[![Gitee](https://img.shields.io/badge/Gitee-PhiloKun/philokun-red?logo=git)](https://gitee.com/PhiloKun/philokun)

---

## 📌 项目简介

`philokun` 是一个轻量、可脚本化的命令行工具，目标是在终端里快速管理个人待办。

- **零依赖云服务**：所有数据以 JSON 形式存在 `~/.philokun/todo.json`，重装命令也不丢数据。
- **基于 Cobra**：命令结构清晰，新增子命令只需新增一个文件并在 `init()` 里注册。
- **符合 Unix 习惯**：错误打到 stderr 并以非零退出码结束，方便在脚本里组合使用。

> 当前为 `1.0.0` 版本，已实现待办管理、终端速记、密码保险箱、天气查询、番茄钟、URL 书签、短链服务、二维码生成等功能。

---

## ✨ 功能特性

| 命令 | 作用 |
|------|------|
| `philokun todo add <内容>` | 追加一条待办 |
| `philokun todo list` | 列出全部待办（含完成状态） |
| `philokun todo done <ID>` | 标记某条待办为已完成 |
| `philokun todo undo <ID>` | 将已完成待办退回为未完成 |
| `philokun todo rm <ID>` | 删除一条待办 |
| `philokun note add <内容>` | 记一条闪念笔记 |
| `philokun note list` | 列出全部笔记 |
| `philokun note search <关键词>` | 按关键词搜索笔记 |
| `philokun note edit <ID>` | 修改一条笔记（`-m` 直接给内容） |
| `philokun note rm <ID...>` | 删除笔记，支持批量与确认（`--yes` 跳过） |
| `philokun note clear` | 软清空全部笔记，带确认（`--yes` 跳过），可 `undo` 恢复 |
| `philokun note purge <ID...>` | 物理删除指定笔记（真正移除，不可撤销） |
| `philokun note purge-all` | 物理清空全部笔记（不可撤销） |
| `philokun note undo` | 撤销最近一次软删除/清空，恢复被删笔记 |
| `philokun pass gen -l <长度>` | 生成高强度随机口令 |
| `philokun pass set <名称>` | 加密保存一条口令到本地保险箱 |
| `philokun pass get <名称> [-j]` | 从保险箱解密取出口令（默认纯文本，-j 输出 JSON） |
| `philokun pass rm <名称> [-f]` | 删除保险箱中的记录（-f 跳过确认） |
| `philokun pass list` | 列出保险箱全部记录（含 ID） |
| `philokun weather <城市...>` | 查询天气（当前/小时/7天/空气质量，多城对比） |
| `philokun weather web <城市...>` | 启动毛玻璃卡片天气网页（多城横滑/渐变背景/预警） |
| `philokun pomodoro <分钟数>` | 启动番茄钟倒计时（别名 `pomo`） |
| `philokun url add <别名> <链接>` | 添加一条别名链接 |
| `philokun url list` | 列出全部链接（含 ID） |
| `philokun url open <别名>` | 在浏览器打开别名对应的链接 |
| `philokun shorten create <链接> [-c 短码]` | 创建短链（可自定义短码） |
| `philokun shorten get <短码>` | 查询短码对应的原始链接与点击数 |
| `philokun shorten list` | 列出全部短链 |
| `philokun shorten rm <短码>` | 删除一条短链 |
| `philokun shorten serve [-p 端口]` | 启动本地短链重定向服务 |
| `philokun version` | 打印当前版本号 |
| `philokun qrcode` | 启动本地网页版二维码生成器 |
| `philokun ip` | 查询本机公网 IP（多源自动回退） |
| `philokun http <方法> <URL>` | 发送 HTTP 请求，打印状态码/响应头/body 摘要 |
| `philokun calc <表达式>` | 数学表达式计算（支持 + - * / ^ 与括号，含 sqrt/abs） |
| `philokun base64 <encode\|decode> <文本>` | base64 编码 / 解码（支持从 stdin 读取） |

---

## 📚 功能文档

各功能的详细使用说明与实现/测试说明，见 [`docs/`](./docs) 目录：

- [待办命令 `todo`](./docs/todo.md)
- [笔记命令 `note`](./docs/note.md)
- [密码工具 `pass`](./docs/pass.md)
- [天气查询 `weather`](./docs/weather.md)
- [番茄钟 `pomodoro`](./docs/pomodoro.md)
- [链接管理 `url`](./docs/url.md)
- [短链服务 `shorten`](./docs/shorten.md)
- [版本命令 `version`](./docs/version.md)
- [二维码生成器 `qrcode`](./docs/qrcode.md)
- [公网 IP 查询 `ip`](./docs/ip.md)
- [HTTP 请求 `http`](./docs/http.md)
- [表达式计算器 `calc`](./docs/calc.md)
- [base64 编解码 `base64`](./docs/base64.md)

---

## 🔳 二维码生成器

`philokun qrcode` 会启动一个本地 Web 服务，打开网页即可把文本、网址、数字等任意内容实时转换成二维码图片，并支持下载（PNG）与复制到剪贴板。界面自适应桌面与移动端。

```bash
# 默认监听 http://127.0.0.1:9000
./philokun qrcode

# 自定义监听地址与端口
./philokun qrcode --host 0.0.0.0 --port 8080
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
├── go.mod / go.sum         # Go 模块与依赖（cobra / x-crypto 等）
├── cmd/                    # 命令定义层（Cobra 命令）
│   ├── root.go             # 根命令 philokun
│   ├── todo.go             # todo 分组命令 + add / list / done / rm
│   ├── note.go             # note 分组命令 + add / list / search
│   ├── pass.go             # pass 分组命令 + gen / set / get
│   ├── weather.go          # weather 子命令（Open-Meteo 联网查询）
│   ├── pomodoro.go         # pomodoro 子命令（番茄钟倒计时）
│   ├── url.go              # url 分组命令 + add / list / open
│   ├── shorten.go          # shorten 分组命令 + create / get / list / rm / serve
│   ├── version.go          # version 子命令
│   ├── qrcode.go           # qrcode 子命令（本地 Web 服务 + 二维码 API）
│   ├── ip.go               # ip 子命令（公网 IP 查询，多源回退）
│   ├── http.go             # http 子命令（HTTP 请求快捷版）
│   ├── calc.go             # calc 子命令（数学表达式求值）
│   ├── base64.go           # base64 子命令（编解码）
│   └── qrcode-web/         # 嵌入的二维码网页（输入框/实时生成/下载/复制）
└── internal/
    └── store/
        ├── file.go         # 通用 JSON 读写辅助
        ├── todo.go         # 待办持久化（含完成状态/删除）
        ├── note.go         # 笔记持久化 + 全文检索
        ├── url.go          # URL 书签持久化
        ├── short.go        # 短链持久化（base62 短码 / CRUD / 点击计数）
        ├── pass.go         # 加密保险箱（scrypt + AES-GCM）
        └── qrcode.go       # 二维码 PNG 生成
```

设计上把 **“命令怎么跑”**（cmd）与 **“数据怎么存”**（internal/store）分离：
以后想把存储换成数据库或云同步，只改 `store` 包，命令代码一行都不用动。

---

## 🚀 安装

### 方式一：一键安装脚本（推荐，普通用户）

自动识别操作系统/架构，从 Release 下载对应二进制并安装到 `~/.local/bin`：

```bash
# 安装最新版（默认走 GitHub）
curl -sSfL https://raw.githubusercontent.com/PhiloKun/philokun/main/install.sh | sh

# 安装指定版本
curl -sSfL https://raw.githubusercontent.com/PhiloKun/philokun/main/install.sh | sh -s -- v1.0.0

# 国内加速（走 Gitee，需显式指定版本）
curl -sSfL https://gitee.com/PhiloKun/philokun/raw/main/install.sh | RELEASE_MIRROR=gitee sh -s -- v1.0.0
```

> 脚本默认从 GitHub 下载；若主源失败会自动回退到另一个源。国内网络不佳时
> 直接用上面的 `RELEASE_MIRROR=gitee` 一行即可稳定安装。

安装完成后若提示 `~/.local/bin` 不在 PATH，执行：

```bash
export PATH="$PATH:$HOME/.local/bin"
# 或写入 shell 配置（Zsh 用户）:
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.zshrc && source ~/.zshrc
```

然后 `philokun version` 验证。

### 方式二：下载 Release 二进制（手动）

到 [GitHub Releases](https://github.com/PhiloKun/philokun/releases) 或
[Gitee Releases](https://gitee.com/PhiloKun/philokun/releases) 下载对应平台的文件
（如 `philokun-darwin-arm64`、`philokun-windows-amd64.exe`），解压后放入 PATH 即可。

### 方式三：go install（需 Go 1.26+，适合开发者）

```bash
go install github.com/philokun@latest      # 二进制装到 $GOPATH/bin，记得加入 PATH
philokun version
```

### 开发者本地编译

```bash
git clone https://github.com/PhiloKun/philokun.git
cd philokun
go build -o philokun .
./philokun version
```

---

## 🚀 快速开始

---

## 💡 使用示例

```bash
$ philokun todo add 完成 Git 教程仓库
已添加待办: 完成 Git 教程仓库

$ philokun todo list
1. 完成 Git 教程仓库

$ philokun version
philokun version 1.0.0
```

---

## 🗄️ 数据存储

所有数据以 JSON 形式保存在用户主目录下的 `~/.philokun/`，不依赖任何云服务：

| 文件 | 用途 |
|------|------|
| `todo.json` | 待办（含 ID / 完成状态） |
| `notes.json` | 闪念笔记 |
| `urls.json` | URL 书签 |
| `shorts.json` | 短链映射（短码 ↔ 长链接，含点击数） |
| `vault.json` | 加密口令保险箱（scrypt + AES-GCM） |

`todo.json` 结构示例：

```json
{
  "seq": 2,
  "todos": [
    { "id": 1, "text": "完成 Git 教程仓库", "done": true },
    { "id": 2, "text": "给 philokun 加完成状态", "done": false }
  ]
}
```

- 文件不存在时自动创建；读取不到时返回空列表而非报错。
- 若 `todo.json` 是旧的字符串数组格式，会自动迁移为含 ID / 完成状态的对象格式。
- 若文件严重损坏无法解析，会备份为 `todo.json.bak.<时间戳>` 后重置，避免命令崩溃。
- 该路径在用户主目录下，不污染项目目录，命令重装数据不丢。
- `vault.json` 中只保存派生密钥所用的 salt 与密文，主密码**永不落盘**。

---

## 🛠️ 开发指南

**依赖**

```bash
go mod tidy     # 整理依赖
go build ./...  # 编译全部包
go test ./...   # 运行测试
```

**测试**

每个功能在开发时都通过编写单元测试验证（CRUD、加密往返、口令生成、命令注册等），
测试在验证通过后即删除，保持仓库整洁。如需本地复跑，可临时用标准库 `testing` 编写：

```bash
go test ./...      # 运行全部包测试
go test ./... -v   # 查看每个用例的详细输出
```

测试通过 `t.Setenv("HOME", tmp)` 把数据重定向到临时目录，不会污染你真实的 `~/.philokun`。

**新增一个子命令**（以 `note` 为例）

1. 在 `cmd/` 下新建 `note.go`，定义 `noteCmd` 并在 `init()` 中 `rootCmd.AddCommand(noteCmd)`；
2. 业务逻辑所需的读写放进 `internal/store`，保持分层清晰；
3. `go build -o philokun .` 后试用。

**版本号发布**

`cmd/version.go` 中的 `version` 变量即当前版本。进阶做法是用编译注入，免改源码：

```bash
go build -ldflags "-X github.com/philokun/cmd.version=1.2.3" -o philokun .
```

**发版（生成多平台二进制）**

仓库内置交叉编译脚本，一键产出 `dist/` 下各平台二进制与校验和：

```bash
./scripts/release.sh 1.0.0          # 指定版本
# 或自动读取 cmd/version.go 的版本号
./scripts/release.sh
```

把 `dist/philokun-*` 与 `dist/checksums-*.sha256` 上传到 GitHub / Gitee 的 Release
（tag 形如 `v1.0.0`），普通用户即可用上面的「一键安装脚本」或「下载二进制」方式安装。

> `dist/` 已在 `.gitignore` 中忽略，二进制请走 Release 而非入库。

---

## 🔗 相关链接

- GitHub 仓库：<https://github.com/PhiloKun/philokun>
- Gitee 仓库：<https://gitee.com/PhiloKun/philokun>

---

## 📝 许可证

本项目采用 **MIT 许可证** 发布。

---

*最后更新：2026-08-12（新增 ip / http / calc / base64 四个命令）*
