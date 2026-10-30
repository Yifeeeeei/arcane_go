package card_maker

import "testing"

func TestSkillUsageCostBadgeAlwaysDraws(t *testing.T) {
	zero := 0
	skill := &CardInfo{Type: TYPE_ABILITY_ZH}
	unit := &CardInfo{Type: TYPE_UNIT_ZH}
	if !shouldDrawNewSchemaStat("activation", nil, skill) {
		t.Fatal("skill with blank usage cost should draw a zero badge")
	}
	if !shouldDrawNewSchemaStat("activation", &zero, skill) {
		t.Fatal("skill with explicit zero usage cost should draw a badge")
	}
	if shouldDrawNewSchemaStat("activation", &zero, unit) {
		t.Fatal("non-skill zero activation value should remain hidden")
	}
}
