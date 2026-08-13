# zip — zip 压缩 / 解压

把文件或目录打包成 zip 归档，或从 zip 解压到目录。基于标准库 `archive/zip`，零第三方依赖。

## 用法

| 命令 | 作用 |
|------|------|
| `philokun zip c <输出.zip> <文件/目录...>` | 创建 zip（可同时打包多个文件 / 目录） |
| `philokun zip x <归档.zip>` | 解压到当前目录 |
| `philokun zip x <归档.zip> -d <目录>` | 解压到指定目录 |

## 示例

```bash
# 打包
philokun zip c out.zip a.txt b.txt dir/
# 已创建 out.zip

# 解压到当前目录
philokun zip x out.zip

# 解压到指定目录
philokun zip x out.zip -d /tmp/unzipped
```

## 实现说明

- 创建：目录用 `filepath.Walk` 递归加入；zip 内路径去掉基目录前缀，只保留相对名。
- 解压：逐文件写入，并做**路径穿越防护**（`filepath.Clean` + 前缀校验），拒绝跳出目标目录的恶意条目。
- 解压时自动创建必要的父目录，目录权限 `0755`、文件 `0644`。
