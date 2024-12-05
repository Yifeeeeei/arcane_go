package card_maker

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
)

type CardMaker struct {
	Config             Config
	DefaultPlaceholder string
}

func NewCardMaker(config Config) CardMaker {
	return CardMaker{
		Config:             config,
		DefaultPlaceholder: "\u00A0\u00A0",
	}
}

func (cardMaker *CardMaker) translator(zh string) string {
	if zh == ELEM_AIR_ZH {
		return ELEM_AIR_EN
	} else if zh == ELEM_DARK_ZH {
		return ELEM_DARK_EN
	} else if zh == ELEM_EARTH_ZH {
		return ELEM_EARTH_EN
	} else if zh == ELEM_FIRE_ZH {
		return ELEM_FIRE_EN
	} else if zh == ELEM_LIGHT_ZH {
		return ELEM_LIGHT_EN
	} else if zh == ELEM_WATER_ZH {
		return ELEM_WATER_EN
	} else if zh == ELEM_NONE_ZH {
		return ELEM_NONE_EN
	}
	log.Fatalf("Unknown element: %v, returning none\n", zh)
	return ELEM_NONE_EN
}

func (cardMaker *CardMaker) getStringWidth(text string, font *truetype.Font, fontSize float64) int {
	face := truetype.NewFace(font, &truetype.Options{Size: fontSize})
	var width fixed.Int26_6
	for _, runeValue := range text {
		awidth, _ := face.GlyphAdvance(runeValue)
		width += awidth
	}
	widthInPixels := int(width >> 6)
	return widthInPixels
}

func (cardMaker *CardMaker) getStringHeight(font *truetype.Font, fontSize float64) int {
	face := truetype.NewFace(font, &truetype.Options{Size: fontSize})
	metrics := face.Metrics()
	return int((metrics.Ascent + metrics.Descent) >> 6)
}

type placeHolderRecord struct {
	Category string
	Row      int
	Offset   int
}

func strLen(str string) int {
	return len([]rune(str))
}

func (cardMaker *CardMaker) textWrapAndPlaceHolderReplacement(text string, width int, font *truetype.Font, fontSize float64) ([]string, []placeHolderRecord) {
	result := []string{}
	currentLine := ""
	currentLengthCount := 0
	placeHolderLocations := []placeHolderRecord{}
	index := 0
	runes := []rune(text)
	for index < len(runes) {
		character := string(runes[index])
		var recordPlaceHolder []rune
		recordPlaceHolder = nil
		if character == "\\" {
			characters := string(runes[index : index+2])
			if _, ok := cardMaker.Config.PlacdholderToImage[characters]; ok {
				recordPlaceHolder = runes[index : index+2]
				character = string(cardMaker.DefaultPlaceholder)
				index++
			}
		}
		length_taken := cardMaker.getStringWidth(character, font, fontSize)
		if currentLengthCount+length_taken > width {
			result = append(result, currentLine)
			currentLine = ""
			currentLengthCount = 0
		}
		if recordPlaceHolder != nil {
			placeHolderLocations = append(placeHolderLocations, placeHolderRecord{
				Category: string(recordPlaceHolder),
				Row:      len(result),
				Offset:   currentLengthCount,
			})
		}
		currentLine += character
		currentLengthCount += length_taken

		index++

	}
	if currentLine != "" {
		result = append(result, currentLine)
	}
	return result, placeHolderLocations
}

func (cardMaker *CardMaker) adjustImage(img *image.RGBA, targetWidth, targetHeight int) *image.RGBA {
	// Get the original image dimensions
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// Calculate aspect ratios
	imageAspectRatio := float64(width) / float64(height)
	targetAspectRatio := float64(targetWidth) / float64(targetHeight)

	var cropRect image.Rectangle

	// Crop the image based on its aspect ratio compared to the target
	if imageAspectRatio > targetAspectRatio {
		// Image is too wide, crop horizontally
		idealWidth := int(float64(height) * targetAspectRatio)
		left := (width - idealWidth) / 2
		right := left + idealWidth
		cropRect = image.Rect(left, 0, right, height)
	} else if imageAspectRatio < targetAspectRatio {
		// Image is too tall, crop vertically
		idealHeight := int(float64(width) / targetAspectRatio)
		top := (height - idealHeight) / 2
		bottom := top + idealHeight
		cropRect = image.Rect(0, top, width, bottom)
	} else {
		// If the aspect ratios are the same, no cropping is needed
		cropRect = img.Bounds()
	}

	// Crop the image using the calculated crop rectangle
	croppedImg := img.SubImage(cropRect)

	// Resize the cropped image to the target dimensions
	newImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(newImg, newImg.Bounds(), croppedImg, croppedImg.Bounds(), draw.Over, nil)

	return newImg
}

func (cardMaker *CardMaker) loadImage(filePath string) (*image.RGBA, error) {
	img, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	imgData, _, err := image.Decode(img)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(imgData.Bounds())
	draw.Draw(rgba, rgba.Bounds(), imgData, image.Point{0, 0}, draw.Src)

	return rgba, nil
}

func (cardMaker *CardMaker) getImageWithoutExtension(imagePref string) (*image.RGBA, error) {
	extensionList := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
	for _, extension := range extensionList {
		filePath := imagePref + extension

		_, err := os.Stat(filePath)
		if err == nil {
			rgba, err := cardMaker.loadImage(filePath)
			if err != nil {
				log.Fatalf("error loading image %v: %v\n", filePath, err)
				continue
			}
			return rgba, nil
		}
	}
	return nil, fmt.Errorf("no image found for %v", imagePref)
}

func (cardMaker *CardMaker) getDrawing(cardInfo *CardInfo) (*image.RGBA, error) {
	img, err := cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.DrawingPath, cardInfo.Number))
	if err != nil {
		return nil, fmt.Errorf("error getting drawing: %v", err)
	}
	return img, nil
}

func (cardMaker *CardMaker) getBackground(cardInfo *CardInfo) (*image.RGBA, error) {
	bgImage, err := cardMaker.getImageWithoutExtension(
		filepath.Join(
			cardMaker.Config.GeneralPath, cardMaker.Config.ElementBack[cardInfo.Category],
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting background: %v", err)
	}
	bgImage = cardMaker.adjustImage(
		bgImage, cardMaker.Config.CardWidth, cardMaker.Config.CardHeight,
	)
	return bgImage, nil
}

func (cardMaker *CardMaker) getBorder(cardInfo *CardInfo) (*image.RGBA, error) {
	borderImage, err := cardMaker.getImageWithoutExtension(
		filepath.Join(
			cardMaker.Config.GeneralPath, cardMaker.Config.TypeBorder[cardInfo.Type],
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting border: %v", err)
	}
	borderImage = cardMaker.adjustImage(
		borderImage, cardMaker.Config.BorderWidth, cardMaker.Config.BorderHeight,
	)
	return borderImage, nil
}

func (cardMaker *CardMaker) reverseColor(originColor color.RGBA) color.RGBA {
	return color.RGBA{
		R: 255 - originColor.R,
		G: 255 - originColor.G,
		B: 255 - originColor.B,
		A: originColor.A,
	}
}

func (cardMaker *CardMaker) getDigit(numberString string, leftToRightPos int) int {
	if leftToRightPos < 0 || leftToRightPos >= len(numberString) {
		return 0
	}
	return int(numberString[leftToRightPos] - '0')
}

func (cardMaker *CardMaker) isLegend(cardInfo *CardInfo) bool {
	return cardInfo.Type == TYPE_HERO_ZH || (cardMaker.getDigit(cardInfo.Number, 2) == 1)
}

func (cardMaker *CardMaker) drawRoundedRect(img *image.RGBA, left, top, right, bottom, radius, borderWidth int, fillColor, borderColor color.RGBA) *image.RGBA {
	// Create a new gg context on top of the given image dimensions
	dc := gg.NewContextForRGBA(img)

	// Convert the fill and border colors from our custom Color struct

	// Set the fill color
	dc.SetRGBA255(int(fillColor.R), int(fillColor.G), int(fillColor.B), int(fillColor.A))

	// Draw the filled rounded rectangle
	dc.DrawRoundedRectangle(float64(left), float64(top), float64(right-left), float64(bottom-top), float64(radius))
	dc.Fill()

	// Set the border color and width
	dc.SetRGBA255(int(borderColor.R), int(borderColor.G), int(borderColor.B), int(borderColor.A))
	dc.SetLineWidth(float64(borderWidth))

	// Draw the border around the rounded rectangle
	dc.DrawRoundedRectangle(float64(left), float64(top), float64(right-left), float64(bottom-top), float64(radius))
	dc.Stroke()
	return img
}

func (cardMaker *CardMaker) drawRect(img *image.RGBA, left, top, right, bottom, borderWidth int, fillColor, borderColor color.RGBA) *image.RGBA {
	return cardMaker.drawRoundedRect(img, left, top, right, bottom, 0, borderWidth, fillColor, borderColor)
}

func (cardMaker *CardMaker) drawBottomBlock(baseImage *image.RGBA, cardInfo *CardInfo) *image.RGBA {
	top := cardMaker.Config.DrawingToUpper + cardMaker.Config.DrawingHeight
	left := (cardMaker.Config.CardWidth - cardMaker.Config.BottomBlockWidth) / 2
	bottom := cardMaker.Config.DrawingToUpper + cardMaker.Config.DrawingHeight + cardMaker.Config.BottomBlockHeight
	right := left + cardMaker.Config.BottomBlockWidth
	colorFill := cardMaker.Config.BottomBlockColor
	if cardMaker.Config.ReverseColorForHero && cardMaker.GetCardType(cardInfo) == TYPE_HERO_ZH {
		colorFill = cardMaker.reverseColor(colorFill)
	}
	img := cardMaker.drawRect(
		baseImage,
		left,
		top,
		right,
		bottom,
		0,
		colorFill,
		color.RGBA{},
	)
	return img
}
func (cardMaker *CardMaker) overlayImageOntoBase(baseImage, overlayImage *image.RGBA, left, top int) *image.RGBA {
	// Create a new RGBA image that has the same size as the baseImage
	result := image.NewRGBA(baseImage.Bounds())

	// Draw the base image onto the result
	draw.Draw(result, baseImage.Bounds(), baseImage, image.Point{}, draw.Src)

	// Define the position where the overlay image will be placed on the base image
	overlayRect := image.Rect(left, top, left+overlayImage.Bounds().Dx(), top+overlayImage.Bounds().Dy())

	// Draw the overlay image onto the result image at the specified position
	draw.Draw(result, overlayRect, overlayImage, image.Point{}, draw.Over)

	return result
}

func (cardMaker *CardMaker) prepareOutline(cardInfo *CardInfo) (*image.RGBA, error) {
	if cardMaker.isLegend(cardInfo) {
		baseImage, err := cardMaker.getDrawing(cardInfo)
		if err != nil {
			return nil, err
		}
		baseImage = cardMaker.adjustImage(
			baseImage, cardMaker.Config.CardWidth, cardMaker.Config.CardHeight,
		)

		baseImage = cardMaker.drawBottomBlock(baseImage, cardInfo)

		return baseImage, nil

	} else {
		baseImage, err := cardMaker.getBackground(cardInfo)
		if err != nil {
			return nil, err
		}
		drawingImage, err := cardMaker.getDrawing(cardInfo)
		if err != nil {
			return nil, err
		}
		baseImage = cardMaker.adjustImage(baseImage, cardMaker.Config.CardWidth, cardMaker.Config.CardHeight)
		drawingImage = cardMaker.adjustImage(drawingImage, cardMaker.Config.DrawingWidth, cardMaker.Config.DrawingHeight)
		img := cardMaker.overlayImageOntoBase(
			baseImage,
			drawingImage,
			(cardMaker.Config.CardWidth-cardMaker.Config.DrawingWidth)/2,
			cardMaker.Config.DrawingToUpper,
		)
		img = cardMaker.drawBottomBlock(img, cardInfo)
		borderImage, err := cardMaker.getBorder(cardInfo)
		if err != nil {
			return nil, err
		}
		img = cardMaker.overlayImageOntoBase(
			img,
			borderImage,
			(cardMaker.Config.CardWidth-cardMaker.Config.BorderWidth)/2,
			(cardMaker.Config.CardHeight-cardMaker.Config.BorderHeight)/2,
		)
		return img, nil
	}
}

func (cardMaker *CardMaker) addTextToImage(baseImage *image.RGBA, text string, left, top int, font *truetype.Font, fontSize float64, textColor color.RGBA) *image.RGBA {
	// Create a new freetype context
	c := freetype.NewContext()

	// Set the destination image to draw the text
	c.SetDst(baseImage)

	// Set the source color (text color)
	c.SetSrc(image.NewUniform(textColor))

	// Set the font and font size
	c.SetFont(font)
	c.SetFontSize(fontSize)

	// Set the DPI (dots per inch) for high-quality text rendering
	c.SetDPI(72)

	// Set the clip region to the entire image
	c.SetClip(baseImage.Bounds())

	// Set the draw target to the image
	c.SetDst(baseImage)

	// Create a point to position the text on the image
	pt := freetype.Pt(left, top+int(c.PointToFixed(fontSize)>>6)) // Add a small offset to position the text properly

	// Draw the text onto the image
	_, err := c.DrawString(text, pt)
	if err != nil {
		log.Fatalf("Failed to draw text: %v", err)
	}
	return baseImage
}

func (cardmaker *CardMaker) getCategoryImage(category string) (*image.RGBA, error) {
	return cardmaker.getImageWithoutExtension(filepath.Join(cardmaker.Config.GeneralPath, cardmaker.Config.ElementImages[category]))
}

func (cardMaker *CardMaker) getRectFillColor(cardInfo *CardInfo) color.RGBA {
	if cardInfo.Category == ELEM_LIGHT_ZH {
		return color.RGBA{255, 240, 184, 255}
	} else if cardInfo.Category == ELEM_DARK_ZH {
		return color.RGBA{200, 200, 200, 255}
	} else if cardInfo.Category == ELEM_FIRE_ZH {
		return color.RGBA{255, 169, 167, 255}
	} else if cardInfo.Category == ELEM_WATER_ZH {
		return color.RGBA{167, 233, 255, 255}
	} else if cardInfo.Category == ELEM_AIR_ZH {
		return color.RGBA{253, 227, 255, 255}
	} else if cardInfo.Category == ELEM_EARTH_ZH {
		return color.RGBA{233, 203, 177, 255}
	} else if cardInfo.Category == ELEM_NONE_ZH {
		return color.RGBA{250, 250, 250, 255}
	} else {
		log.Fatalf("invalid category: %v\n", cardInfo.Category)
		return color.RGBA{}
	}
}

func (cardMaker *CardMaker) getRectOutlineColor(cardInfo *CardInfo) color.RGBA {
	if cardInfo.Category == ELEM_LIGHT_ZH {
		return color.RGBA{0, 0, 0, 255}
	} else if cardInfo.Category == ELEM_DARK_ZH {
		return color.RGBA{0, 0, 0, 255}
	} else if cardInfo.Category == ELEM_FIRE_ZH {
		return color.RGBA{0, 0, 0, 255}
	} else if cardInfo.Category == ELEM_WATER_ZH {
		return color.RGBA{0, 0, 0, 255}
	} else if cardInfo.Category == ELEM_AIR_ZH {
		return color.RGBA{0, 0, 0, 255}
	} else if cardInfo.Category == ELEM_EARTH_ZH {
		return color.RGBA{0, 0, 0, 255}
	} else if cardInfo.Category == ELEM_NONE_ZH {
		return color.RGBA{0, 0, 0, 255}
	} else {
		log.Fatalf("invalid category: %v\n", cardInfo.Category)
		return color.RGBA{}
	}
}

func (cardMaker *CardMaker) loadFont(fontPath string) (*truetype.Font, error) {
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}

	return font, nil
}

func (cardMaker *CardMaker) drawCategoryAndName(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	textFont, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.NameFont))
	if err != nil {
		return baseImage, err
	}
	lengthEstimate := cardMaker.getStringWidth(cardInfo.Name, textFont, float64(cardMaker.Config.NameFontSize))
	rectangleWidth := lengthEstimate + 2*(cardMaker.Config.NameTextToLeft-cardMaker.Config.NameRectLeft)
	left := cardMaker.Config.NameRectLeft
	top := cardMaker.Config.NameRectTop
	right := left + rectangleWidth
	bottom := top + cardMaker.Config.NameRectHeight
	img := cardMaker.drawRoundedRect(
		baseImage,
		left,
		top,
		right,
		bottom,
		cardMaker.Config.NameRectRadius,
		cardMaker.Config.NameRectOutlineWidth,
		cardMaker.getRectFillColor(cardInfo),
		cardMaker.getRectOutlineColor(cardInfo),
	)
	textLeft := cardMaker.Config.NameTextToLeft + cardMaker.Config.NameTextLeftCompensation
	textHeight := cardMaker.getStringHeight(textFont, float64(cardMaker.Config.NameFontSize))
	textTop := cardMaker.Config.NameRectTop + (cardMaker.Config.NameRectHeight-textHeight)/2 - cardMaker.Config.NameTextTopCompensation
	// add name
	img = cardMaker.addTextToImage(img, cardInfo.Name, textLeft, textTop, textFont, float64(cardMaker.Config.NameFontSize), cardMaker.Config.NameTextFontColor)
	// add category
	categoryImage, err := cardMaker.getCategoryImage(cardInfo.Category)
	if err != nil {
		return img, err
	}
	categoryImage = cardMaker.adjustImage(categoryImage, cardMaker.Config.NameCategoryWidth, cardMaker.Config.NameCategoryWidth)
	img = cardMaker.overlayImageOntoBase(
		img,
		categoryImage,
		cardMaker.Config.NameCategoryLeft,
		cardMaker.Config.NameCategoryTop,
	)
	return img, nil
}

type ElemAndVal struct {
	Elem string
	Val  int
}

func (cardMaker *CardMaker) drawCost(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {

	allCosts := []ElemAndVal{}
	for ele, val := range cardInfo.ElementsCost {
		if val > 0 {
			allCosts = append(allCosts, ElemAndVal{Elem: ele, Val: val})
		}
	}
	if len(allCosts) == 0 {
		return baseImage, nil
	}

	font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.CostFont))
	if err != nil {
		return baseImage, err
	}
	numberLength := 0
	for _, elemAndCost := range allCosts {
		numberLength += cardMaker.getStringWidth(elemAndCost.Elem, font, float64(cardMaker.Config.CostFontSize))
	}
	categoryLength := len(allCosts) * cardMaker.Config.CostCategoryWidth
	totalLength := numberLength + categoryLength + len(allCosts)*cardMaker.Config.CostPadding*2 + cardMaker.Config.CostPadding
	rectTop := cardMaker.Config.CostRectTop
	rectLeft := cardMaker.Config.CostRectLeft
	rectRight := rectLeft + totalLength
	rectBottom := rectTop + cardMaker.Config.CostRectHeight

	baseImage = cardMaker.drawRoundedRect(
		baseImage,
		rectLeft,
		rectTop,
		rectRight,
		rectBottom,
		cardMaker.Config.CostRectRadius,
		cardMaker.Config.CostRectOutlineWidth,
		cardMaker.getRectFillColor(cardInfo),
		cardMaker.getRectOutlineColor(cardInfo),
	)
	leftPointer := rectLeft + cardMaker.Config.CostPadding
	textHeight := cardMaker.getStringHeight(font, float64(cardMaker.Config.CostFontSize))
	textTop := rectTop + (cardMaker.Config.CostRectHeight-textHeight)/2 - cardMaker.Config.CostFontCompensation
	categoryTop := rectTop + (cardMaker.Config.CostRectHeight-cardMaker.Config.CostCategoryWidth)/2

	// sort all the costs, put the corersponding element to front
	for i, elemAndCost := range allCosts {
		if elemAndCost.Elem == cardInfo.Category {
			// remove this
			allCosts = append(allCosts[:i], allCosts[i+1:]...)
			// add to front
			allCosts = append([]ElemAndVal{elemAndCost}, allCosts...)
			break
		}
	}
	// draw the elements
	for _, tup := range allCosts {
		baseImage = cardMaker.addTextToImage(
			baseImage,
			strconv.Itoa(tup.Val),
			leftPointer,
			textTop,
			font,
			float64(cardMaker.Config.CostFontSize),
			cardMaker.Config.CostFontColor,
		)
		leftPointer += cardMaker.getStringWidth(strconv.Itoa(tup.Val), font, float64(cardMaker.Config.CostFontSize)) + cardMaker.Config.CostPadding
		// draw the category
		categoryImage, err := cardMaker.getCategoryImage(tup.Elem)
		if err != nil {
			return baseImage, err
		}
		categoryImage = cardMaker.adjustImage(categoryImage, cardMaker.Config.CostCategoryWidth, cardMaker.Config.CostCategoryWidth)
		baseImage = cardMaker.overlayImageOntoBase(
			baseImage,
			categoryImage,
			leftPointer,
			categoryTop,
		)
		leftPointer += cardMaker.Config.CostCategoryWidth + cardMaker.Config.CostPadding
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) GetCardType(cardInfo *CardInfo) string {
	dig := cardMaker.getDigit(cardInfo.Number, 0)
	if dig == 1 {
		return TYPE_UNIT_ZH
	} else if dig == 2 {
		return TYPE_ITEM_ZH
	} else if dig == 3 {
		return TYPE_ABILITY_ZH
	} else if dig == 4 {
		return TYPE_HERO_ZH
	} else {
		log.Fatalf("invalid card type: %v\n", cardInfo.Number)
		return TYPE_HERO_ZH
	}
}

func (cardMaker *CardMaker) drawTypeLogo(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	cardType := cardMaker.GetCardType(cardInfo)
	var typeLogo *image.RGBA
	var err error
	if cardType == TYPE_HERO_ZH {
		typeLogo, err = cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "hero_logo"))
	} else if cardType == TYPE_UNIT_ZH {
		typeLogo, err = cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "unit_logo"))
	} else if cardType == TYPE_ABILITY_ZH {
		typeLogo, err = cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "ability_logo"))
	} else if cardType == TYPE_ITEM_ZH {
		typeLogo, err = cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "item_logo"))
	} else {
		return baseImage, fmt.Errorf("invalid card type: %v", cardType)
	}
	if err != nil {
		return baseImage, err
	}
	typeLogo = cardMaker.adjustImage(typeLogo, cardMaker.Config.TypeLogoWidth, cardMaker.Config.TypeLogoWidth)

	baseImage = cardMaker.overlayImageOntoBase(
		baseImage,
		typeLogo,
		cardMaker.Config.TypeLogoLeft,
		cardMaker.Config.TypeLogoToBlockTop+cardMaker.Config.DrawingToUpper+cardMaker.Config.DrawingHeight,
	)
	return baseImage, nil
}

func (cardMaker *CardMaker) drawTag(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.TagFont))
	if err != nil {
		return baseImage, err
	}
	color := cardMaker.Config.TagFontColor
	if cardMaker.Config.ReverseColorForHero && cardMaker.GetCardType(cardInfo) == TYPE_HERO_ZH {
		color = cardMaker.reverseColor(color)
	}
	baseImage = cardMaker.addTextToImage(
		baseImage,
		cardInfo.Tag,
		cardMaker.Config.TagTextLeft,
		cardMaker.Config.TagTextToBlockTop+cardMaker.Config.DrawingToUpper+cardMaker.Config.DrawingHeight,
		font,
		float64(cardMaker.Config.TagFontSize),
		color,
	)
	return baseImage, nil
}

func (cardMaker *CardMaker) drawDescriptionAndQuote(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	// this function will dynamically adjust the font size
	descriptionFontSize := cardMaker.Config.DescriptionFontSize
	quoteFontSize := cardMaker.Config.QuoteFontSize
	descriptionLineSpacing := cardMaker.Config.DescriptionLineSpacing
	quoteLineSpacing := cardMaker.Config.QuoteLineSpacing

	descriptionFont, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.DescriptionFont))
	if err != nil {
		return baseImage, err
	}
	quoteFont, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.QuoteFont))
	if err != nil {
		return baseImage, err
	}
	estimatedTotalHeight := 0

	// estimate description height
	descriptionTextwrapWidthPixel := float64(cardMaker.Config.CardWidth - 2*cardMaker.Config.DescriptionTextLeft)
	descriptionTextwrapWidth := int(descriptionTextwrapWidthPixel)
	descriptionWrappedText, placeHolderLocations := cardMaker.textWrapAndPlaceHolderReplacement(cardInfo.Description, descriptionTextwrapWidth, descriptionFont, float64(descriptionFontSize))
	descriptionTextHeight := cardMaker.getStringHeight(descriptionFont, float64(descriptionFontSize))
	descriptionHeight := descriptionLineSpacing + (descriptionTextHeight+descriptionLineSpacing)*len(descriptionWrappedText)

	// extimate quote height
	quoteTextwrapWidthPixel := float64(cardMaker.Config.CardWidth - 2*cardMaker.Config.QuoteTextLeft)
	quoteTextwrapWidth := int(quoteTextwrapWidthPixel)
	quoteWrappedText, _ := cardMaker.textWrapAndPlaceHolderReplacement(cardInfo.Quote, quoteTextwrapWidth, quoteFont, float64(quoteFontSize))
	quoteTextHeight := cardMaker.getStringHeight(quoteFont, float64(quoteFontSize))
	quoteHeight := quoteLineSpacing + (quoteTextHeight+quoteLineSpacing)*len(quoteWrappedText)

	estimatedTotalHeight = descriptionHeight + quoteHeight

	for estimatedTotalHeight > cardMaker.Config.BottomBlockHeight-cardMaker.Config.DescriptionTextToBlockTop-cardMaker.Config.QuoteTextToBlockBottom {
		alpha := 0.9
		descriptionFontSize = int(float64(descriptionFontSize) * alpha)
		quoteFontSize = int(float64(quoteFontSize) * alpha)
		descriptionLineSpacing = int(float64(descriptionLineSpacing) * alpha)
		quoteLineSpacing = int(float64(quoteLineSpacing) * alpha)

		// estimate description height
		descriptionTextwrapWidthPixel = float64(cardMaker.Config.CardWidth - 2*cardMaker.Config.DescriptionTextLeft)
		descriptionTextwrapWidth = int(descriptionTextwrapWidthPixel)
		descriptionWrappedText, placeHolderLocations = cardMaker.textWrapAndPlaceHolderReplacement(cardInfo.Description, descriptionTextwrapWidth, descriptionFont, float64(descriptionFontSize))
		descriptionTextHeight = cardMaker.getStringHeight(descriptionFont, float64(descriptionFontSize))
		descriptionHeight = descriptionLineSpacing + (descriptionTextHeight+descriptionLineSpacing)*len(descriptionWrappedText)

		// extimate quote height
		quoteTextwrapWidthPixel = float64(cardMaker.Config.CardWidth - 2*cardMaker.Config.QuoteTextLeft)
		quoteTextwrapWidth = int(quoteTextwrapWidthPixel)
		quoteWrappedText, _ = cardMaker.textWrapAndPlaceHolderReplacement(cardInfo.Quote, quoteTextwrapWidth, quoteFont, float64(quoteFontSize))
		quoteTextHeight := cardMaker.getStringHeight(quoteFont, float64(quoteFontSize))
		quoteHeight := quoteLineSpacing + (quoteTextHeight+quoteLineSpacing)*len(quoteWrappedText)

		estimatedTotalHeight = descriptionHeight + quoteHeight

	}
	// start drawing
	// draw description

	discriptionTopPointer := cardMaker.Config.DescriptionTextToBlockTop + cardMaker.Config.DrawingToUpper + cardMaker.Config.DrawingHeight
	descriptionColor := cardMaker.Config.DescriptionFontColor
	if cardMaker.Config.ReverseColorForHero && cardMaker.GetCardType(cardInfo) == TYPE_HERO_ZH {
		descriptionColor = cardMaker.reverseColor(descriptionColor)
	}
	for _, line := range descriptionWrappedText {
		baseImage = cardMaker.addTextToImage(
			baseImage,
			line,
			cardMaker.Config.DescriptionTextLeft,
			discriptionTopPointer,
			descriptionFont,
			float64(descriptionFontSize),
			descriptionColor,
		)
		discriptionTopPointer += descriptionTextHeight + descriptionLineSpacing
	}
	// draw placeholders in description
	for _, placeholder := range placeHolderLocations {
		placeholderName, row, col := placeholder.Category, placeholder.Row, placeholder.Offset
		placeholderImage, err := cardMaker.getImageWithoutExtension(cardMaker.Config.PlacdholderToImage[placeholderName])
		if err != nil {
			return baseImage, err
		}
		w := cardMaker.getStringWidth(cardMaker.DefaultPlaceholder, descriptionFont, float64(descriptionFontSize))
		h := cardMaker.getStringHeight(descriptionFont, float64(descriptionFontSize))
		imageSquareWidth := min(w, h) - 2
		leftCompensation := (max(w, h)-imageSquareWidth)/2 - 2
		x := cardMaker.Config.DescriptionTextLeft + col + leftCompensation
		y := cardMaker.Config.DescriptionTextToBlockTop + cardMaker.Config.DrawingToUpper + cardMaker.Config.DrawingHeight + row*(descriptionTextHeight+descriptionLineSpacing) + 3
		// same as font size
		placeholderImage = cardMaker.adjustImage(placeholderImage, imageSquareWidth, imageSquareWidth)
		baseImage = cardMaker.overlayImageOntoBase(baseImage, placeholderImage, x, y)
	}
	// draw quote
	quoteBottomPointer := cardMaker.Config.BottomBlockHeight + cardMaker.Config.DrawingToUpper + cardMaker.Config.DrawingHeight - cardMaker.Config.QuoteTextToBlockBottom - quoteTextHeight
	quoteColor := cardMaker.Config.QuoteFontColor
	if cardMaker.Config.ReverseColorForHero && cardMaker.GetCardType(cardInfo) == TYPE_HERO_ZH {
		quoteColor = cardMaker.reverseColor(quoteColor)
	}
	slices.Reverse(quoteWrappedText)
	for _, line := range quoteWrappedText {
		baseImage = cardMaker.addTextToImage(
			baseImage,
			line,
			(cardMaker.Config.CardWidth-cardMaker.getStringWidth(line, quoteFont, float64(quoteFontSize)))/2,
			quoteBottomPointer,
			quoteFont,
			float64(quoteFontSize),
			quoteColor,
		)
		quoteBottomPointer -= quoteTextHeight + quoteLineSpacing
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) drawGain(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	allGains := []ElemAndVal{}
	for ele, val := range cardInfo.ElementsGain {
		if val > 0 {
			allGains = append(allGains, ElemAndVal{Elem: ele, Val: val})
		}
	}
	if len(allGains) == 0 {
		return baseImage, nil
	}
	font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.GainFont))
	if err != nil {
		return baseImage, err
	}
	numberLength := 0
	for _, elemAndCost := range allGains {
		numberLength += cardMaker.getStringWidth(elemAndCost.Elem, font, float64(cardMaker.Config.GainFontSize))
	}
	categoryLength := len(allGains) * cardMaker.Config.GainCategoryWidth
	totalLength := numberLength + categoryLength + len(allGains)*cardMaker.Config.GainPadding*2 + cardMaker.Config.GainPadding
	rectTop := cardMaker.Config.GainRectTop
	rectRight := cardMaker.Config.GainRectRight
	rectLeft := rectRight - totalLength
	rectBottom := rectTop + cardMaker.Config.GainRectHeight
	baseImage = cardMaker.drawRoundedRect(
		baseImage,
		rectLeft,
		rectTop,
		rectRight,
		rectBottom,
		cardMaker.Config.GainRectRadius,
		cardMaker.Config.GainRectOutlineWidth,
		cardMaker.getRectFillColor(cardInfo),
		cardMaker.getRectOutlineColor(cardInfo),
	)
	rightPointer := rectRight - cardMaker.Config.GainPadding - cardMaker.Config.GainCategoryWidth
	textHeight := cardMaker.getStringHeight(font, float64(cardMaker.Config.GainFontSize))
	textTop := rectTop + (cardMaker.Config.GainRectHeight-textHeight)/2 - cardMaker.Config.GainFontCompensation
	categoryTop := rectTop + (cardMaker.Config.GainRectHeight-cardMaker.Config.GainCategoryWidth)/2
	// sort all the gains, put the corresponding element to the head
	for i, elemAndCost := range allGains {
		if elemAndCost.Elem == cardInfo.Category {
			// remove this
			allGains = append(allGains[:i], allGains[i+1:]...)
			// add to front
			allGains = append([]ElemAndVal{elemAndCost}, allGains...)
			break
		}
	}
	// reverse the order
	slices.Reverse(allGains)
	for _, tup := range allGains {
		// draw the category
		categoryImage, err := cardMaker.getCategoryImage(tup.Elem)
		if err != nil {
			return baseImage, err
		}
		categoryImage = cardMaker.adjustImage(categoryImage, cardMaker.Config.GainCategoryWidth, cardMaker.Config.GainCategoryWidth)
		baseImage = cardMaker.overlayImageOntoBase(
			baseImage,
			categoryImage,
			rightPointer,
			categoryTop,
		)
		rightPointer -= cardMaker.getStringWidth(strconv.Itoa(tup.Val), font, float64(cardMaker.Config.GainFontSize)) + cardMaker.Config.GainPadding
		// draw the number
		baseImage = cardMaker.addTextToImage(
			baseImage,
			strconv.Itoa(tup.Val),
			rightPointer,
			textTop,
			font,
			float64(cardMaker.Config.GainFontSize),
			cardMaker.Config.GainFontColor,
		)
		rightPointer -= cardMaker.Config.GainCategoryWidth + cardMaker.Config.GainPadding
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) getAttackImage() (*image.RGBA, error) {
	return cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "attack"))
}
func (cardMaker *CardMaker) getLifeImage() (*image.RGBA, error) {
	return cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "life"))
}

func (cardMaker *CardMaker) drawLifeAndAttack(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	leftPointer := 0
	if cardInfo.Life >= 0 {
		lifeImage, err := cardMaker.getLifeImage()
		if err != nil {
			return baseImage, err
		}
		lifeImage = cardMaker.adjustImage(lifeImage, cardMaker.Config.LifeIconWidth, cardMaker.Config.LifeIconWidth)
		font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.LifeFont))
		if err != nil {
			return baseImage, err
		}
		estimatedLength := cardMaker.getStringWidth(strconv.Itoa(cardInfo.Life), font, float64(cardMaker.Config.LifeFontSize)) + cardMaker.Config.LifePadding*3 + cardMaker.Config.LifeIconWidth
		left := cardMaker.Config.LifeRectLeft
		top := cardMaker.Config.LifeRectTop
		right := left + estimatedLength
		bottom := top + cardMaker.Config.LifeRectHeight
		baseImage = cardMaker.drawRoundedRect(
			baseImage,
			left,
			top,
			right,
			bottom,
			cardMaker.Config.LifeRectRadius,
			cardMaker.Config.LifeRectOutlineWidth,
			cardMaker.getRectFillColor(cardInfo),
			cardMaker.getRectOutlineColor(cardInfo),
		)
		leftPointer = left + cardMaker.Config.LifePadding
		lifeTop := top + (cardMaker.Config.LifeRectHeight-cardMaker.Config.LifeIconWidth)/2
		baseImage = cardMaker.overlayImageOntoBase(
			baseImage,
			lifeImage,
			leftPointer,
			lifeTop,
		)
		leftPointer += cardMaker.Config.LifeIconWidth + cardMaker.Config.LifePadding
		lifeTextTop := top + (cardMaker.Config.LifeRectHeight-cardMaker.getStringHeight(font, float64(cardMaker.Config.LifeFontSize)))/2 - cardMaker.Config.LifeFontCompensation
		baseImage = cardMaker.addTextToImage(
			baseImage,
			strconv.Itoa(cardInfo.Life),
			leftPointer,
			lifeTextTop,
			font,
			float64(cardMaker.Config.LifeFontSize),
			cardMaker.Config.LifeFontColor,
		)
	}
	if cardInfo.Attack < 0 {
		return baseImage, nil
	} else {
		attackImage, err := cardMaker.getAttackImage()
		if err != nil {
			return baseImage, err
		}
		attackImage = cardMaker.adjustImage(attackImage, cardMaker.Config.AttackIconWidth, cardMaker.Config.AttackIconWidth)
		font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.AttackFont))
		if err != nil {
			return baseImage, err
		}
		estimatedLength := cardMaker.getStringWidth(strconv.Itoa(cardInfo.Attack), font, float64(cardMaker.Config.AttackFontSize)) + cardMaker.Config.AttackPadding*3 + cardMaker.Config.AttackIconWidth
		if cardInfo.Life < 0 {
			leftPointer = cardMaker.Config.LifeRectLeft
		} else {
			leftPointer += cardMaker.Config.LifeRectLeft + cardMaker.Config.LifePadding + cardMaker.getStringWidth(strconv.Itoa(cardInfo.Life), font, float64(cardMaker.Config.LifeFontSize))
		}
		left := leftPointer
		top := cardMaker.Config.AttackRectTop
		right := left + estimatedLength
		bottom := top + cardMaker.Config.AttackRectHeight
		baseImage = cardMaker.drawRoundedRect(
			baseImage,
			left,
			top,
			right,
			bottom,
			cardMaker.Config.AttackRectRadius,
			cardMaker.Config.AttackRectOutlineWidth,
			cardMaker.getRectFillColor(cardInfo),
			cardMaker.getRectOutlineColor(cardInfo),
		)
		leftPointer = left + cardMaker.Config.AttackPadding
		attackTop := top + (cardMaker.Config.AttackRectHeight-cardMaker.Config.AttackIconWidth)/2
		baseImage = cardMaker.overlayImageOntoBase(
			baseImage,
			attackImage,
			leftPointer,
			attackTop,
		)
		leftPointer += cardMaker.Config.AttackIconWidth + cardMaker.Config.AttackPadding
		attackTextTop := top + (cardMaker.Config.AttackRectHeight-cardMaker.getStringHeight(font, float64(cardMaker.Config.AttackFontSize)))/2 - cardMaker.Config.AttackFontCompensation
		baseImage = cardMaker.addTextToImage(
			baseImage,
			strconv.Itoa(cardInfo.Attack),
			leftPointer,
			attackTextTop,
			font,
			float64(cardMaker.Config.AttackFontSize),
			cardMaker.Config.AttackFontColor,
		)
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) getPowerImage() (*image.RGBA, error) {
	return cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "power"))
}

func (cardMaker *CardMaker) getDurationImage() (*image.RGBA, error) {
	return cardMaker.getImageWithoutExtension(filepath.Join(cardMaker.Config.GeneralPath, "duration"))
}

func (cardMaker *CardMaker) drawPowerOrDuration(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	var image *image.RGBA
	var text string
	var err error
	if cardInfo.Duration < 0 && cardInfo.Power < 0 {
		return baseImage, nil
	}
	if cardInfo.Duration >= 0 {
		image, err = cardMaker.getDurationImage()
		if err != nil {
			return baseImage, err
		}
		text = strconv.Itoa(cardInfo.Duration)
	} else {
		image, err = cardMaker.getPowerImage()
		if err != nil {
			return baseImage, err
		}
		text = strconv.Itoa(cardInfo.Power)
	}
	image = cardMaker.adjustImage(image, cardMaker.Config.PowerOrDurationIconWidth, cardMaker.Config.PowerOrDurationIconWidth)
	font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.PowerOrDurationFont))
	if err != nil {
		return baseImage, err
	}
	estimatedLength := cardMaker.getStringWidth(text, font, float64(cardMaker.Config.PowerOrDurationFontSize)) + cardMaker.Config.PowerOrDurationPadding*3 + cardMaker.Config.PowerOrDurationIconWidth
	right := cardMaker.Config.PowerOrDurationRectRight
	top := cardMaker.Config.PowerOrDurationRectTop
	left := right - estimatedLength
	bottom := top + cardMaker.Config.PowerOrDurationRectHeight
	baseImage = cardMaker.drawRoundedRect(
		baseImage,
		left,
		top,
		right,
		bottom,
		cardMaker.Config.PowerOrDurationRectRadius,
		cardMaker.Config.PowerOrDurationRectOutlineWidth,
		cardMaker.getRectFillColor(cardInfo),
		cardMaker.getRectOutlineColor(cardInfo),
	)
	rightPointer := right - cardMaker.Config.PowerOrDurationPadding - cardMaker.getStringWidth(text, font, float64(cardMaker.Config.PowerOrDurationFontSize))
	powerOrDurationTextTop := top + (cardMaker.Config.PowerOrDurationRectHeight-cardMaker.getStringHeight(font, float64(cardMaker.Config.PowerOrDurationFontSize)))/2 - cardMaker.Config.PowerOrDurationFontCompensation
	baseImage = cardMaker.addTextToImage(
		baseImage,
		text,
		rightPointer,
		powerOrDurationTextTop,
		font,
		float64(cardMaker.Config.PowerOrDurationFontSize),
		cardMaker.Config.PowerOrDurationFontColor,
	)
	rightPointer -= cardMaker.Config.LifeIconWidth + cardMaker.Config.LifePadding
	powerOrDurationTop := top + (cardMaker.Config.PowerOrDurationRectHeight-cardMaker.Config.PowerOrDurationIconWidth)/2
	baseImage = cardMaker.overlayImageOntoBase(
		baseImage,
		image,
		rightPointer,
		powerOrDurationTop,
	)
	return baseImage, nil
}

func (cardMaker *CardMaker) drawExpense(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	allExpenses := []ElemAndVal{}
	for ele, val := range cardInfo.ElementsExpense {
		if val > 0 {
			allExpenses = append(allExpenses, ElemAndVal{Elem: ele, Val: val})
		}
	}
	if len(allExpenses) == 0 {
		return baseImage, nil
	}
	font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.ExpenseFont))
	if err != nil {
		return baseImage, err
	}
	numberLength := 0
	for _, elemAndCost := range allExpenses {
		numberLength += cardMaker.getStringWidth(elemAndCost.Elem, font, float64(cardMaker.Config.ExpenseFontSize))
	}
	categoryLength := len(allExpenses) * cardMaker.Config.ExpenseCategoryWidth
	totalLength := numberLength + categoryLength + len(allExpenses)*cardMaker.Config.ExpensePadding*2 + cardMaker.Config.ExpensePadding
	rectTop := cardMaker.Config.ExpenseRectTop
	rectRight := cardMaker.Config.ExpenseRectRight
	rectLeft := rectRight - totalLength
	rectBottom := rectTop + cardMaker.Config.ExpenseRectHeight
	baseImage = cardMaker.drawRoundedRect(
		baseImage,
		rectLeft,
		rectTop,
		rectRight,
		rectBottom,
		cardMaker.Config.ExpenseRectRadius,
		cardMaker.Config.ExpenseRectOutlineWidth,
		cardMaker.getRectFillColor(cardInfo),
		cardMaker.getRectOutlineColor(cardInfo),
	)
	rightPointer := rectRight - cardMaker.Config.ExpensePadding - cardMaker.Config.ExpenseCategoryWidth
	textHeight := cardMaker.getStringHeight(font, float64(cardMaker.Config.ExpenseFontSize))
	textTop := rectTop + (cardMaker.Config.ExpenseRectHeight-textHeight)/2 - cardMaker.Config.ExpenseFontCompensation
	categoryTop := rectTop + (cardMaker.Config.ExpenseRectHeight-cardMaker.Config.ExpenseCategoryWidth)/2
	// sort all the gains, put the corresponding element to the head
	for i, elemAndCost := range allExpenses {
		if elemAndCost.Elem == cardInfo.Category {
			// remove this
			allExpenses = append(allExpenses[:i], allExpenses[i+1:]...)
			// add to front
			allExpenses = append([]ElemAndVal{elemAndCost}, allExpenses...)
			break
		}
	}
	// reverse the order
	slices.Reverse(allExpenses)
	for _, tup := range allExpenses {
		// draw the category
		categoryImage, err := cardMaker.getCategoryImage(tup.Elem)
		if err != nil {
			return baseImage, err
		}
		categoryImage = cardMaker.adjustImage(categoryImage, cardMaker.Config.ExpenseCategoryWidth, cardMaker.Config.ExpenseCategoryWidth)
		baseImage = cardMaker.overlayImageOntoBase(
			baseImage,
			categoryImage,
			rightPointer,
			categoryTop,
		)
		rightPointer -= cardMaker.getStringWidth(strconv.Itoa(tup.Val), font, float64(cardMaker.Config.ExpenseFontSize)) + cardMaker.Config.ExpensePadding
		// draw the number
		baseImage = cardMaker.addTextToImage(
			baseImage,
			strconv.Itoa(tup.Val),
			rightPointer,
			textTop,
			font,
			float64(cardMaker.Config.ExpenseFontSize),
			cardMaker.Config.ExpenseFontColor,
		)
		rightPointer -= cardMaker.Config.ExpenseCategoryWidth + cardMaker.Config.ExpensePadding
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) drawNumber(baseImage *image.RGBA, cardInfo *CardInfo) (*image.RGBA, error) {
	font, err := cardMaker.loadFont(filepath.Join(cardMaker.Config.FontPath, cardMaker.Config.NumberFont))
	if err != nil {
		return baseImage, err
	}
	color := cardMaker.Config.NumberFontColor
	if cardMaker.Config.ReverseColorForHero && cardMaker.GetCardType(cardInfo) == TYPE_HERO_ZH {
		color = cardMaker.reverseColor(color)
	}
	baseImage = cardMaker.addTextToImage(
		baseImage,
		"No."+cardInfo.Number,
		cardMaker.Config.CardWidth-cardMaker.Config.NumberTextToRight-cardMaker.getStringWidth("No."+cardInfo.Number, font, float64(cardMaker.Config.NumberFontSize)),
		cardMaker.Config.DrawingToUpper+cardMaker.Config.DrawingHeight+cardMaker.Config.NumberTextToBlockTop,
		font,
		float64(cardMaker.Config.NumberFontSize),
		color,
	)
	return baseImage, nil
}
func (cardMaker *CardMaker) makeUnitCard(cardInfo *CardInfo) (*image.RGBA, error) {
	baseImage, err := cardMaker.prepareOutline(cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCategoryAndName(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCost(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTypeLogo(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTag(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawDescriptionAndQuote(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawGain(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawLifeAndAttack(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawNumber(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) makeAbilityCard(cardInfo *CardInfo) (*image.RGBA, error) {
	baseImage, err := cardMaker.prepareOutline(cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCategoryAndName(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCost(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawExpense(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTypeLogo(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTag(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawDescriptionAndQuote(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawPowerOrDuration(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawNumber(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawLifeAndAttack(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) makeItemCard(cardInfo *CardInfo) (*image.RGBA, error) {
	baseImage, err := cardMaker.prepareOutline(cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCategoryAndName(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCost(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawExpense(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTypeLogo(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTag(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawDescriptionAndQuote(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawGain(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawPowerOrDuration(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawNumber(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawLifeAndAttack(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) makeHeroCard(cardInfo *CardInfo) (*image.RGBA, error) {
	baseImage, err := cardMaker.prepareOutline(cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCategoryAndName(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawCost(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTypeLogo(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawTag(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawDescriptionAndQuote(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawGain(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawLifeAndAttack(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	baseImage, err = cardMaker.drawNumber(baseImage, cardInfo)
	if err != nil {
		return nil, err
	}
	return baseImage, nil
}

func (cardMaker *CardMaker) MakeCard(cardInfo *CardInfo) (*image.RGBA, error) {
	cardType := cardMaker.GetCardType(cardInfo)
	if cardType == TYPE_UNIT_ZH {
		return cardMaker.makeUnitCard(cardInfo)
	} else if cardType == TYPE_ABILITY_ZH {
		return cardMaker.makeAbilityCard(cardInfo)
	} else if cardType == TYPE_ITEM_ZH {
		return cardMaker.makeItemCard(cardInfo)
	} else if cardType == TYPE_HERO_ZH {
		return cardMaker.makeHeroCard(cardInfo)
	}
	return nil, fmt.Errorf("unknown card type %s", cardType)
}

func (cardMaker *CardMaker) SaveImage(img *image.RGBA, path string) error {
	extensionName := filepath.Ext(path)
	if extensionName == ".jpeg" || extensionName == ".jpg" {
		outputFile, err := os.Create(path)
		if err != nil {
			return err
		}
		defer outputFile.Close()
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 100})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported extension name %s", extensionName)
	}
	return nil
}
