# ArcaneGo

奥术纪元卡牌批量导入与制图工具，要求 Go 1.22 或更新版本。仓库同时保留旧版数字卡号流程和新版字符串卡号流程，两套配置、插画与输出目录彼此独立。

## 快速运行

旧版沿用 `mass_producer_params_xlsx.json`、`drawings` 和 `output`：

```bash
go run .
```

新版使用 `new_mass_producer_params.json`、`new_drawings` 和 `new_output`。首次运行先整理插画：

```bash
go run ./cmd/prepare-new-drawings
go run . new
```

只校验数据、不生成卡图：

```bash
go run . validate mass_producer_params_xlsx.json
go run . validate new_mass_producer_params.json
```

只预览指定新版卡牌：

```bash
go run . render BFH001 BFC003 BWS002 BAI002
```

旧版也可按编号单独预览：

```bash
go run . render-old 4111001 1111001
```

## 新版数据约定

- 工作簿是卡牌规则数据的唯一真源。新版编号使用 `B + 属性码 + 类型码 + 三位序号`，例如 `BFH001`（基础包／火焰／人物）、`BFC001`（基础包／火焰／伙伴）、`BFS001`（基础包／火焰／技能）、`BAI001`（基础包／奥术／道具）。导入仍以表中显式字段为准，不会反向用编号覆盖规则数据。
- 支持人物、奥术、火、水四张工作表，共 134 张卡。
- 法力是单一整数；负载条件和负载是元素向量，两者互不替代。
- 空白数值与明确填写的 `0` 会分别保存。
- 新版已移除旧“威力／防御”字段；技能使用“攻击”和“冷却”。

新版卡面暂用四种可替换的图形徽章：菱形表示法力费用、闪电表示发动法力、环形箭头表示冷却、四角星表示人物技能槽。徽章均带圆框、底色和数值；数值为零或空白时不显示。

新版配置指向 `new_fork/奥术纪元_基础包卡牌设计_v0.2.xlsm` 和插画迁移表。`prepare-new-drawings` 会把已有旧插画按新版编号复制到 `new_drawings`；没有旧插画的原创卡会列入缺失清单。

输出包括：

- `new_output/all_card_infos.json`
- `new_output/simplified_card_infos.json`
- 以新版字符串卡号命名的卡图

## 测试

```bash
go test ./...
```

测试覆盖新版整数、元素、表头扫描和完整工作簿的 134 张卡数量校验；若本地没有 `new_fork` 工作簿，集成测试会跳过。

## 牌组规则

新版牌组为 1 张人物牌加 40 张主牌，技能混入主牌；普通牌至多 2 张，传奇牌至多 1 张。技能与其他主牌共用主牌牌背，人物牌使用独立牌背。

`DeckBuild.lua` 仍属于旧 Tabletop Simulator 场景脚本，目前不支持新版字符串卡号和新版牌组结构，不能用于验证新版牌组。
