package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

const defaultWorkbookPath = "sheets/重置版/奥术纪元_基础包卡牌设计_v0.2.xlsm"
const defaultMigrationPath = "sheets/重置版/arcane_artwork_migration.json"

var elementCodes = map[string]string{"奥术": "A", "无": "A", "火": "F", "火焰": "F", "水": "W", "水纹": "W", "气": "G", "大气": "G", "地": "E", "地脉": "E", "光": "L", "光辉": "L", "暗": "D", "暗影": "D"}
var typeCodes = map[string]string{"人物": "H", "伙伴": "C", "技能": "S", "道具": "I"}
var rarityCodes = map[string]string{"传奇": "L", "普通": "N", "衍生": "D"}
var sequencePattern = regexp.MustCompile(`[0-9]{3}$`)

type cardUpdate struct {
	NewID   string
	Name    string
	Rarity  string
	Row     int
	Sheet   string
	Subtype string
	Type    string
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
	book, err := excelize.OpenFile(workbookPath)
	must(err)
	defer book.Close()

	idMap := map[string]string{}
	updates := map[string]cardUpdate{}
	for _, sheet := range book.GetSheetList() {
		rows, err := book.GetRows(sheet)
		must(err)
		headerRow := -1
		columns := map[string]int{}
		for rowIndex, row := range rows {
			for column, value := range row {
				columns[strings.TrimSpace(value)] = column
			}
			if _, ok := columns["卡号"]; ok {
				headerRow = rowIndex
				break
			}
			columns = map[string]int{}
		}
		if headerRow < 0 {
			continue
		}
		for rowIndex := headerRow + 1; rowIndex < len(rows); rowIndex++ {
			row := rows[rowIndex]
			oldID := cell(row, columns["卡号"])
			if oldID == "" {
				continue
			}
			sequence := sequencePattern.FindString(oldID)
			if sequence == "" {
				panic(fmt.Sprintf("%s: card ID must end in a three-digit sequence", oldID))
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
			newID := "B" + elementCode + typeCode
			if typeCode != "H" {
				rarity = cell(row, columns["稀有度"])
				rarityCode, ok := rarityCodes[rarity]
				if !ok {
					panic(fmt.Sprintf("%s: unknown rarity %q", oldID, rarity))
				}
				newID += rarityCode
			}
			newID += sequence
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
		fmt.Printf("migrated %d IDs in %s\n", len(idMap), workbookPath)
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
		panic(fmt.Sprintf("workbook has %d cards but migration has %d", len(updates), len(cards)))
	}
	for _, raw := range cards {
		card := raw.(map[string]any)
		oldID := card["new_id"].(string)
		update, ok := updates[oldID]
		if !ok {
			panic("mapping missing for " + oldID)
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
	}

	must(book.Save())
	must(renameDrawings(idMap))
	workbookBytes, err := os.ReadFile(workbookPath)
	must(err)
	sources, ok := document["sources"].(map[string]any)
	if !ok {
		panic("migration sources object missing")
	}
	sources["current_workbook_sha256"] = fmt.Sprintf("%x", sha256.Sum256(workbookBytes))
	b, err = json.MarshalIndent(document, "", "  ")
	must(err)
	b = append(b, '\n')
	tempJSON := migrationPath + ".tmp"
	must(os.WriteFile(tempJSON, b, 0644))
	must(os.Rename(tempJSON, migrationPath))
	fmt.Printf("migrated %d IDs and renamed matching new_drawings files\n", len(idMap))
}

func renameDrawings(idMap map[string]string) error {
	entries, err := os.ReadDir("new_drawings")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		extension := filepath.Ext(entry.Name())
		oldID := strings.TrimSuffix(entry.Name(), extension)
		newID, ok := idMap[oldID]
		if !ok || newID == oldID {
			continue
		}
		source := filepath.Join("new_drawings", entry.Name())
		destination := filepath.Join("new_drawings", newID+extension)
		if _, err = os.Stat(destination); err == nil {
			return fmt.Errorf("cannot rename %s: destination %s already exists", source, destination)
		} else if !os.IsNotExist(err) {
			return err
		}
		if err = os.Rename(source, destination); err != nil {
			return err
		}
	}
	return nil
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
