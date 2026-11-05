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

// Heroes: B + A/F/W/... + H + 001.
// Main cards: B + A/F/W/... + C/S/I + L/N/D + 001, where the fourth
// character identifies legendary, normal, or derived cards.
var newCardID = regexp.MustCompile(`^B[AFWGEDL](?:H[0-9]{3}|[CSI][LND][0-9]{3})$`)
var legacyCardID = regexp.MustCompile(`^[0-9]{7}$`)
var elementIDCodes = map[string]byte{"奥术": 'A', "火": 'F', "水": 'W', "气": 'G', "地": 'E', "光": 'L', "暗": 'D'}
var typeIDCodes = map[string]byte{card_maker.TYPE_HERO_ZH: 'H', card_maker.TYPE_UNIT_ZH: 'C', card_maker.TYPE_ABILITY_ZH: 'S', card_maker.TYPE_ITEM_ZH: 'I'}
var rarityIDCodes = map[string]byte{"传奇": 'L', "普通": 'N', "衍生": 'D'}

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
var elementOutputDirectories = map[string]string{"火": "火", "火焰": "火", "水": "水", "水纹": "水", "气": "气", "大气": "气", "地": "地", "地脉": "地", "光": "光", "光辉": "光", "暗": "暗", "暗影": "暗", "奥术": "无", "无": "无"}

func parseElements(raw string) (card_maker.Elements, error) {
	result := card_maker.Elements{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result, nil
	}
	normalized := strings.NewReplacer("；", " ", ";", " ", "，", " ", ",", " ", "/", " ", "\\", " ", "+", " ").Replace(raw)
	re := regexp.MustCompile(`(?:([\p{Han}]+)\s*([0-9]+(?:\.0+)?)|([0-9]+(?:\.0+)?)\s*([\p{Han}]+))`)
	matches := re.FindAllStringSubmatch(normalized, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid element vector %q", raw)
	}
	consumed := re.ReplaceAllString(normalized, "")
	if strings.TrimSpace(consumed) != "" {
		return nil, fmt.Errorf("invalid element vector %q", raw)
	}
	for _, match := range matches {
		elementText, valueText := match[1], match[2]
		if elementText == "" {
			elementText, valueText = match[4], match[3]
		}
		name, ok := elementAliases[elementText]
		if !ok {
			return nil, fmt.Errorf("unknown element %q", elementText)
		}
		v, err := parseOptionalInt(valueText, "element value")
		if err != nil {
			return nil, err
		}
		result[name] += *v
	}
	return result, nil
}

func parseLegacyElements(raw string) card_maker.Elements {
	result := card_maker.Elements{}
	re := regexp.MustCompile(`([0-9]+)\s*([\p{Han}]+)`)
	for _, match := range re.FindAllStringSubmatch(strings.TrimSpace(raw), -1) {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		element, ok := elementAliases[match[2]]
		if !ok {
			continue
		}
		result[element] += value
	}
	return result
}

func parseSheet(rows [][]string) ([]Row, error) {
	headerAt := -1
	for i, row := range rows {
		var hasID, hasName bool
		for _, c := range row {
			c = strings.TrimSpace(c)
			hasID = hasID || c == "卡号" || c == "编号"
			hasName = hasName || c == "卡名" || c == "名称"
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
	if err = validateNewCardIDSemantics(ci); err != nil {
		return ci, err
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
	outputElement := elementOutputDirectories[ci.Category]
	if outputElement == "" {
		outputElement = ci.Category
	}
	ci.OutputPath = filepath.ToSlash(filepath.Join(m.Params.OutputPath, ci.SetName, ci.Type, outputElement, ci.Number+".jpg"))
	if a, ok := m.artwork[id]; ok && a.LegacyID != nil {
		ci.ArtRef = *a.LegacyID
	}
	return ci, nil
}

func validateNewCardIDSemantics(card card_maker.CardInfo) error {
	element := elementAliases[card.Category]
	expectedElement, ok := elementIDCodes[element]
	if !ok || len(card.Number) < 3 || card.Number[1] != expectedElement {
		return fmt.Errorf("%s: ID element does not match 属性 %q", card.Number, card.Category)
	}
	expectedType, ok := typeIDCodes[card.Type]
	if !ok || card.Number[2] != expectedType {
		return fmt.Errorf("%s: ID type does not match 类型 %q", card.Number, card.Type)
	}
	if card.Type == card_maker.TYPE_HERO_ZH {
		if len(card.Number) != 6 {
			return fmt.Errorf("%s: hero ID must not contain a rarity code", card.Number)
		}
		return nil
	}
	expectedRarity, ok := rarityIDCodes[card.Rarity]
	if !ok || len(card.Number) != 7 || card.Number[3] != expectedRarity {
		return fmt.Errorf("%s: ID rarity does not match 稀有度 %q", card.Number, card.Rarity)
	}
	return nil
}

func legacyType(id string) (string, error) {
	if len(id) != 7 {
		return "", fmt.Errorf("legacy card id must have 7 digits, got %q", id)
	}
	switch id[0] {
	case '1':
		return card_maker.TYPE_UNIT_ZH, nil
	case '2':
		return card_maker.TYPE_ITEM_ZH, nil
	case '3':
		return card_maker.TYPE_ABILITY_ZH, nil
	case '4':
		return card_maker.TYPE_HERO_ZH, nil
	default:
		return "", fmt.Errorf("unknown legacy card type in %q", id)
	}
}

func legacyInt(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return -1
	}
	return value
}

func (m *MassProducer) legacyCardFromRow(row Row, setName string) (card_maker.CardInfo, error) {
	id := strings.TrimSpace(row["编号"])
	cardType, err := legacyType(id)
	if err != nil {
		return card_maker.CardInfo{}, err
	}
	ci := card_maker.CardInfo{Number: id, Type: cardType, Name: row["名称"], Category: row["属性"], Tag: row["种类"], Description: row["效果"], Quote: row["引言"], ElementsCost: parseLegacyElements(row["条件"]), ElementsGain: parseLegacyElements(row["负载"]), ElementsExpense: parseLegacyElements(row["代价"]), Attack: legacyInt(row["攻击"]), Life: legacyInt(row["生命"]), Power: legacyInt(row["威力"]), Duration: legacyInt(row["时间"]), VersionName: setName, Legacy: true}
	if row["标签"] != "" {
		ci.Tag = row["标签"]
	}
	ci.OutputPath = filepath.ToSlash(filepath.Join(m.Params.OutputPath, setName, ci.Type, ci.Category, ci.Number+".jpg"))
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
				newID := strings.TrimSpace(row["卡号"])
				legacyID := strings.TrimSpace(row["编号"])
				var ci card_maker.CardInfo
				if newCardID.MatchString(newID) {
					ci, err = m.cardFromRow(row, sheet, setName)
				} else if legacyCardID.MatchString(legacyID) {
					ci, err = m.legacyCardFromRow(row, setName)
				} else {
					continue
				}
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
	if len(m.AllCardInfos) > 0 && m.AllCardInfos[0].Legacy {
		return nil
	}
	counts := map[string]int{}
	for _, c := range m.AllCardInfos {
		counts[c.Type]++
	}
	threeElementBase := len(m.AllCardInfos) == 134 && counts["人物"] == 14 && counts["伙伴"] == 48 && counts["技能"] == 28 && counts["道具"] == 44
	sevenElementBaseV1 := len(m.AllCardInfos) == 343 && counts["人物"] == 34 && counts["伙伴"] == 129 && counts["技能"] == 72 && counts["道具"] == 108
	sevenElementBaseLatest := len(m.AllCardInfos) == 338 && counts["人物"] == 28 && counts["伙伴"] == 130 && counts["技能"] == 72 && counts["道具"] == 108
	if !threeElementBase && !sevenElementBaseV1 && !sevenElementBaseLatest {
		byName := map[string][]string{}
		for _, card := range m.AllCardInfos {
			byName[card.Name] = append(byName[card.Name], card.Number)
		}
		duplicates := []string{}
		for name, ids := range byName {
			if name != "" && len(ids) > 1 {
				sort.Strings(ids)
				duplicates = append(duplicates, fmt.Sprintf("%s=%s", name, strings.Join(ids, ",")))
			}
		}
		sort.Strings(duplicates)
		return fmt.Errorf("unexpected card counts: total=%d types=%v duplicate_names=%v", len(m.AllCardInfos), counts, duplicates)
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
	if err := m.pruneStaleOutputImages(); err != nil {
		return err
	}
	if err := m.renderCards(m.AllCardInfos); err != nil {
		return err
	}
	return m.writeMetadata()
}

func (m *MassProducer) pruneStaleOutputImages() error {
	root := filepath.Clean(m.Params.OutputPath)
	if root == "." || root == string(filepath.Separator) || filepath.VolumeName(root)+string(filepath.Separator) == root {
		return fmt.Errorf("refusing to clean unsafe output path %q", m.Params.OutputPath)
	}
	desired := map[string]bool{}
	for _, card := range m.AllCardInfos {
		desired[filepath.Clean(filepath.FromSlash(card.OutputPath))] = true
	}
	removed := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if entry.IsDir() || desired[filepath.Clean(path)] {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".png", ".jpg", ".jpeg", ".webp", ".tiff", ".jfif":
			if err = os.Remove(path); err != nil {
				return err
			}
			removed++
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if removed > 0 {
		fmt.Printf("removed %d stale output images\n", removed)
	}
	return nil
}

func (m *MassProducer) renderCards(cards []card_maker.CardInfo) error {
	missing := []string{}
	substituted := []string{}
	for i := range cards {
		ci := &cards[i]
		art := m.findArtwork(*ci)
		if art == "" {
			if !m.Params.UseBlankArtworkForMissing {
				missing = append(missing, ci.Number)
				continue
			}
			substituted = append(substituted, ci.Number)
		}
		cfg := m.Config.Copy()
		if art == "" {
			cfg.UseBlankDrawing = true
		} else {
			cfg.StaticDrawing = true
			cfg.StaticDrawingPath = strings.TrimSuffix(art, filepath.Ext(art))
		}
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
	if len(substituted) > 0 {
		fmt.Printf("used blank artwork for %d cards: %s\n", len(substituted), strings.Join(substituted, ", "))
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
