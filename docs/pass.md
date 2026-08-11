# `pass` 本地密码工具

提供两个能力：**生成高强度随机口令** 与 **加密保存账号口令到本地保险箱**。所有数据纯本地、零云服务。

## 子命令

### `philokun pass gen -l <长度>`

用 `crypto/rand` 生成密码学安全的随机口令（默认 16 位），字符集已剔除易混淆字符（0/O、1/l/I）。

```bash
philokun pass gen -l 20
# Kf3$mP9xQw2!bB7nM4vR
```

### `philokun pass set <名称>`

把一条口令加密存入 `~/.philokun/vault.json`。可通过 `-v` 直接传值，否则在终端安全输入（不回显）。

```bash
philokun pass set email
# 请输入要保存的口令: ********
# 请输入保险箱主密码（用于加密，不会存储）: ********
# 已加密保存 "email"。
```

### `philokun pass get <名称> [-j/--json]`

用主密码解密取出口令。

默认直接输出纯文本：

```bash
philokun pass get email
# 请输入保险箱主密码: ********
# my-secret-token
```

使用 `-j` / `--json` 时，输出结构化的 JSON 对象，包含密码明文、提取时间戳（UTC，RFC3339）与保险箱标识符：

```bash
philokun pass get email -j
# 请输入保险箱主密码: ********
# {
#   "password": "my-secret-token",
#   "retrieved_at": "2026-08-11T09:07:53Z",
#   "vault": "vault.json"
# }
```

JSON 输出会正确转义特殊字符，便于在脚本或其他编程环境中安全解析；原始密码值不变，仅调整外层格式。

### `philokun pass list`

列出保险箱中的全部记录（含数字 ID 标识，不解密口令内容）。别名 `ls`。

```bash
philokun pass list
# 保险箱中的记录：
# 1. email
# 2. wifi
```

### `philokun pass rm <名称> [-f/--force]`

删除保险箱中的指定记录。默认会要求确认（输入 `y` 才执行），加 `-f` 跳过确认。

```bash
philokun pass rm email
# 确认要删除保险箱中的记录 "email" 吗？此操作不可恢复 (y/N): y
# 已删除保险箱中的记录 "email"。

philokun pass rm email -f
# 已删除保险箱中的记录 "email"。
```

记录不存在时不会报错，仅提示无需删除；保险箱未初始化时给出明确错误。

## 加密方案

- **密钥派生**：主密码 + 随机 `salt`，经 `golang.org/x/crypto/scrypt`（N=2^15, r=8, p=1）派生出 32 字节 AES-256 密钥。
- **对称加密**：`crypto/aes` + `cipher.NewGCM`（AES-GCM），提供机密性与完整性校验。
- **落盘内容**：仅保存 `salt`、密文与随机 `nonce`。**主密码永不写入磁盘**，因此即使 `vault.json` 泄露，没有主密码也无法解密。
- 每个 secret 使用独立随机 `nonce`，避免重放风险。

## 数据存储

```
~/.philokun/vault.json
```

```json
{
  "seq": 2,
  "salt": "Base64...",
  "secrets": {
    "email": { "id": 1, "ciphertext": "Base64...", "nonce": "Base64..." },
    "wifi":  { "id": 2, "ciphertext": "Base64...", "nonce": "Base64..." }
  }
}
```

## 实现说明

| 层 | 文件 | 职责 |
|----|------|------|
| 命令 | `cmd/pass.go` | `gen` / `set` / `get` / `list` 子命令，密码用 `golang.org/x/term` 隐藏回显 |
| 存储 | `internal/store/pass.go` | `SetSecret` / `GetSecret` / `ListSecrets`（scrypt + AES-GCM，自增 ID） |

> 主密码一旦忘记无法找回，请务必牢记。

每条 secret 的 `id` 为自增数字标识，用于列表展示；旧数据（无 `id`）在首次列出时按名称排序自动分配并写回。口令内容始终仅以密文存储，列表只展示 `id` 与名称。
