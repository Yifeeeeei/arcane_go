package main

import (
	"os"
	"testing"

	"github.com/Yifeeeeei/arcane_go/card_maker"
)

func TestParseOptionalIntKeepsBlankAndZeroDistinct(t *testing.T) {
	blank, err := parseOptionalInt("", "value")
	if err != nil || blank != nil {
		t.Fatalf("blank = %v, %v", blank, err)
	}
	zero, err := parseOptionalInt("0", "value")
	if err != nil || zero == nil || *zero != 0 {
		t.Fatalf("zero = %v, %v", zero, err)
	}
	two, err := parseOptionalInt("2.0", "value")
	if err != nil || two == nil || *two != 2 {
		t.Fatalf("2.0 = %v, %v", two, err)
	}
	for _, bad := range []string{"2.5", "-1", "oops"} {
		if _, err := parseOptionalInt(bad, "value"); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestParseElementsNewOrderAndAliases(t *testing.T) {
	elems, err := parseElements("火焰2 / 水8 奥术1")
	if err != nil {
		t.Fatal(err)
	}
	if elems["火"] != 2 || elems["水"] != 8 || elems["奥术"] != 1 {
		t.Fatalf("unexpected elements: %#v", elems)
	}
	if _, err = parseElements("雷3"); err == nil {
		t.Fatal("unknown element accepted")
	}
	legacyOrder, err := parseElements("2奥术 / 1火")
	if err != nil || legacyOrder["奥术"] != 2 || legacyOrder["火"] != 1 {
		t.Fatalf("number-first elements = %#v, %v", legacyOrder, err)
	}
}

func TestLegacySchemaStillParses(t *testing.T) {
	if got := parseLegacyElements("2火 1无"); got["火"] != 2 || got["奥术"] != 1 {
		t.Fatalf("legacy elements = %#v", got)
	}
	tests := map[string]string{"1111001": "伙伴", "2111001": "道具", "3111001": "技能", "4111001": "人物"}
	for id, want := range tests {
		got, err := legacyType(id)
		if err != nil || got != want {
			t.Errorf("legacyType(%s) = %q, %v; want %q", id, got, err, want)
		}
	}
}

func TestParseSheetFindsDescriptiveHeader(t *testing.T) {
	rows := [][]string{{""}, {"标题"}, {"卡号", "卡名", "效果"}, {"BFCN001", "测试", "原文"}}
	parsed, err := parseSheet(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || parsed[0]["卡号"] != "BFCN001" {
		t.Fatalf("unexpected rows: %#v", parsed)
	}
}

func TestLegalCardIDsGateRows(t *testing.T) {
	for _, id := range []string{"BFCN001", "BFSD001", "4111001"} {
		if !newCardID.MatchString(id) && !legacyCardID.MatchString(id) {
			t.Errorf("legal card ID %q was rejected", id)
		}
	}
	for _, value := range []string{"伙伴牌", "说明", "BWIN", "123"} {
		if newCardID.MatchString(value) || legacyCardID.MatchString(value) {
			t.Errorf("non-card row %q passed the card ID gate", value)
		}
	}
}

func TestNewCardIDEncodesMainCardRarity(t *testing.T) {
	for _, id := range []string{"BFH001", "BGH001", "BFCL001", "BGCN001", "BFCN003", "BFCD001", "BASL001", "BWIN014"} {
		if !newCardID.MatchString(id) {
			t.Errorf("valid redesigned ID rejected: %s", id)
		}
	}
	for _, id := range []string{"BFC001", "BRH001", "BFCH001", "BFCX001", "BFHＮ001"} {
		if newCardID.MatchString(id) {
			t.Errorf("invalid redesigned ID accepted: %s", id)
		}
	}
}

func TestNewCardIDSemanticsMatchCardFields(t *testing.T) {
	valid := card_maker.CardInfo{Number: "BFCN001", Category: "火", Type: card_maker.TYPE_UNIT_ZH, Rarity: "普通"}
	if err := validateNewCardIDSemantics(valid); err != nil {
		t.Fatal(err)
	}
	for _, card := range []card_maker.CardInfo{
		{Number: "BGCN001", Category: "火", Type: card_maker.TYPE_UNIT_ZH, Rarity: "普通"},
		{Number: "BFSN001", Category: "火", Type: card_maker.TYPE_UNIT_ZH, Rarity: "普通"},
		{Number: "BFCL001", Category: "火", Type: card_maker.TYPE_UNIT_ZH, Rarity: "普通"},
	} {
		if err := validateNewCardIDSemantics(card); err == nil {
			t.Errorf("semantic mismatch accepted: %#v", card)
		}
	}
}

func TestRedesignedWorkbook(t *testing.T) {
	if _, err := os.Stat("sheets/重置版/奥术纪元_基础包卡牌设计_v0.2.xlsm"); err != nil {
		t.Skip("local redesign workbook is not present")
	}
	m, err := NewMassProducer("new_mass_producer_params.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Validate(); err != nil {
		t.Fatal(err)
	}
	byID := map[string]card_maker.CardInfo{}
	quotes := 0
	for _, c := range m.AllCardInfos {
		byID[c.Number] = c
		if c.Quote != "" {
			quotes++
		}
	}
	for _, id := range []string{"BFH001", "BGH001", "BAIL001", "BACL001", "BFCL001", "BWIL001"} {
		if _, ok := byID[id]; !ok {
			t.Errorf("missing %s", id)
		}
	}
	if quotes == 0 {
		t.Error("expected at least one nonempty quote")
	}
	bh001 := byID["BFH001"]
	if bh001.AttackValue == nil || *bh001.AttackValue != 0 || bh001.LifeValue == nil || *bh001.LifeValue != 15 || bh001.SkillSlots == nil || *bh001.SkillSlots != 3 || bh001.ActivationMana != nil {
		t.Errorf("unexpected BFH001: %s", bh001.String())
	}
	bdh005 := byID["BDH005"]
	if bdh005.ActivationMana != nil || bdh005.Cooldown != nil {
		t.Errorf("non-skill usage cost or cooldown was retained: %s", bdh005.String())
	}
	if got := byID["BACL001"].OutputPath; got != "new_output/基础包/伙伴/无/BACL001.jpg" {
		t.Errorf("short element output path = %q", got)
	}
}

func TestMissingArtworkCanUseBlankDrawing(t *testing.T) {
	if _, err := os.Stat("sheets/重置版/奥术纪元_基础包卡牌设计_v0.2.xlsm"); err != nil {
		t.Skip("local redesign workbook is not present")
	}
	m, err := NewMassProducer("new_mass_producer_params.json")
	if err != nil {
		t.Fatal(err)
	}
	m.Params.DrawingPaths = []string{t.TempDir()}
	m.Params.OutputPath = t.TempDir()
	m.Params.UseBlankArtworkForMissing = true
	if err = m.ProduceIDs([]string{"BACN002"}); err != nil {
		t.Fatal(err)
	}
	var outputPath string
	for _, card := range m.AllCardInfos {
		if card.Number == "BACN002" {
			outputPath = card.OutputPath
			break
		}
	}
	if outputPath == "" {
		t.Fatal("BACN002 output path not found")
	}
	if _, err = os.Stat(outputPath); err != nil {
		t.Fatalf("blank-artwork card was not generated at %s: %v", outputPath, err)
	}
}
