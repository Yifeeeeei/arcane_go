package card_maker

import (
	"image/color"
	"path/filepath"
)

type Config struct {
	// is this is enabled, card maker will only use the image pointed by the path as drawing
	StaticDrawing     bool   `json:"static_drawing"`
	StaticDrawingPath string `json:"static_drawing_path"`
	// path
	GeneralPath        string            `json:"general_path"`
	DrawingPath        string            `json:"drawing_path"`
	FontPath           string            `json:"font_path"`
	ElementImages      map[string]string `json:"element_images"`
	ElementBack        map[string]string `json:"element_back"`
	PlacdholderToImage map[string]string `json:"placeholder_to_image"`
	// card
	CardWidth  int `json:"card_width"`
	CardHeight int `json:"card_height"`
	// drawing
	DrawingWidth   int `json:"drawing_width"`
	DrawingHeight  int `json:"drawing_height"`
	DrawingToUpper int `json:"drawing_to_upper"`
	// border
	BorderWidth  int               `json:"border_width"`
	BorderHeight int               `json:"border_height"`
	TypeBorder   map[string]string `json:"type_border"`
	// hero
	ReverseColorForHero bool `json:"reverse_color_for_hero"`
	// bottom block
	BottomBlockWidth        int        `json:"bottom_block_width"`
	BottomBlockHeight       int        `json:"bottom_block_height"`
	BottomBlockColor        color.RGBA `json:"bottom_block_color"`
	BottomBlockLegendRadius int        `json:"bottom_block_legend_radius"`
	BottomBlockTransparency int        `json:"bottom_block_transparency"`
	// upper left: element + name
	NameFontSize             int        `json:"name_font_size"`
	NameFont                 string     `json:"name_font"`
	NameRectHeight           int        `json:"name_rect_height"`
	NameRectLeft             int        `json:"name_rect_left"`
	NameRectTop              int        `json:"name_rect_top"`
	NameRectRadius           int        `json:"name_rect_radius"`
	NameTextToLeft           int        `json:"name_text_to_left"`
	NameTextLeftCompensation int        `json:"name_text_left_compensation"`
	NameTextTopCompensation  int        `json:"name_text_top_compensation"`
	NameRectOutlineWidth     int        `json:"name_rect_outline_width"`
	NameTextFontColor        color.RGBA `json:"name_text_font_color"`
	NameCategoryWidth        int        `json:"name_category_width"`
	NameCategoryLeft         int        `json:"name_category_left"`
	NameCategoryTop          int        `json:"name_category_top"`
	// middle: cost
	CostFontSize         int        `json:"cost_font_size"`
	CostFont             string     `json:"cost_font"`
	CostCategoryWidth    int        `json:"cost_category_width"`
	CostFontCompensation int        `json:"cost_font_compensation"`
	CostFontColor        color.RGBA `json:"cost_font_color"`
	CostPadding          int        `json:"cost_padding"`
	CostRectTop          int        `json:"cost_rect_top"`
	CostRectLeft         int        `json:"cost_rect_left"`
	CostRectHeight       int        `json:"cost_rect_height"`
	CostRectRadius       int        `json:"cost_rect_radius"`
	CostRectOutlineWidth int        `json:"cost_rect_outline_width"`
	// middle: expence
	ExpenseFontSize         int        `json:"expense_font_size"`
	ExpenseFont             string     `json:"expense_font"`
	ExpenseCategoryWidth    int        `json:"expense_category_width"`
	ExpenseFontCompensation int        `json:"expense_font_compensation"`
	ExpenseFontColor        color.RGBA `json:"expense_font_color"`
	ExpensePadding          int        `json:"expense_padding"`
	ExpenseRectTop          int        `json:"expense_rect_top"`
	ExpenseRectRight        int        `json:"expense_rect_right"`
	ExpenseRectHeight       int        `json:"expense_rect_height"`
	ExpenseRectRadius       int        `json:"expense_rect_radius"`
	ExpenseRectOutlineWidth int        `json:"expense_rect_outline_width"`
	// type
	TypeLogoWidth      int `json:"type_logo_width"`
	TypeLogoLeft       int `json:"type_logo_left"`
	TypeLogoToBlockTop int `json:"type_logo_to_block_top"`
	// tag
	TagFont           string     `json:"tag_font"`
	TagFontSize       int        `json:"tag_font_size"`
	TagFontColor      color.RGBA `json:"tag_font_color"`
	TagTextLeft       int        `json:"tag_text_left"`
	TagTextToBlockTop int        `json:"tag_text_to_block_top"`
	// description
	DescriptionFont           string     `json:"description_font"`
	DescriptionFontSize       int        `json:"description_font_size"`
	DescriptionFontColor      color.RGBA `json:"description_font_color"`
	DescriptionTextLeft       int        `json:"description_text_left"`
	DescriptionTextToBlockTop int        `json:"description_text_to_block_top"`
	DescriptionLineSpacing    int        `json:"description_line_spacing"`
	// quote
	QuoteFont              string     `json:"quote_font"`
	QuoteFontSize          int        `json:"quote_font_size"`
	QuoteFontColor         color.RGBA `json:"quote_font_color"`
	QuoteTextLeft          int        `json:"quote_text_left"`
	QuoteTextToBlockBottom int        `json:"quote_text_to_block_bottom"`
	QuoteLineSpacing       int        `json:"quote_line_spacing"`
	// bottom: gain
	GainFontSize         int `json:"gain_font_size"`
	GainFont             string
	GainCategoryWidth    int        `json:"gain_category_width"`
	GainFontCompensation int        `json:"gain_font_compensation"`
	GainFontColor        color.RGBA `json:"gain_font_color"`
	GainPadding          int        `json:"gain_padding"`
	GainRectTop          int        `json:"gain_rect_top"`
	GainRectRight        int        `json:"gain_rect_right"`
	GainRectHeight       int        `json:"gain_rect_height"`
	GainRectRadius       int        `json:"gain_rect_radius"`
	GainRectOutlineWidth int        `json:"gain_rect_outline_width"`
	// bottom: life
	LifeFontSize         int `json:"life_font_size"`
	LifeFont             string
	LifeIconWidth        int        `json:"life_icon_width"`
	LifeFontCompensation int        `json:"life_font_compensation"`
	LifeFontColor        color.RGBA `json:"life_font_color"`
	LifePadding          int        `json:"life_padding"`
	LifeRectTop          int        `json:"life_rect_top"`
	LifeRectLeft         int        `json:"life_rect_left"`
	LifeRectHeight       int        `json:"life_rect_height"`
	LifeRectRadius       int        `json:"life_rect_radius"`
	LifeRectOutlineWidth int        `json:"life_rect_outline_width"`

	// bottom: attack
	AttackFontSize         int        `json:"attack_font_size"`
	AttackFont             string     `json:"attack_font"`
	AttackIconWidth        int        `json:"attack_icon_width"`
	AttackFontCompensation int        `json:"attack_font_compensation"`
	AttackFontColor        color.RGBA `json:"attack_font_color"`
	AttackPadding          int        `json:"attack_padding"`
	AttackRectTop          int        `json:"attack_rect_top"`
	// AttackRectLeft         int        `json:"attack_rect_left"`
	AttackRectHeight       int `json:"attack_rect_height"`
	AttackRectRadius       int `json:"attack_rect_radius"`
	AttackRectOutlineWidth int `json:"attack_rect_outline_width"`
	// bottom: power or duration
	PowerOrDurationFontSize         int `json:"power_or_duration_font_size"`
	PowerOrDurationFont             string
	PowerOrDurationIconWidth        int        `json:"power_or_duration_icon_width"`
	PowerOrDurationFontCompensation int        `json:"power_or_duration_font_compensation"`
	PowerOrDurationFontColor        color.RGBA `json:"power_or_duration_font_color"`
	PowerOrDurationPadding          int        `json:"power_or_duration_padding"`
	PowerOrDurationRectTop          int        `json:"power_or_duration_rect_top"`
	PowerOrDurationRectRight        int        `json:"power_or_duration_rect_right"`
	PowerOrDurationRectHeight       int        `json:"power_or_duration_rect_height"`
	PowerOrDurationRectRadius       int        `json:"power_or_duration_rect_radius"`
	PowerOrDurationRectOutlineWidth int        `json:"power_or_duration_rect_outline_width"`
	// number
	NumberFontSize       int        `json:"number_font_size"`
	NumberFont           string     `json:"number_font"`
	NumberFontColor      color.RGBA `json:"number_font_color"`
	NumberTextToRight    int        `json:"number_text_to_right"`
	NumberTextToBlockTop int        `json:"number_text_to_block_top"`
}

func (c *Config) Copy() *Config {
	newConfig := &Config{
		StaticDrawing:                   c.StaticDrawing,
		StaticDrawingPath:               c.StaticDrawingPath,
		GeneralPath:                     c.GeneralPath,
		DrawingPath:                     c.DrawingPath,
		FontPath:                        c.FontPath,
		ElementImages:                   c.ElementImages,
		ElementBack:                     c.ElementBack,
		PlacdholderToImage:              c.PlacdholderToImage,
		CardWidth:                       c.CardWidth,
		CardHeight:                      c.CardHeight,
		DrawingWidth:                    c.DrawingWidth,
		DrawingHeight:                   c.DrawingHeight,
		DrawingToUpper:                  c.DrawingToUpper,
		BorderWidth:                     c.BorderWidth,
		BorderHeight:                    c.BorderHeight,
		TypeBorder:                      c.TypeBorder,
		ReverseColorForHero:             c.ReverseColorForHero,
		BottomBlockWidth:                c.BottomBlockWidth,
		BottomBlockHeight:               c.BottomBlockHeight,
		BottomBlockColor:                c.BottomBlockColor,
		BottomBlockLegendRadius:         c.BottomBlockLegendRadius,
		BottomBlockTransparency:         c.BottomBlockTransparency,
		NameFontSize:                    c.NameFontSize,
		NameFont:                        c.NameFont,
		NameRectHeight:                  c.NameRectHeight,
		NameRectLeft:                    c.NameRectLeft,
		NameRectTop:                     c.NameRectTop,
		NameRectRadius:                  c.NameRectRadius,
		NameTextToLeft:                  c.NameTextToLeft,
		NameTextLeftCompensation:        c.NameTextLeftCompensation,
		NameTextTopCompensation:         c.NameTextTopCompensation,
		NameRectOutlineWidth:            c.NameRectOutlineWidth,
		NameTextFontColor:               c.NameTextFontColor,
		NameCategoryWidth:               c.NameCategoryWidth,
		NameCategoryLeft:                c.NameCategoryLeft,
		NameCategoryTop:                 c.NameCategoryTop,
		CostFontSize:                    c.CostFontSize,
		CostFont:                        c.CostFont,
		CostCategoryWidth:               c.CostCategoryWidth,
		CostFontCompensation:            c.CostFontCompensation,
		CostFontColor:                   c.CostFontColor,
		CostPadding:                     c.CostPadding,
		CostRectTop:                     c.CostRectTop,
		CostRectLeft:                    c.CostRectLeft,
		CostRectHeight:                  c.CostRectHeight,
		CostRectRadius:                  c.CostRectRadius,
		CostRectOutlineWidth:            c.CostRectOutlineWidth,
		ExpenseFontSize:                 c.ExpenseFontSize,
		ExpenseFont:                     c.ExpenseFont,
		ExpenseCategoryWidth:            c.ExpenseCategoryWidth,
		ExpenseFontCompensation:         c.ExpenseFontCompensation,
		ExpenseFontColor:                c.ExpenseFontColor,
		ExpensePadding:                  c.ExpensePadding,
		ExpenseRectTop:                  c.ExpenseRectTop,
		ExpenseRectRight:                c.ExpenseRectRight,
		ExpenseRectHeight:               c.ExpenseRectHeight,
		ExpenseRectRadius:               c.ExpenseRectRadius,
		ExpenseRectOutlineWidth:         c.ExpenseRectOutlineWidth,
		TypeLogoWidth:                   c.TypeLogoWidth,
		TypeLogoLeft:                    c.TypeLogoLeft,
		TypeLogoToBlockTop:              c.TypeLogoToBlockTop,
		TagFont:                         c.TagFont,
		TagFontSize:                     c.TagFontSize,
		TagFontColor:                    c.TagFontColor,
		TagTextLeft:                     c.TagTextLeft,
		TagTextToBlockTop:               c.TagTextToBlockTop,
		DescriptionFont:                 c.DescriptionFont,
		DescriptionFontSize:             c.DescriptionFontSize,
		DescriptionFontColor:            c.DescriptionFontColor,
		DescriptionTextLeft:             c.DescriptionTextLeft,
		DescriptionTextToBlockTop:       c.DescriptionTextToBlockTop,
		DescriptionLineSpacing:          c.DescriptionLineSpacing,
		QuoteFont:                       c.QuoteFont,
		QuoteFontSize:                   c.QuoteFontSize,
		QuoteFontColor:                  c.QuoteFontColor,
		QuoteTextLeft:                   c.QuoteTextLeft,
		QuoteTextToBlockBottom:          c.QuoteTextToBlockBottom,
		QuoteLineSpacing:                c.QuoteLineSpacing,
		GainFontSize:                    c.GainFontSize,
		GainFont:                        c.GainFont,
		GainCategoryWidth:               c.GainCategoryWidth,
		GainFontCompensation:            c.GainFontCompensation,
		GainFontColor:                   c.GainFontColor,
		GainPadding:                     c.GainPadding,
		GainRectTop:                     c.GainRectTop,
		GainRectRight:                   c.GainRectRight,
		GainRectHeight:                  c.GainRectHeight,
		GainRectRadius:                  c.GainRectRadius,
		GainRectOutlineWidth:            c.GainRectOutlineWidth,
		LifeFontSize:                    c.LifeFontSize,
		LifeFont:                        c.LifeFont,
		LifeIconWidth:                   c.LifeIconWidth,
		LifeFontCompensation:            c.LifeFontCompensation,
		LifeFontColor:                   c.LifeFontColor,
		LifePadding:                     c.LifePadding,
		LifeRectTop:                     c.LifeRectTop,
		LifeRectLeft:                    c.LifeRectLeft,
		LifeRectHeight:                  c.LifeRectHeight,
		LifeRectRadius:                  c.LifeRectRadius,
		LifeRectOutlineWidth:            c.LifeRectOutlineWidth,
		AttackFontSize:                  c.AttackFontSize,
		AttackFont:                      c.AttackFont,
		AttackIconWidth:                 c.AttackIconWidth,
		AttackFontCompensation:          c.AttackFontCompensation,
		AttackFontColor:                 c.AttackFontColor,
		AttackPadding:                   c.AttackPadding,
		AttackRectTop:                   c.AttackRectTop,
		AttackRectHeight:                c.AttackRectHeight,
		AttackRectRadius:                c.AttackRectRadius,
		AttackRectOutlineWidth:          c.AttackRectOutlineWidth,
		PowerOrDurationFontSize:         c.PowerOrDurationFontSize,
		PowerOrDurationFont:             c.PowerOrDurationFont,
		PowerOrDurationIconWidth:        c.PowerOrDurationIconWidth,
		PowerOrDurationFontCompensation: c.PowerOrDurationFontCompensation,
		PowerOrDurationFontColor:        c.PowerOrDurationFontColor,
		PowerOrDurationPadding:          c.PowerOrDurationPadding,
		PowerOrDurationRectTop:          c.PowerOrDurationRectTop,
		PowerOrDurationRectRight:        c.PowerOrDurationRectRight,
		PowerOrDurationRectHeight:       c.PowerOrDurationRectHeight,
		PowerOrDurationRectRadius:       c.PowerOrDurationRectRadius,
		PowerOrDurationRectOutlineWidth: c.PowerOrDurationRectOutlineWidth,
		NumberFontSize:                  c.NumberFontSize,
		NumberFont:                      c.NumberFont,
		NumberFontColor:                 c.NumberFontColor,
		NumberTextToRight:               c.NumberTextToRight,
		NumberTextToBlockTop:            c.NumberTextToBlockTop,
	}
	return newConfig

}

func NewDefaultConfig(sizeRatio int, generalPath, drawingPath, fontPath string) Config {

	return Config{
		StaticDrawing:     false,
		StaticDrawingPath: "",

		GeneralPath: generalPath,
		DrawingPath: drawingPath,
		FontPath:    fontPath,
		ElementImages: map[string]string{
			"光": "ele_light",
			"暗": "ele_dark",
			"水": "ele_water",
			"火": "ele_fire",
			"气": "ele_air",
			"地": "ele_earth",
			"无": "ele_none",
			"法": "ele_mana",
		},
		ElementBack: map[string]string{
			"光": "back_light",
			"暗": "back_dark",
			"水": "back_water",
			"火": "back_fire",
			"气": "back_air",
			"地": "back_earth",
			"无": "back_none",
		},
		PlacdholderToImage: map[string]string{
			"\\光": filepath.Join(generalPath, "ele_light"),
			"\\暗": filepath.Join(generalPath, "ele_dark"),
			"\\水": filepath.Join(generalPath, "ele_water"),
			"\\火": filepath.Join(generalPath, "ele_fire"),
			"\\气": filepath.Join(generalPath, "ele_air"),
			"\\地": filepath.Join(generalPath, "ele_earth"),
			"\\无": filepath.Join(generalPath, "ele_none"),
			"\\法": filepath.Join(generalPath, "ele_mana"),
			"\\血": filepath.Join(generalPath, "life"),
			"\\攻": filepath.Join(generalPath, "attack"),
			"\\威": filepath.Join(generalPath, "power"),
			"\\持": filepath.Join(generalPath, "duration"),
		},
		CardWidth:               590 * sizeRatio,
		CardHeight:              860 * sizeRatio,
		DrawingWidth:            540 * sizeRatio,
		DrawingHeight:           540 * sizeRatio,
		DrawingToUpper:          35 * sizeRatio,
		BorderWidth:             580 * sizeRatio,
		BorderHeight:            830 * sizeRatio,
		TypeBorder:              map[string]string{"伙伴": "border_unit", "技能": "border_ability", "道具": "border_item"},
		ReverseColorForHero:     true,
		BottomBlockWidth:        540 * sizeRatio,
		BottomBlockHeight:       260 * sizeRatio,
		BottomBlockColor:        color.RGBA{255, 255, 255, 155},
		BottomBlockLegendRadius: 0 * sizeRatio,

		NameFontSize:                    40 * sizeRatio,
		NameFont:                        "Pangmen.ttf",
		NameRectHeight:                  60 * sizeRatio,
		NameRectLeft:                    60 * sizeRatio,
		NameRectTop:                     10 * sizeRatio,
		NameRectRadius:                  10 * sizeRatio,
		NameTextToLeft:                  90 * sizeRatio,
		NameTextLeftCompensation:        5,
		NameTextTopCompensation:         0,
		NameRectOutlineWidth:            3 * sizeRatio,
		NameTextFontColor:               color.RGBA{0, 0, 0, 255},
		NameCategoryWidth:               80 * sizeRatio,
		NameCategoryLeft:                5 * sizeRatio,
		NameCategoryTop:                 5 * sizeRatio,
		CostFontSize:                    30 * sizeRatio,
		CostFont:                        "ShareTechMono-Regular.ttf",
		CostCategoryWidth:               30 * sizeRatio,
		CostFontCompensation:            2,
		CostFontColor:                   color.RGBA{0, 0, 0, 255},
		CostPadding:                     5 * sizeRatio,
		CostRectTop:                     530 * sizeRatio,
		CostRectLeft:                    10 * sizeRatio,
		CostRectHeight:                  50 * sizeRatio,
		CostRectRadius:                  25 * sizeRatio,
		CostRectOutlineWidth:            3 * sizeRatio,
		ExpenseFontSize:                 30 * sizeRatio,
		ExpenseFont:                     "ShareTechMono-Regular.ttf",
		ExpenseCategoryWidth:            30 * sizeRatio,
		ExpenseFontCompensation:         2,
		ExpenseFontColor:                color.RGBA{0, 0, 0, 255},
		ExpensePadding:                  5 * sizeRatio,
		ExpenseRectTop:                  530 * sizeRatio,
		ExpenseRectRight:                580 * sizeRatio,
		ExpenseRectHeight:               50 * sizeRatio,
		ExpenseRectRadius:               25 * sizeRatio,
		ExpenseRectOutlineWidth:         3 * sizeRatio,
		TypeLogoWidth:                   30 * sizeRatio,
		TypeLogoLeft:                    50 * sizeRatio,
		TypeLogoToBlockTop:              15 * sizeRatio,
		TagFont:                         "LXGWWenKaiMono-Regular.ttf",
		TagFontSize:                     24 * sizeRatio,
		TagFontColor:                    color.RGBA{0, 0, 0, 255},
		TagTextLeft:                     (50 + 40) * sizeRatio,
		TagTextToBlockTop:               15 * sizeRatio,
		DescriptionFont:                 "LXGWWenKaiMono-Regular.ttf",
		DescriptionFontSize:             24 * sizeRatio,
		DescriptionFontColor:            color.RGBA{0, 0, 0, 255},
		DescriptionTextLeft:             50 * sizeRatio,
		DescriptionTextToBlockTop:       55 * sizeRatio,
		DescriptionLineSpacing:          10 * sizeRatio,
		QuoteFont:                       "LXGWWenKaiMono-Light.ttf",
		QuoteFontSize:                   20 * sizeRatio,
		QuoteFontColor:                  color.RGBA{0, 0, 0, 255},
		QuoteTextLeft:                   100 * sizeRatio,
		QuoteTextToBlockBottom:          40 * sizeRatio,
		QuoteLineSpacing:                5 * sizeRatio,
		GainFontSize:                    30 * sizeRatio,
		GainFont:                        "ShareTechMono-Regular.ttf",
		GainCategoryWidth:               30 * sizeRatio,
		GainFontCompensation:            1,
		GainFontColor:                   color.RGBA{0, 0, 0, 255},
		GainPadding:                     5 * sizeRatio,
		GainRectTop:                     800 * sizeRatio,
		GainRectRight:                   580 * sizeRatio,
		GainRectHeight:                  50 * sizeRatio,
		GainRectRadius:                  25 * sizeRatio,
		GainRectOutlineWidth:            3 * sizeRatio,
		LifeFontSize:                    30 * sizeRatio,
		LifeFont:                        "ShareTechMono-Regular.ttf",
		LifeIconWidth:                   30 * sizeRatio,
		LifeFontCompensation:            2,
		LifeFontColor:                   color.RGBA{0, 0, 0, 255},
		LifePadding:                     5 * sizeRatio,
		LifeRectTop:                     800 * sizeRatio,
		LifeRectLeft:                    10 * sizeRatio,
		LifeRectHeight:                  50 * sizeRatio,
		LifeRectRadius:                  25 * sizeRatio,
		LifeRectOutlineWidth:            3 * sizeRatio,
		AttackFontSize:                  30 * sizeRatio,
		AttackFont:                      "ShareTechMono-Regular.ttf",
		AttackIconWidth:                 30 * sizeRatio,
		AttackFontCompensation:          2,
		AttackFontColor:                 color.RGBA{0, 0, 0, 255},
		AttackPadding:                   5 * sizeRatio,
		AttackRectTop:                   800 * sizeRatio,
		AttackRectHeight:                50 * sizeRatio,
		AttackRectRadius:                25 * sizeRatio,
		AttackRectOutlineWidth:          3 * sizeRatio,
		PowerOrDurationFontSize:         30 * sizeRatio,
		PowerOrDurationFont:             "ShareTechMono-Regular.ttf",
		PowerOrDurationIconWidth:        30 * sizeRatio,
		PowerOrDurationFontCompensation: 2,
		PowerOrDurationFontColor:        color.RGBA{0, 0, 0, 255},
		PowerOrDurationPadding:          5 * sizeRatio,
		PowerOrDurationRectTop:          800 * sizeRatio,
		PowerOrDurationRectRight:        580 * sizeRatio,
		PowerOrDurationRectHeight:       50 * sizeRatio,
		PowerOrDurationRectRadius:       25 * sizeRatio,
		PowerOrDurationRectOutlineWidth: 3 * sizeRatio,
		NumberFontSize:                  20 * sizeRatio,
		NumberFont:                      "ShareTechMono-Regular.ttf",
		NumberFontColor:                 color.RGBA{0, 0, 0, 255},
		NumberTextToRight:               50 * sizeRatio,
		NumberTextToBlockTop:            17 * sizeRatio,
	}
}
