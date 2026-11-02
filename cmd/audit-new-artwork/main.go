package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const workbookPath = "sheets/重置版/奥术纪元_基础包卡牌设计_v0.2.xlsm"
const migrationPath = "sheets/重置版/arcane_artwork_migration.json"

type migrationDocument struct {
	Cards []migrationCard `json:"cards"`
}

type migrationCard struct {
	LegacyID           *string `json:"legacy_id"`
	LegacyRow          *int    `json:"legacy_row"`
	LegacySheet        *string `json:"legacy_sheet"`
	LegacyWorkbook     *string `json:"legacy_workbook"`
	Name               string  `json:"name"`
	NewID              string  `json:"new_id"`
	Status             string  `json:"status"`
	NameMatchConfirmed bool    `json:"name_match_confirmed"`
}

func main() {
	current, err := loadCurrentCards()
	must(err)
	b, err := os.ReadFile(migrationPath)
	must(err)
	var migration migrationDocument
	must(json.Unmarshal(b, &migration))
	legacyImages, err := indexImages("drawings")
	must(err)
	newImages, err := indexImages("new_drawings")
	must(err)
	workbooks := map[string]*excelize.File{}
	defer func() {
		for _, book := range workbooks {
			_ = book.Close()
		}
	}()

	errors := []string{}
	missingArtwork := []string{}
	nameReviews := []string{}
	exactNames, normalizedNames, matchedImages := 0, 0, 0
	seen := map[string]bool{}
	for _, card := range migration.Cards {
		seen[card.NewID] = true
		currentName, ok := current[card.NewID]
		if !ok {
			errors = append(errors, fmt.Sprintf("%s is in migration but not workbook", card.NewID))
			continue
		}
		if currentName != card.Name {
			errors = append(errors, fmt.Sprintf("%s name mismatch: workbook=%q migration=%q", card.NewID, currentName, card.Name))
		}
		newImage := newImages[card.NewID]
		if card.LegacyID == nil {
			if newImage == "" {
				missingArtwork = append(missingArtwork, card.NewID+" "+currentName)
			}
			continue
		}
		if card.LegacyWorkbook == nil || card.LegacySheet == nil || card.LegacyRow == nil {
			errors = append(errors, card.NewID+" has incomplete legacy source metadata")
			continue
		}
		legacyName, err := readLegacyName(workbooks, *card.LegacyWorkbook, *card.LegacySheet, *card.LegacyRow)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s legacy name: %v", card.NewID, err))
			continue
		}
		if legacyName == currentName {
			exactNames++
		} else if normalizeName(legacyName) == normalizeName(currentName) {
			normalizedNames++
		} else if card.NameMatchConfirmed {
			normalizedNames++
		} else {
			nameReviews = append(nameReviews, fmt.Sprintf("%s current=%q legacy=%q", card.NewID, currentName, legacyName))
		}
		legacyImage := legacyImages[*card.LegacyID]
		if legacyImage == "" {
			errors = append(errors, fmt.Sprintf("%s legacy image %s is missing", card.NewID, *card.LegacyID))
			continue
		}
		if newImage == "" {
			errors = append(errors, fmt.Sprintf("%s new drawing is missing", card.NewID))
			continue
		}
		legacyHash, err := fileHash(legacyImage)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		newHash, err := fileHash(newImage)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		if legacyHash != newHash {
			errors = append(errors, fmt.Sprintf("%s image hash differs from legacy %s", card.NewID, *card.LegacyID))
			continue
		}
		matchedImages++
	}
	for id, name := range current {
		if !seen[id] {
			errors = append(errors, fmt.Sprintf("%s %q is missing from migration", id, name))
		}
	}

	fmt.Printf("verified %d drawings are byte-identical to their selected legacy images\n", matchedImages)
	fmt.Printf("names: %d exact, %d equivalent after punctuation/scroll-suffix normalization, %d need review\n", exactNames, normalizedNames, len(nameReviews))
	for _, item := range nameReviews {
		fmt.Println("  NAME REVIEW:", item)
	}
	fmt.Printf("%d cards intentionally have no legacy drawing\n", len(missingArtwork))
	for _, item := range missingArtwork {
		fmt.Println("  missing:", item)
	}
	if len(errors) > 0 {
		for _, item := range errors {
			fmt.Println("ERROR:", item)
		}
		panic(fmt.Sprintf("artwork audit failed with %d errors", len(errors)))
	}
}

func normalizeName(name string) string {
	name = strings.NewReplacer(" ", "", "·", "", "“", "", "”", "", "\"", "").Replace(name)
	return strings.TrimSuffix(name, "卷轴")
}

func loadCurrentCards() (map[string]string, error) {
	book, err := excelize.OpenFile(workbookPath)
	if err != nil {
		return nil, err
	}
	defer book.Close()
	cards := map[string]string{}
	for _, sheet := range book.GetSheetList() {
		rows, err := book.GetRows(sheet)
		if err != nil {
			return nil, err
		}
		header, columns := findHeader(rows)
		if header < 0 {
			continue
		}
		for _, row := range rows[header+1:] {
			id := cell(row, columns["卡号"])
			if id != "" {
				cards[id] = cell(row, columns["卡名"])
			}
		}
	}
	return cards, nil
}

func readLegacyName(cache map[string]*excelize.File, relativePath, sheet string, rowNumber int) (string, error) {
	path := filepath.Join("sheets", filepath.FromSlash(relativePath))
	book := cache[path]
	if book == nil {
		var err error
		book, err = excelize.OpenFile(path)
		if err != nil {
			return "", err
		}
		cache[path] = book
	}
	rows, err := book.GetRows(sheet)
	if err != nil {
		return "", err
	}
	if rowNumber < 1 || rowNumber > len(rows) {
		return "", fmt.Errorf("%s/%s row %d is out of range", relativePath, sheet, rowNumber)
	}
	_, columns := findHeader(rows[:rowNumber])
	nameColumn, ok := columns["名称"]
	if !ok {
		nameColumn, ok = columns["卡名"]
	}
	if !ok {
		return "", fmt.Errorf("%s/%s has no name column", relativePath, sheet)
	}
	return cell(rows[rowNumber-1], nameColumn), nil
}

func findHeader(rows [][]string) (int, map[string]int) {
	for rowIndex, row := range rows {
		columns := map[string]int{}
		for column, value := range row {
			columns[strings.TrimSpace(value)] = column
		}
		if _, hasID := columns["卡号"]; hasID {
			return rowIndex, columns
		}
		if _, hasID := columns["编号"]; hasID {
			return rowIndex, columns
		}
	}
	return -1, map[string]int{}
}

func indexImages(root string) (map[string]string, error) {
	images := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".png", ".jpg", ".jpeg", ".webp", ".tiff", ".jfif":
			images[strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))] = path
		}
		return nil
	})
	return images, err
}

func fileHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
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
