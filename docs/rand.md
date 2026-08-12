# uuid / rand — 随机生成

生成 UUID v4 或密码学安全的随机字符串 / 数字。

## uuid

| 用法 | 作用 |
|------|------|
| `philokun uuid` | 生成 1 个 UUID v4 |
| `philokun uuid -n 5` | 批量生成 5 个 |

```bash
$ philokun uuid -n 2
2d76cdcc-4b1a-4e3f-b649-d5f7140c8c60
ba63811d-dc12-4269-89f9-4fbeebbe80f7
```

实现：RFC 4122 v4，版本位/变体位按规范设置，随机源为 `crypto/rand`。

## rand

| 用法 | 作用 |
|------|------|
| `philokun rand -l 16` | 16 位随机字符串（默认字符集，去易混淆字符） |
| `philokun rand -l 32 -c hex` | 32 位十六进制串 |
| `philokun rand -n 1 100` | [1,100] 闭区间随机整数 |
| `philokun rand -i 5 -n 1 100` | 批量 5 个随机整数 |

字符集 `-c`：`default`（去 0/O/1/l/I 的可读集）/ `hex` / `alpha` / `num` / `alnum`。

```bash
$ philokun rand -l 8
3enwyUyN

$ philokun rand -n 1 100
42
```

实现：所有随机性来自 `crypto/rand`（密码学安全），整数用拒绝采样落在 `[min,max]`。

## 与 pass gen 的区别

- `pass gen` 专为口令场景优化（固定可读字符集、带符号）。
- `rand` / `uuid` 更通用：可指定长度、字符集、批量、数字区间，适合 token、文件名、测试数据等。
