# hash — 哈希计算（md5 / sha1 / sha256 / sha512）

计算字符串或文件的哈希值，用于校验完整性或生成摘要。纯标准库实现，零云依赖。

## 用法

| 命令 | 作用 |
|------|------|
| `philokun hash <文本>` | 计算字符串的 sha256（默认） |
| `philokun hash -a sha256 -a md5 <文本>` | 同时计算多种算法（`-a` 可多次） |
| `philokun hash -f -a sha256 <文件>` | 计算文件哈希（`-f` 进入文件模式） |

## 示例

```bash
# 字符串哈希（默认 sha256）
philokun hash hello
# SHA256   2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824

# 多算法
philokun hash -a md5 -a sha1 hello

# 从 stdin 读
echo -n "secret" | philokun hash -a sha256

# 文件哈希
philokun hash -f -a sha256 ./philokun
# SHA256  xxxx  ./philokun
```

## 实现说明

- 使用 `crypto/md5`、`crypto/sha1`、`crypto/sha256`、`crypto/sha512` 标准库。
- 文件模式用流式 `io.Copy`，对大文件也只占用常量内存。
- 不支持的算法名会给出明确报错。
