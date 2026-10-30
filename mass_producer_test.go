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
	rows := [][]string{{""}, {"标题"}, {"卡号", "卡名", "效果"}, {"BFC001", "测试", "原文"}}
	parsed, err := parseSheet(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || parsed[0]["卡号"] != "BFC001" {
		t.Fatalf("unexpected rows: %#v", parsed)
	}
}

func TestRedesignedWorkbook(t *testing.T) {
	if _, err := os.Stat("new_fork/奥术纪元_基础包卡牌设计_v0.2.xlsm"); err != nil {
		t.Skip("local redesign workbook is not present")
	}
	m, err := NewMassProducer("template_mass_producer_params_xlsx.json")
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
	for _, id := range []string{"BFH001", "BAI001", "BFS001", "BFC001", "BWI001"} {
		if _, ok := byID[id]; !ok {
			t.Errorf("missing %s", id)
		}
	}
	if quotes != 37 {
		t.Errorf("nonempty quotes = %d, want 37", quotes)
	}
	bh001 := byID["BFH001"]
	if bh001.AttackValue == nil || *bh001.AttackValue != 0 || bh001.LifeValue == nil || *bh001.LifeValue != 20 || bh001.SkillSlots == nil || *bh001.SkillSlots != 3 || bh001.ActivationMana != nil {
		t.Errorf("unexpected BFH001: %s", bh001.String())
	}
	bh002 := byID["BFH002"]
	if bh002.ActivationMana == nil || *bh002.ActivationMana != 0 {
		t.Errorf("BFH002 explicit zero was lost: %s", bh002.String())
	}
	bf003 := byID["BFS001"]
	if bf003.Type != "技能" || bf003.Cooldown == nil || *bf003.Cooldown != 2 || bf003.ActivationMana == nil || *bf003.ActivationMana != 4 {
		t.Errorf("unexpected BFS001: %s", bf003.String())
	}
}
