# rename — 批量重命名

在目录内批量重命名文件，支持前缀、后缀、替换、序号四种模式。一次只启用一种模式，建议先用 `--dry-run` 预览。

## 用法

| 命令 | 作用 |
|------|------|
| `philokun rename --dir <目录> --prefix <前缀>` | 加前缀 |
| `philokun rename --dir <目录> --suffix <后缀>` | 加后缀（位于扩展名之前） |
| `philokun rename --dir <目录> --replace <旧> --with <新>` | 把文件名中的 `<旧>` 替换为 `<新>` |
| `philokun rename --dir <目录> --index <前缀>` | 按序号重命名：`<前缀>1`、`<前缀>2` … |
| `philokun rename ... -r` | 递归子目录 |
| `philokun rename ... --dry-run` | 仅预览，不真正重命名 |

## 示例

```bash
# 加前缀
philokun rename --dir ./photos --prefix "trip-"

# 空格转下划线（先预览）
philokun rename --dir ./docs --replace " " --with "_" --dry-run
philokun rename --dir ./docs --replace " " --with "_"

# 按序号重命名
philokun rename --dir ./data --index "img_"

# 加后缀（保留扩展名）
philokun rename --dir ./data --suffix "_bak" -r
```

## 实现说明

- 目录通过 `--dir` 指定；`--replace` 必须与 `--with` 搭配使用。
- 后缀模式只插在扩展名之前，原扩展名保留。
- 目标文件名已存在时跳过并提示，避免覆盖。
- `--dry-run` 仅打印「原 -> 新」对照，不改动磁盘；不加该 flag 才真正执行。
