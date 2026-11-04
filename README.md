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

只刷新 JSON 元数据并清理旧编号残留卡图：

```bash
go run . metadata new_mass_producer_params.json
```

只预览指定新版卡牌：

```bash
go run . render BFH001 BFCN003 BWSN002 BAIN002
```

旧版也可按编号单独预览：

```bash
go run . render-old 4111001 1111001
```

## 新版数据约定

- 工作簿是卡牌规则数据的唯一真源。属性码为 `A/F/W/G/E/L/D`（无／火／水／气／地／光／暗）。人物编号使用 `B + 属性码 + H + 三位序号`，例如 `BFH001`、`BGH001`。其他卡使用 `B + 属性码 + 类型码 + 牌级码 + 三位序号`：`L` 为传奇、`N` 为普通、`D` 为衍生，例如 `BFCL001`（基础包／火／伙伴／传奇）、`BGSN002`（基础包／气／技能／普通）、`BFSD001`（基础包／火／技能／衍生）。每个“属性 + 类型 + 牌级”组合都从 `001` 独立编号。
- 支持人物、无、火、水、气、地、光、暗和衍生工作表，共 343 张卡。各属性工作表统一按伙伴牌、技能牌、道具牌分区。
- 法力是单一整数；负载条件和负载是元素向量，两者互不替代。
- 空白数值与明确填写的 `0` 会分别保存。
- 新版已移除旧“威力／防御”字段；技能使用“攻击”和“冷却”。

新版卡面暂用四种可替换的图形徽章：菱形表示法力费用、闪电表示技能使用花费、沙漏表示冷却、四角星表示人物技能槽。徽章均带边框、底色和数值；技能牌始终显示使用花费（包括 0），其他数值为零或空白时不显示。

新版配置指向 `sheets/重置版/奥术纪元_基础包卡牌设计_七系补全_整理版.xlsx` 和同目录的插画迁移表。`prepare-new-drawings` 会把已有旧插画按新版编号同步到 `new_drawings`，并删除已被新编号替代的旧文件名；没有旧插画的原创卡会列入缺失清单。配置项 `use_blank_artwork_for_missing` 可选；设为 `true` 时，缺失插画会使用灰色空白图继续生成卡面，适合完整流程测试。

输出包括：

- `new_output/all_card_infos.json`
- `new_output/simplified_card_infos.json`
- 以新版字符串卡号命名的卡图

卡图目录统一使用七系短名称 `无／火／水／气／地／光／暗`；工作簿中的 `奥术／火焰／水纹／大气／地脉／光辉／暗影` 会在输出路径中自动转换，不影响卡面规则数据。

如需核对新版图片是否确实来自迁移表指定的老版同名卡，可运行：

```bash
go run ./cmd/audit-new-artwork
```

工作簿扩充卡池后，可先按当前新旧 JSON 的唯一同名卡重建迁移表，再复制原画并审计：

```bash
go run ./cmd/rebuild-artwork-migration
go run ./cmd/prepare-new-drawings
go run ./cmd/audit-new-artwork
```

重新整理当前工作簿的编号和属性表分区时运行：

```bash
go run ./cmd/migrate-new-ids
go run ./cmd/prepare-new-drawings
```

为另一份新版工作簿补齐牌级编号和分区、但不修改当前迁移表和 `new_drawings` 时，可运行：

```bash
go run ./cmd/migrate-new-ids --workbook "sheets/重置版/工作簿.xlsm" --workbook-only
```

## 测试

```bash
go test ./...
```

测试覆盖新版整数、元素、合法卡号门控、表头扫描和完整七系工作簿的 343 张卡数量校验；若本地没有 `sheets/重置版` 工作簿，集成测试会跳过。

## 牌组规则

新版牌组为 1 张人物牌加 40 张主牌，技能混入主牌；普通牌至多 2 张，传奇牌至多 1 张。技能与其他主牌共用主牌牌背，人物牌使用独立牌背。

`DeckBuild.lua` 保留为旧数字卡号的 Tabletop Simulator 场景脚本。新版请使用 `DeckBuildNew.lua`：它直接按新版自描述卡号构造 `ArcaneImages/new_output` 卡图地址，人物牌使用独立牌背，其余主牌共用主牌牌背。输入以空白分隔卡号，支持 `2xBFCN003` 数量写法，并可用 `//` 分成多个牌堆。
