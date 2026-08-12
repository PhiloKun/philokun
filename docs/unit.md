# unit — 单位换算

在同类单位之间做常用换算，零依赖，适合快速心算核对。

## 用法

```bash
philokun unit length 1 m cm      # 长度：1 米 -> 厘米
philokun unit weight 1 kg lb     # 重量：1 千克 -> 磅
philokun unit temp 100 c f        # 温度：100 摄氏度 -> 华氏度
```

参数顺序：`<类别> <数值> <源单位> <目标单位>`。

## 支持类别与单位

**length（长度）**：`m` `cm` `km` `mm` `in`(英寸) `ft`(英尺) `mi`(英里) `yd`(码)

**weight（重量）**：`kg` `g` `t`(吨) `lb`(磅) `oz`(盎司)

**temp（温度）**：`c`(摄氏) `f`(华氏) `k`(开尔文)

## 示例

```bash
$ philokun unit length 1 m cm
100 cm

$ philokun unit weight 1 kg lb
2.20459 lb

$ philokun unit temp 100 c f
212 f

$ philokun unit temp 32 f c
0 c
```

## 说明

- 长度/重量以标准换算系数（米 / 千克为基准）计算。
- 温度先转到摄氏度再转到目标单位。
- 整数结果不显示小数点（如 `100 cm`），否则按 `%.6g` 紧凑显示。
- 不识别的单位或类别会给出明确错误提示。
