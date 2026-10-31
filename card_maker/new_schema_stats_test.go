package card_maker

import (
	"image"
	"testing"
)

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

func TestNewSchemaIconsDecodeAndRemoveCooldownPreviewBackground(t *testing.T) {
	config := NewDefaultConfig(1, "resources/general", "", "resources/fonts")
	maker := NewCardMaker(config)
	for _, kind := range []string{"mana", "activation", "slots", "cooldown"} {
		icon, err := maker.loadNewSchemaIcon(kind)
		if err != nil {
			t.Fatalf("load %s: %v", kind, err)
		}
		if icon.Bounds().Empty() {
			t.Fatalf("%s has empty bounds", kind)
		}
		corners := []image.Point{
			icon.Bounds().Min,
			{X: icon.Bounds().Max.X - 1, Y: icon.Bounds().Min.Y},
			{X: icon.Bounds().Min.X, Y: icon.Bounds().Max.Y - 1},
			{X: icon.Bounds().Max.X - 1, Y: icon.Bounds().Max.Y - 1},
		}
		for _, point := range corners {
			_, _, _, alpha := icon.At(point.X, point.Y).RGBA()
			if alpha != 0 {
				t.Fatalf("%s corner %v alpha = %d, want 0", kind, point, alpha)
			}
		}
	}
}

func TestNewSchemaManaAndCooldownPlaceholders(t *testing.T) {
	config := NewDefaultConfig(1, "resources/general", "", "resources/fonts")
	if _, ok := config.PlacdholderToImage["\\法"]; !ok {
		t.Fatal("mana placeholder is not configured")
	}
	if _, ok := config.PlacdholderToImage["\\冷"]; !ok {
		t.Fatal("cooldown placeholder is not configured")
	}
	config.UseBlankDrawing = true
	maker := NewCardMaker(config)
	card := &CardInfo{
		Number:      "PLACEHOLDER-DEMO",
		Type:        TYPE_ABILITY_ZH,
		Name:        "占位符测试",
		Category:    ELEM_ARCANE_ZH,
		Description: "支付\\法2，冷却\\冷1。",
	}
	if _, err := maker.MakeCard(card); err != nil {
		t.Fatalf("render new schema placeholders: %v", err)
	}
}
