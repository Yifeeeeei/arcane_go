package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Yifeeeeei/arcane_go/card_maker"
	"github.com/xuri/excelize/v2"
	"golang.org/x/image/tiff"
)

// IDs encode set, element, type, and a three-digit sequence:
// B + A/F/W/... + H/C/S/I + 001.
var newCardID = regexp.MustCompile(`^B[AFWEDLR][HCSI][0-9]{3}$`)

type Row map[string]string
type artworkCard struct {
	NewID     string  `json:"new_id"`
	LegacyID  *string `json:"legacy_id"`
	LegacySet *string `json:"legacy_set"`
}
type artworkMigration struct {
	Cards []artworkCard `json:"cards"`
}

type MassProducer struct {
	Params       *MassProducerParams
	AllCardInfos []card_maker.CardInfo
	Config       *card_maker.Config
	artwork      map[string]artworkCard
	artworkFiles map[string]string
}

func NewMassProducer(paramPath string) (*MassProducer, error) {
	b, err := os.ReadFile(paramPath)
	if err != nil {
		return nil, err
	}
	p := &MassProducerParams{}
	if err = json.Unmarshal(b, p); err != nil {
		return nil, err
	}
	if len(p.XlsxPaths) == 0 {
		return nil, fmt.Errorf("xlsx_paths cannot be empty")
	}
	if p.OutputPath == "" {
		p.OutputPath = "output"
	}
	if p.SizeRatio == 0 {
		p.SizeRatio = 2
	}
	cfg := card_maker.NewDefaultConfig(p.SizeRatio, p.GeneralPath, "", p.FontPath)
	m := &MassProducer{Params: p, Config: &cfg, artwork: map[string]artworkCard{}, artworkFiles: map[string]string{}}
	if p.ArtworkMigrationPath != "" {
		b, err = os.ReadFile(p.ArtworkMigrationPath)
		if err != nil {
			return nil, err
		}
		var migration artworkMigration
		if err = json.Unmarshal(b, &migration); err != nil {
			return nil, err
		}
		for _, a := range migration.Cards {
			m.artwork[a.NewID] = a
		}
	}
	return m, nil
}

func parseOptionalInt(raw, field string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil || f < 0 || f != float64(int(f)) {
		return nil, fmt.Errorf("%s must be a non-negative integer, got %q", field, raw)
	}
	v := int(f)
	return &v, nil
}

var elementAliases = map[string]string{"火": "火", "火焰": "火", "水": "水", "水纹": "水", "气": "气", "大气": "气", "地": "地", "地脉": "地", "光": "光", "光辉": "光", "暗": "暗", "暗影": "暗", "奥术": "奥术", "无": "奥术"}

func parseElements(raw string) (card_maker.Elements, error) {
	result := card_maker.Elements{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result, nil
	}
	normalized := strings.NewReplacer("；", " ", ";", " ", "，", " ", ",", " ", "/", " ", "\\", " ", "+", " ").Replace(raw)
	re := regexp.MustCompile(`([\p{Han}]+)\s*([0-9]+(?:\.0+)?)`)
	matches := re.FindAllStringSubmatch(normalized, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid element vector %q", raw)
	}
	consumed := re.ReplaceAllString(normalized, "")
	if strings.TrimSpace(consumed) != "" {
		return nil, fmt.Errorf("invalid element vector %q", raw)
	}
	for _, match := range matches {
		name, ok := elementAliases[match[1]]
		if !ok {
			return nil, fmt.Errorf("unknown element %q", match[1])
		}
		v, err := parseOptionalInt(match[2], "element value")
		if err != nil {
			return nil, err
		}
		result[name] += *v
	}
	return result, nil
}

func parseSheet(rows [][]string) ([]Row, error) {
	headerAt := -1
	for i, row := range rows {
		var hasID, hasName bool
		for _, c := range row {
			c = strings.TrimSpace(c)
			hasID = hasID || c == "卡号"
			hasName = hasName || c == "卡名"
		}
		if hasID && hasName {
			headerAt = i
			break
		}
	}
	if headerAt < 0 {
		return nil, fmt.Errorf("header row containing 卡号 and 卡名 not found")
	}
	header := rows[headerAt]
	out := []Row{}
	for _, cells := range rows[headerAt+1:] {
		r := Row{}
		nonempty := false
		for i, h := range header {
			if i >= len(cells) {
				continue
			}
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			r[h] = cells[i]
			nonempty = nonempty || strings.TrimSpace(cells[i]) != ""
		}
		if nonempty {
			out = append(out, r)
		}
	}
	return out, nil
}

func setCodeFromID(id string) string {
	if len(id) < 1 {
		return ""
	}
	return id[:1]
}

func (m *MassProducer) cardFromRow(row Row, sheet, setName string) (card_maker.CardInfo, error) {
	id := strings.TrimSpace(row["卡号"])
	if !newCardID.MatchString(id) {
		return card_maker.CardInfo{}, fmt.Errorf("invalid card id %q", id)
	}
	name := strings.TrimSpace(row["卡名"])
	if name == "" {
		return card_maker.CardInfo{}, fmt.Errorf("%s: 卡名 is required", id)
	}
	ci := card_maker.CardInfo{Number: id, SetCode: setCodeFromID(id), SetName: setName, Name: name, Category: strings.TrimSpace(row["属性"]), Subtype: strings.TrimSpace(row["子类型"]), Rarity: strings.TrimSpace(row["稀有度"]), Description: row["效果"], Quote: row["引言"], Source: sheet}
	var err error
	if sheet == "人物" {
		ci.Type = card_maker.TYPE_HERO_ZH
	} else {
		ci.Type = strings.TrimSpace(row["类型"])
		if ci.Type != card_maker.TYPE_UNIT_ZH && ci.Type != card_maker.TYPE_ABILITY_ZH && ci.Type != card_maker.TYPE_ITEM_ZH {
			return ci, fmt.Errorf("%s: invalid 类型 %q", id, ci.Type)
		}
	}
	if ci.ManaCost, err = parseOptionalInt(row["入场／学习／使用／盖放"], "mana cost"); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	activationRaw := row["发动法力"]
	if sheet == "人物" {
		activationRaw = row["主动法力"]
	}
	if ci.ActivationMana, err = parseOptionalInt(activationRaw, "activation mana"); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	if ci.AttackValue, err = parseOptionalInt(row["攻击"], "attack"); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	ci.Attack = -1
	if ci.AttackValue != nil {
		ci.Attack = *ci.AttackValue
	}
	if ci.LifeValue, err = parseOptionalInt(row["生命"], "life"); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	ci.Life = -1
	if ci.LifeValue != nil {
		ci.Life = *ci.LifeValue
	}
	if ci.Cooldown, err = parseOptionalInt(row["冷却"], "cooldown"); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	if ci.SkillSlots, err = parseOptionalInt(row["技能槽"], "skill slots"); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	if ci.TechRequirement, err = parseElements(row["负载条件"]); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	if ci.ElementsGain, err = parseElements(row["负载"]); err != nil {
		return ci, fmt.Errorf("%s: %w", id, err)
	}
	ci.ElementsCost = ci.TechRequirement
	ci.ElementsExpense = card_maker.Elements{}
	ci.Duration = -1
	if ci.Cooldown != nil {
		ci.Duration = *ci.Cooldown
	}
	ci.Power = -1
	ci.Tag = ci.Subtype
	ci.VersionName = ci.SetName
	ci.OutputPath = filepath.ToSlash(filepath.Join(m.Params.OutputPath, ci.SetName, ci.Type, ci.Category, ci.Number+".jpg"))
	if a, ok := m.artwork[id]; ok && a.LegacyID != nil {
		ci.ArtRef = *a.LegacyID
	}
	return ci, nil
}

func (m *MassProducer) LoadCards() error {
	var cards []card_maker.CardInfo
	seen := map[string]bool{}
	for i, path := range m.Params.XlsxPaths {
		book, err := excelize.OpenFile(path)
		if err != nil {
			return err
		}
		setName := "基础包"
		if i < len(m.Params.VersionNames) && m.Params.VersionNames[i] != "" {
			setName = m.Params.VersionNames[i]
		}
		for _, sheet := range book.GetSheetList() {
			rows, err := book.GetRows(sheet)
			if err != nil {
				book.Close()
				return err
			}
			parsed, err := parseSheet(rows)
			if err != nil {
				continue
			}
			for _, row := range parsed {
				ci, err := m.cardFromRow(row, sheet, setName)
				if err != nil {
					book.Close()
					return err
				}
				if seen[ci.Number] {
					book.Close()
					return fmt.Errorf("duplicate card id %s", ci.Number)
				}
				seen[ci.Number] = true
				cards = append(cards, ci)
			}
		}
		book.Close()
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].Number < cards[j].Number })
	m.AllCardInfos = cards
	return nil
}

func (m *MassProducer) Validate() error {
	if err := m.LoadCards(); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, c := range m.AllCardInfos {
		counts[c.Type]++
	}
	if len(m.AllCardInfos) != 134 || counts["人物"] != 14 || counts["伙伴"] != 48 || counts["技能"] != 28 || counts["道具"] != 44 {
		return fmt.Errorf("unexpected card counts: total=%d types=%v", len(m.AllCardInfos), counts)
	}
	return nil
}

func (m *MassProducer) findArtwork(ci card_maker.CardInfo) string {
	if len(m.artworkFiles) == 0 {
		for _, root := range m.Params.DrawingPaths {
			_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
				if err != nil || entry.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				switch ext {
				case ".png", ".jpg", ".jpeg", ".webp", ".tiff", ".jfif":
					m.artworkFiles[strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))] = path
				}
				return nil
			})
		}
	}
	ids := []string{ci.Number}
	if ci.ArtRef != "" {
		ids = append(ids, ci.ArtRef)
	}
	for _, id := range ids {
		if path := m.artworkFiles[id]; path != "" {
			return path
		}
	}
	return ""
}

func (m *MassProducer) ProduceIDs(ids []string) error {
	previousOverwrite := m.Params.Overwrite
	m.Params.Overwrite = true
	defer func() { m.Params.Overwrite = previousOverwrite }()
	if err := m.LoadCards(); err != nil {
		return err
	}
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	selected := make([]card_maker.CardInfo, 0, len(ids))
	for _, card := range m.AllCardInfos {
		if wanted[card.Number] {
			selected = append(selected, card)
			delete(wanted, card.Number)
		}
	}
	if len(wanted) != 0 {
		missing := make([]string, 0, len(wanted))
		for id := range wanted {
			missing = append(missing, id)
		}
		sort.Strings(missing)
		return fmt.Errorf("unknown card ids: %s", strings.Join(missing, ", "))
	}
	return m.renderCards(selected)
}

func (m *MassProducer) Produce() error {
	if err := m.LoadCards(); err != nil {
		return err
	}
	if err := m.renderCards(m.AllCardInfos); err != nil {
		return err
	}
	return m.writeMetadata()
}

func (m *MassProducer) renderCards(cards []card_maker.CardInfo) error {
	missing := []string{}
	for i := range cards {
		ci := &cards[i]
		art := m.findArtwork(*ci)
		if art == "" {
			missing = append(missing, ci.Number)
			continue
		}
		cfg := m.Config.Copy()
		cfg.StaticDrawing = true
		cfg.StaticDrawingPath = strings.TrimSuffix(art, filepath.Ext(art))
		maker := card_maker.NewCardMaker(*cfg)
		renderCard := *ci
		if category, ok := elementAliases[renderCard.Category]; ok {
			renderCard.Category = category
		}
		img, err := maker.MakeCard(&renderCard)
		if err != nil {
			return fmt.Errorf("%s: %w", ci.Number, err)
		}
		if err = os.MkdirAll(filepath.Dir(filepath.FromSlash(ci.OutputPath)), 0755); err != nil {
			return err
		}
		if _, err = os.Stat(filepath.FromSlash(ci.OutputPath)); err == nil && !m.Params.Overwrite {
			return fmt.Errorf("file %s already exists", ci.OutputPath)
		}
		if err = saveImage(img, filepath.FromSlash(ci.OutputPath), m.Params.IsPrintingVersion); err != nil {
			return err
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing artwork for %d cards: %s", len(missing), strings.Join(missing, ", "))
	}
	return nil
}

func (m *MassProducer) writeMetadata() error {
	if err := os.MkdirAll(m.Params.OutputPath, 0755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(m.AllCardInfos, "", "  ")
	if err := os.WriteFile(filepath.Join(m.Params.OutputPath, "all_card_infos.json"), b, 0644); err != nil {
		return err
	}
	simple := map[string]string{}
	for _, c := range m.AllCardInfos {
		simple[c.Number] = filepath.ToSlash(c.OutputPath)
	}
	b, _ = json.MarshalIndent(simple, "", "  ")
	return os.WriteFile(filepath.Join(m.Params.OutputPath, "simplified_card_infos.json"), b, 0644)
}

func saveImage(img *image.RGBA, path string, printing bool) error {
	if printing {
		path = strings.TrimSuffix(path, filepath.Ext(path)) + ".tiff"
		cmyk := image.NewCMYK(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				c, m, yv, k := color.RGBToCMYK(uint8(r>>8), uint8(g>>8), uint8(b>>8))
				cmyk.Set(x, y, color.CMYK{C: c, M: m, Y: yv, K: k})
			}
		}
		f, e := os.Create(path)
		if e != nil {
			return e
		}
		defer f.Close()
		return tiff.Encode(f, cmyk, nil)
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
}
