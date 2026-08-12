# json — JSON 格式化 / 压缩 / 校验

处理 JSON 文本：美化（带缩进）、压缩成一行、或仅做语法校验。零云依赖，常配合 `http` 命令处理 API 输出。

## 子命令

| 子命令 | 作用 |
|--------|------|
| `philokun json fmt [文件]` | 格式化（美化）JSON，带 2 空格缩进；省略文件则从 stdin 读取 |
| `philokun json min [文件]` | 压缩 JSON 成一行，去掉多余空白 |
| `philokun json check [文件]` | 仅校验语法是否合法，合法输出 `OK` |

## 示例

```bash
# 从管道格式化
echo '{"a":1,"b":[2,3]}' | philokun json fmt
# {
#   "a": 1,
#   "b": [
#     2,
#     3
#   ]
# }

# 格式化文件
philokun json fmt data.json

# 压缩成一行
philokun json min data.json

# 仅校验
philokun json check data.json   # -> OK

# 配合 http 命令：把 API 返回的 JSON 美化
philokun http GET https://api.github.com/repos/PhiloKun/philokun | philokun json fmt
```

## 实现说明

- 纯标准库 `encoding/json` 实现：`json.Indent` 美化、`json.Compact` 压缩、`json.Valid` 校验。
- 文件不存在或 stdin 为空时给出明确错误；语法错误会输出具体原因。
- 不修改原文件，结果一律打印到 stdout，方便管道串联。
