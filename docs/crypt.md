# crypt — 文件加解密

对单个文件做对称加密 / 解密，使用与密码保险箱（`pass`）同一套算法：scrypt 派生密钥 + AES-256-GCM。密文落盘为 JSON 信封（含随机 salt / nonce），相同明文每次加密结果都不同。

## 用法

```bash
philokun crypt enc secret.txt              # 加密，输出 secret.txt.crypt
philokun crypt enc secret.txt -o x.bin     # 指定输出文件
philokun crypt dec secret.txt.crypt        # 解密，默认输出 secret.txt（去掉 .crypt）
philokun crypt dec secret.txt.crypt -o out # 指定输出文件
```

## 子命令

| 子命令 | 作用 |
|--------|------|
| `crypt enc <文件> [-o 输出]` | 加密文件，密码在终端输入且不回显、不存储 |
| `crypt dec <文件> [-o 输出]` | 解密文件，密码错误会明确报错 |

## 示例

```bash
$ philokun crypt enc note.txt
请输入加密密码（不会存储）: ****
已加密 -> note.txt.crypt

$ philokun crypt dec note.txt.crypt
请输入解密密码: ****
已解密 -> note.txt
```

## 安全说明

- 密钥由你输入的密码经 scrypt（N=1<<15, r=8, p=1）派生，**密码不落盘**。
- 加密格式为 `AES-256-GCM`，带完整性校验：密码错误会被 GCM 直接拒绝（输出“解密失败，可能是密码错误”），且密文无法被篡改。
- 每个文件独立随机生成 salt 与 nonce，无全局密钥。
- 与 `pass`（保险箱）的区别：`crypt` 针对任意文件做一次性加解密；`pass` 是结构化的口令保管。
- 忘记密码 = 数据不可恢复，请妥善记忆。
