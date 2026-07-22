package card_maker

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

func decodeResource(t *testing.T, name string) image.Image {
	t.Helper()

	file, err := os.Open(filepath.Join("resources", "general", name))
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	if format != "png" {
		t.Fatalf("%s format = %q, want png", name, format)
	}
	return img
}

func TestElementBackgroundAssets(t *testing.T) {
	backgrounds := []string{
		"back_air.png",
		"back_dark.png",
		"back_earth.png",
		"back_fire.png",
		"back_light.png",
		"back_none.png",
		"back_water.png",
	}
	want := image.Pt(1180, 1720)

	for _, name := range backgrounds {
		t.Run(name, func(t *testing.T) {
			img := decodeResource(t, name)
			if got := img.Bounds().Size(); got != want {
				t.Fatalf("size = %v, want %v", got, want)
			}
		})
	}
}

func TestTransparentOverlayAssets(t *testing.T) {
	assets := map[string]image.Point{
		"ability_logo.png":   image.Pt(200, 200),
		"attack.png":         image.Pt(300, 300),
		"border_ability.png": image.Pt(1180, 1720),
		"border_item.png":    image.Pt(1180, 1720),
		"border_unit.png":    image.Pt(1180, 1720),
		"duration.png":       image.Pt(300, 300),
		"hero_logo.png":      image.Pt(200, 200),
		"item_logo.png":      image.Pt(200, 200),
		"life.png":           image.Pt(300, 300),
		"power.png":          image.Pt(300, 300),
		"unit_logo.png":      image.Pt(200, 200),
	}

	for name, want := range assets {
		t.Run(name, func(t *testing.T) {
			img := decodeResource(t, name)
			if got := img.Bounds().Size(); got != want {
				t.Fatalf("size = %v, want %v", got, want)
			}

			bounds := img.Bounds()
			corners := []image.Point{
				bounds.Min,
				{X: bounds.Max.X - 1, Y: bounds.Min.Y},
				{X: bounds.Min.X, Y: bounds.Max.Y - 1},
				{X: bounds.Max.X - 1, Y: bounds.Max.Y - 1},
			}
			for _, point := range corners {
				_, _, _, alpha := img.At(point.X, point.Y).RGBA()
				if alpha != 0 {
					t.Fatalf("corner %v alpha = %d, want 0", point, alpha)
				}
			}
		})
	}
}
