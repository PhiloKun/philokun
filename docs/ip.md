# `ip` 公网 IP 查询

查询当前网络的公网 IPv4 地址，纯标准库实现，零外部依赖。

## 用法

```bash
philokun ip
# 公网 IP: 110.80.5.54
```

## 实现说明

依次尝试多个公共接口，自动选用第一个可用的，仅读取 IP 文本、不做任何其他上报：

| 顺序 | 接口 | 说明 |
|------|------|------|
| 1 | `https://api.ipify.org?format=json` | 境外公共接口 |
| 2 | `https://ipapi.co/ip/` | 境外公共接口 |
| 3 | `https://api.ip.sb/ip` | 境外公共接口 |
| 4 | `http://ip.taobao.com/outGetIpInfo?ip=myip` | 国内可访问（兜底） |

响应解析兼容三种格式：JSON 的 `ip` 字段、纯文本 IP、以及淘宝接口嵌套在 `data.ip` 中的结构。

若全部源都不可用（如网络受限），会报告最后一个错误。
