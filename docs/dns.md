# dns — DNS 解析

查询域名的 DNS 记录，基于标准库 `net` 实现，无需额外依赖或 API key。

## 用法

```bash
philokun dns example.com              # 默认返回 A + AAAA
philokun dns example.com -t a         # 仅 A
philokun dns example.com -t aaaa      # 仅 AAAA
philokun dns example.com -t cname     # CNAME
philokun dns example.com -t mx        # MX（含优先级）
philokun dns example.com -t ns        # NS
philokun dns example.com -t all       # 一次查全部类型
```

## 示例

```bash
$ philokun dns example.com -t a
A    172.66.147.243
A    104.20.23.154

$ philokun dns gmail.com -t mx
MX  5   alt1.gmail-smtp-in.l.google.com
MX  10  alt2.gmail-smtp-in.l.google.com
```

## 说明

- 解析结果直接来自本机 DNS 配置，可能与公共 DNS 略有差异。
- 某类型查询失败时（如域名无 CNAME），会在该行提示失败原因，不影响其他类型。
- 与 `ip`、`http` 同属网络工具族。
