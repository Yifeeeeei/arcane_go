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

func TestBlankDrawingUsesConfiguredDrawingSize(t *testing.T) {
	config := NewDefaultConfig(1, "", "", "")
	config.UseBlankDrawing = true
	maker := NewCardMaker(config)
	drawing, err := maker.getDrawing(&CardInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if drawing.Bounds().Dx() != config.DrawingWidth || drawing.Bounds().Dy() != config.DrawingHeight {
		t.Fatalf("blank drawing size = %v, want %dx%d", drawing.Bounds(), config.DrawingWidth, config.DrawingHeight)
	}
	if got := drawing.RGBAAt(0, 0); got.R != 235 || got.G != 235 || got.B != 235 || got.A != 255 {
		t.Fatalf("blank drawing color = %#v", got)
	}
}
