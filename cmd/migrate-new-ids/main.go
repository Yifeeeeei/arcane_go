package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const defaultWorkbookPath = "sheets/重置版/奥术纪元_基础包卡牌设计_七系补全_整理版.xlsx"
const defaultSourceWorkbookPath = "sheets/重置版/奥术纪元_基础包卡牌设计_七系补全.xlsm"
const defaultMigrationPath = "sheets/重置版/arcane_artwork_migration.json"

var elementCodes = map[string]string{"奥术": "A", "无": "A", "火": "F", "火焰": "F", "水": "W", "水纹": "W", "气": "G", "大气": "G", "地": "E", "地脉": "E", "光": "L", "光辉": "L", "暗": "D", "暗影": "D"}
var typeCodes = map[string]string{"人物": "H", "伙伴": "C", "技能": "S", "道具": "I"}
var rarityCodes = map[string]string{"传奇": "L", "普通": "N", "衍生": "D"}
var elementSheets = map[string]bool{"奥术": true, "火": true, "水": true, "气": true, "地": true, "光": true, "暗": true}
var sectionTypes = []string{"伙伴", "技能", "道具"}

type cardUpdate struct {
	NewID   string
	Name    string
	Rarity  string
	Row     int
	Sheet   string
	Subtype string
	Type    string
}

type workbookRow struct {
	Values []string
	Styles []int
	Height float64
	Type   string
}

func main() {
	workbookPath, migrationPath, workbookOnly := defaultWorkbookPath, defaultMigrationPath, false
	for index := 1; index < len(os.Args); index++ {
		switch os.Args[index] {
		case "--workbook":
			index++
			if index >= len(os.Args) {
				panic("--workbook requires a path")
			}
			workbookPath = os.Args[index]
		case "--migration":
			index++
			if index >= len(os.Args) {
				panic("--migration requires a path")
			}
			migrationPath = os.Args[index]
		case "--workbook-only":
			workbookOnly = true
		default:
			panic("unknown argument: " + os.Args[index])
		}
	}
	must(ensureXLSX(workbookPath))
	book, err := excelize.OpenFile(workbookPath)
	must(err)
	defer book.Close()

	must(formatElementSheets(book))
	idMap := map[string]string{}
	updates := map[string]cardUpdate{}
	seenNewIDs := map[string]string{}
	counters := map[string]int{}
	for _, sheet := range book.GetSheetList() {
		rows, err := book.GetRows(sheet)
		must(err)
		headerRow, columns := findHeader(rows)
		if headerRow < 0 {
			continue
		}
		for rowIndex := headerRow + 1; rowIndex < len(rows); rowIndex++ {
			row := rows[rowIndex]
			oldID := cell(row, columns["卡号"])
			if !strings.HasPrefix(oldID, "B") {
				continue
			}
			element := cell(row, columns["属性"])
			elementCode, ok := elementCodes[element]
			if !ok {
				panic(fmt.Sprintf("%s: unknown element %q", oldID, element))
			}
			cardType := "人物"
			if sheet != "人物" {
				cardType = cell(row, columns["类型"])
			}
			typeCode, ok := typeCodes[cardType]
			if !ok {
				panic(fmt.Sprintf("%s: unknown type %q", oldID, cardType))
			}
			rarity := ""
			prefix := "B" + elementCode + typeCode
			if typeCode != "H" {
				rarity = cell(row, columns["稀有度"])
				if sheet == "衍生" {
					rarity = "衍生"
					rarityAxis, _ := excelize.CoordinatesToCellName(columns["稀有度"]+1, rowIndex+1)
					must(book.SetCellValue(sheet, rarityAxis, rarity))
				}
				rarityCode, ok := rarityCodes[rarity]
				if !ok {
					panic(fmt.Sprintf("%s: unknown rarity %q", oldID, rarity))
				}
				prefix += rarityCode
			}
			counters[prefix]++
			newID := fmt.Sprintf("%s%03d", prefix, counters[prefix])
			if previous, duplicate := seenNewIDs[newID]; duplicate {
				panic(fmt.Sprintf("duplicate new card ID %s for %s and %s", newID, previous, oldID))
			}
			seenNewIDs[newID] = oldID
			idMap[oldID] = newID
			updates[oldID] = cardUpdate{
				NewID: newID, Name: cell(row, columns["卡名"]), Rarity: rarity,
				Row: rowIndex + 1, Sheet: sheet, Subtype: cell(row, columns["子类型"]), Type: cardType,
			}
			axis, _ := excelize.CoordinatesToCellName(columns["卡号"]+1, rowIndex+1)
			must(book.SetCellValue(sheet, axis, newID))
		}
	}

	if workbookOnly {
		must(book.Save())
		fmt.Printf("migrated %d IDs and formatted element sections in %s\n", len(idMap), workbookPath)
		return
	}

	b, err := os.ReadFile(migrationPath)
	must(err)
	var document map[string]any
	must(json.Unmarshal(b, &document))
	cards, ok := document["cards"].([]any)
	if !ok {
		panic("migration cards array missing")
	}
	if len(cards) != len(updates) {
		fmt.Printf("workbook has %d cards while migration has %d; existing mappings will be updated and the migration should be rebuilt after generation\n", len(updates), len(cards))
	}
	updatedCards := make([]any, 0, len(cards))
	for _, raw := range cards {
		card := raw.(map[string]any)
		oldID := card["new_id"].(string)
		update, ok := updates[oldID]
		if !ok {
			fmt.Printf("dropped stale artwork mapping %s\n", oldID)
			continue
		}
		card["new_id"] = update.NewID
		card["name"] = update.Name
		card["row"] = update.Row
		card["sheet"] = update.Sheet
		card["subtype"] = update.Subtype
		card["type"] = update.Type
		if update.Rarity == "" {
			card["rarity"] = nil
		} else {
			card["rarity"] = update.Rarity
		}
		updatedCards = append(updatedCards, card)
	}
	document["cards"] = updatedCards

	must(book.Save())
	workbookBytes, err := os.ReadFile(workbookPath)
	must(err)
	sources, ok := document["sources"].(map[string]any)
	if !ok {
		panic("migration sources object missing")
	}
	sources["current_workbook"] = workbookPath
	sources["current_workbook_sha256"] = fmt.Sprintf("%x", sha256.Sum256(workbookBytes))
	b, err = json.MarshalIndent(document, "", "  ")
	must(err)
	b = append(b, '\n')
	tempJSON := migrationPath + ".tmp"
	must(os.WriteFile(tempJSON, b, 0644))
	must(os.Rename(tempJSON, migrationPath))
	fmt.Printf("migrated %d IDs, updated artwork mappings, and formatted element sections\n", len(idMap))
}

func ensureXLSX(workbookPath string) error {
	if strings.ToLower(filepath.Ext(workbookPath)) != ".xlsx" {
		return nil
	}
	if _, err := os.Stat(workbookPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	source := strings.TrimSuffix(workbookPath, filepath.Ext(workbookPath)) + ".xlsm"
	if workbookPath == defaultWorkbookPath {
		source = defaultSourceWorkbookPath
	}
	book, err := excelize.OpenFile(source)
	if err != nil {
		return fmt.Errorf("open source workbook %s: %w", source, err)
	}
	defer book.Close()
	if err = book.SaveAs(workbookPath); err != nil {
		return fmt.Errorf("convert %s to %s: %w", source, workbookPath, err)
	}
	fmt.Printf("converted %s to standard XLSX %s\n", source, workbookPath)
	return nil
}

func formatElementSheets(book *excelize.File) error {
	sectionStyle, err := book.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 13, Color: "FFE6A3"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4A3215"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "top", Color: "C9A85B", Style: 2}, {Type: "bottom", Color: "C9A85B", Style: 2}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	for _, sheet := range book.GetSheetList() {
		if !elementSheets[sheet] {
			continue
		}
		rows, err := book.GetRows(sheet)
		if err != nil {
			return err
		}
		headerRow, columns := findHeader(rows)
		if headerRow < 0 {
			continue
		}
		maxColumn := len(rows[headerRow])
		if maxColumn < 15 {
			maxColumn = 15
		}
		cards := []workbookRow{}
		lastCardRow := 0
		for rowIndex := headerRow + 1; rowIndex < len(rows); rowIndex++ {
			if strings.HasPrefix(cell(rows[rowIndex], columns["卡号"]), "B") {
				lastCardRow = rowIndex + 1
			}
		}
		if lastCardRow == 0 {
			continue
		}
		for rowIndex := headerRow + 1; rowIndex < lastCardRow; rowIndex++ {
			row := rows[rowIndex]
			if !strings.HasPrefix(cell(row, columns["卡号"]), "B") {
				continue
			}
			if cell(row, columns["卡名"]) == "" {
				fmt.Printf("removed empty placeholder row %s!%d (%s)\n", sheet, rowIndex+1, cell(row, columns["卡号"]))
				continue
			}
			cardType := cell(row, columns["类型"])
			if cardType == "" {
				cardType = inferTypeFromID(cell(row, columns["卡号"]))
			}
			snapshot := workbookRow{Values: make([]string, maxColumn), Styles: make([]int, maxColumn), Type: cardType}
			for column := 1; column <= maxColumn; column++ {
				axis, _ := excelize.CoordinatesToCellName(column, rowIndex+1)
				snapshot.Values[column-1], err = book.GetCellValue(sheet, axis)
				if err != nil {
					return err
				}
				snapshot.Styles[column-1], err = book.GetCellStyle(sheet, axis)
				if err != nil {
					return err
				}
			}
			if snapshot.Values[columns["类型"]] == "" {
				snapshot.Values[columns["类型"]] = cardType
			}
			if snapshot.Values[columns["属性"]] == "" {
				snapshot.Values[columns["属性"]] = inferElementFromID(cell(row, columns["卡号"]))
			}
			if snapshot.Values[columns["稀有度"]] == "" {
				snapshot.Values[columns["稀有度"]] = inferRarityFromID(cell(row, columns["卡号"]))
			}
			snapshot.Height, err = book.GetRowHeight(sheet, rowIndex+1)
			if err != nil {
				return err
			}
			cards = append(cards, snapshot)
		}
		startRow := headerRow + 2
		finalRow := startRow + len(cards) + len(sectionTypes) - 1
		clearThrough := finalRow
		if lastCardRow > clearThrough {
			clearThrough = lastCardRow
		}
		for row := startRow; row <= clearThrough; row++ {
			_ = book.UnmergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("O%d", row))
			for column := 1; column <= maxColumn; column++ {
				axis, _ := excelize.CoordinatesToCellName(column, row)
				if err = book.SetCellValue(sheet, axis, ""); err != nil {
					return err
				}
			}
		}
		byType := map[string][]workbookRow{}
		for _, card := range cards {
			byType[card.Type] = append(byType[card.Type], card)
		}
		if len(byType["伙伴"])+len(byType["技能"])+len(byType["道具"]) != len(cards) {
			return fmt.Errorf("%s contains an unsupported card type", sheet)
		}
		writeRow := startRow
		for _, cardType := range sectionTypes {
			lastAxis, _ := excelize.CoordinatesToCellName(maxColumn, writeRow)
			if err = book.SetCellStyle(sheet, fmt.Sprintf("A%d", writeRow), lastAxis, sectionStyle); err != nil {
				return err
			}
			if err = book.SetCellValue(sheet, fmt.Sprintf("A%d", writeRow), cardType+"牌"); err != nil {
				return err
			}
			if err = book.MergeCell(sheet, fmt.Sprintf("A%d", writeRow), lastAxis); err != nil {
				return err
			}
			if err = book.SetRowHeight(sheet, writeRow, 24); err != nil {
				return err
			}
			writeRow++
			for _, snapshot := range byType[cardType] {
				for column := 1; column <= maxColumn; column++ {
					axis, _ := excelize.CoordinatesToCellName(column, writeRow)
					if err = book.SetCellValue(sheet, axis, snapshot.Values[column-1]); err != nil {
						return err
					}
					if err = book.SetCellStyle(sheet, axis, axis, snapshot.Styles[column-1]); err != nil {
						return err
					}
				}
				if err = book.SetRowHeight(sheet, writeRow, snapshot.Height); err != nil {
					return err
				}
				writeRow++
			}
		}
	}
	return nil
}

func inferTypeFromID(id string) string {
	if len(id) < 3 {
		return ""
	}
	switch id[2] {
	case 'C':
		return "伙伴"
	case 'S':
		return "技能"
	case 'I':
		return "道具"
	default:
		return ""
	}
}

func inferElementFromID(id string) string {
	if len(id) < 2 {
		return ""
	}
	switch id[1] {
	case 'A':
		return "无"
	case 'F':
		return "火"
	case 'W':
		return "水"
	case 'G':
		return "气"
	case 'E':
		return "地"
	case 'L':
		return "光"
	case 'D':
		return "暗"
	default:
		return ""
	}
}

func inferRarityFromID(id string) string {
	if len(id) < 4 {
		return ""
	}
	switch id[3] {
	case 'L':
		return "传奇"
	case 'N':
		return "普通"
	case 'D':
		return "衍生"
	default:
		return ""
	}
}

func findHeader(rows [][]string) (int, map[string]int) {
	for rowIndex, row := range rows {
		columns := map[string]int{}
		for column, value := range row {
			columns[strings.TrimSpace(value)] = column
		}
		if _, ok := columns["卡号"]; ok {
			return rowIndex, columns
		}
	}
	return -1, nil
}

func cell(row []string, column int) string {
	if column < 0 || column >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[column])
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
