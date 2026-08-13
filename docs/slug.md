# slug — 生成 URL 友好的 slug

把标题、中文或带符号的文本转换为 URL 友好的 slug（小写、连字符分隔）。零云依赖。

## 用法

| 命令 | 作用 |
|------|------|
| `philokun slug <文本>` | 生成 slug（默认分隔符 `-`） |
| `philokun slug -s <分隔符> <文本>` | 指定分隔符（如 `_`） |

## 示例

```bash
philokun slug "Hello, World!"
# hello-world

philokun slug "如何学习 Go 语言!" -s _
# 如何学习_go_语言

philokun slug "My Post Title"
# my-post-title
```

## 实现说明

- 文本先小写化；字母、数字、CJK 统一表意文字（U+4E00–U+9FFF）保留。
- 空格与 ` -_/.,&+#%?!()` 等符号转为分隔符；连续分隔符合并、首尾去除。
- 中文等非 ASCII 内容会按原字符保留（UTF-8），适合中英混排的 URL。
