package main

import (
	"image/jpeg"
	"os"
	"testing"

	"github.com/Yifeeeeei/arcane_go/card_maker"
)

func TestGen(t *testing.T) {
	// TestGen is a test function.
	// it will always take drawing.jpg as the drawing image
	// -- change below
	drawingPath := "."

	cardInfo := &card_maker.CardInfo{
		Number:      "1411101",
		Type:        "技能",
		Name:        "人鱼公主菲尔",
		Tag:         "传奇异兽",
		Category:    "水",
		Description: "\\?\\?\\?\\?1221\\无\\?\\?32321\\暗2\\血\\威\\持\\攻\\火3\\水4\\气5\\地6\\光7\\暗\\?\\?\\?43\\?\\?\\?\\?\\?\\?\\?\\?\\?\\?\\?一张手牌检\\?\\?\\?索一张传奇以外暗的水火道具牌啊阿啊阿啊阿啊阿啊阿啊阿啊阿啊阿啊阿啊阿啊阿啊abcdeft12143241AAAAAAA\\气",
		Quote:       "我是人鱼大王,超级大王，大法师打发时光啊大帅哥发的放大身份高贵撒地方爱上啊",
		ElementsCost: card_maker.NewElements(map[string]int{
			"水": 11,
			"火": 2,
		}),
		ElementsExpense: card_maker.NewElements(map[string]int{
			"水": 1,
			"气": 2,
		}),
		ElementsGain: card_maker.NewElements(map[string]int{
			"光": 5,
		}),
		Duration: -1,
		Power:    7,
		Attack:   -1,
		Life:     1,
	}
	// -- change above

	config := card_maker.NewDefaultConfig(2, "card_maker/resources/general", drawingPath, "card_maker/resources/fonts")

	cardMaker := card_maker.NewCardMaker(config)
	img, err := cardMaker.MakeCard(cardInfo)
	if err != nil {
		t.Error(err)
	}

	//save it to 11111.jpeg
	outFile, err := os.Create("output_image.jpg")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// Set JPEG quality (from 1 to 100, where 100 is the best quality)
	options := &jpeg.Options{Quality: 100}

	// Encode the RGBA image as a JPEG and save to the file
	err = jpeg.Encode(outFile, img, options)
	if err != nil {
		panic(err)
	}

	println("JPEG image saved successfully as output_image.jpg")
}
