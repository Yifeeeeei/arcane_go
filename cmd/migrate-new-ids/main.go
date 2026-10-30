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

const workbookPath = "new_fork/奥术纪元_基础包卡牌设计_v0.2.xlsm"
const migrationPath = "new_fork/arcane_artwork_migration.json"

var elementCodes = map[string]string{"奥术": "A", "无": "A", "火": "F", "火焰": "F", "水": "W", "水纹": "W", "地": "E", "地脉": "E", "暗": "D", "暗影": "D", "光": "L", "光辉": "L", "气": "R", "大气": "R"}
var typeCodes = map[string]string{"人物": "H", "伙伴": "C", "技能": "S", "道具": "I"}

func main() {
	book, err := excelize.OpenFile(workbookPath)
	must(err)
	defer book.Close()
	idMap := map[string]string{}
	counters := map[string]int{}
	for _, sheet := range book.GetSheetList() {
		rows, err := book.GetRows(sheet)
		must(err)
		headerRow, idColumn, elementColumn, typeColumn := -1, -1, -1, -1
		for rowIndex, row := range rows {
			for column, value := range row {
				switch strings.TrimSpace(value) {
				case "卡号":
					idColumn = column
					headerRow = rowIndex
				case "属性":
					elementColumn = column
				case "类型":
					typeColumn = column
				}
			}
			if headerRow == rowIndex && idColumn >= 0 && elementColumn >= 0 {
				break
			}
		}
		if headerRow < 0 {
			continue
		}
		for rowIndex := headerRow + 1; rowIndex < len(rows); rowIndex++ {
			row := rows[rowIndex]
			if idColumn >= len(row) || strings.TrimSpace(row[idColumn]) == "" {
				continue
			}
			oldID := strings.TrimSpace(row[idColumn])
			element := cell(row, elementColumn)
			cardType := "人物"
			if sheet != "人物" {
				cardType = cell(row, typeColumn)
			}
			elementCode, ok := elementCodes[element]
			if !ok {
				panic(fmt.Sprintf("%s: unknown element %q", oldID, element))
			}
			typeCode, ok := typeCodes[cardType]
			if !ok {
				panic(fmt.Sprintf("%s: unknown type %q", oldID, cardType))
			}
			key := elementCode + typeCode
			counters[key]++
			newID := fmt.Sprintf("B%s%s%03d", elementCode, typeCode, counters[key])
			idMap[oldID] = newID
			axis, _ := excelize.CoordinatesToCellName(idColumn+1, rowIndex+1)
			must(book.SetCellValue(sheet, axis, newID))
		}
	}
	tempWorkbook := workbookPath + ".tmp.xlsm"
	must(book.SaveAs(tempWorkbook))
	must(os.Rename(tempWorkbook, workbookPath))
	b, err := os.ReadFile(migrationPath)
	must(err)
	var document map[string]any
	must(json.Unmarshal(b, &document))
	cards, ok := document["cards"].([]any)
	if !ok {
		panic("migration cards array missing")
	}
	for _, raw := range cards {
		card := raw.(map[string]any)
		oldID := card["new_id"].(string)
		newID, ok := idMap[oldID]
		if !ok {
			panic("mapping missing for " + oldID)
		}
		card["new_id"] = newID
	}
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
	fmt.Printf("migrated %d IDs in %s and %s\n", len(idMap), filepath.Base(workbookPath), filepath.Base(migrationPath))
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
