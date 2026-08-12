# `base64` 编解码

对文本做 base64 编码或解码，使用标准库 `encoding/base64`。可用于快速处理 token、签名、二进制文本化等场景。

## 用法

```bash
# 编码
philokun base64 encode "hello world"
# aGVsbG8gd29ybGQ=

# 解码
philokun base64 decode "aGVsbG8gd29ybGQ="
# hello world

# 从 stdin 读取（管道）
echo hello | philokun base64 encode -
# aGVsbG8=
```

## 参数

| 参数 | 说明 |
|------|------|
| 第一个参数 | `encode` 或 `decode` |
| 第二个参数 | 待处理文本；传入 `-` 表示从 stdin 读取 |
| （无第二参数） | 从 stdin 读取 |

解码时若输入不是合法 base64，会报错提示。使用标准 base64 编码表（非 URL-safe）。

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/base64.go` | 解析 encode/decode 与输入来源，调用 `base64.StdEncoding` |
