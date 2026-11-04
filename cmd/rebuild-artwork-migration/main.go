package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const migrationPath = "sheets/重置版/arcane_artwork_migration.json"

type cardInfo struct {
	Number      string `json:"number"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Subtype     string `json:"subtype"`
	Rarity      string `json:"rarity"`
	Source      string `json:"source"`
	VersionName string `json:"version_name"`
}

type params struct {
	XlsxPaths []string `json:"xlsx_paths"`
}

func main() {
	current := readCards("new_output/all_card_infos.json")
	legacy := readCards("output/all_card_infos.json")
	legacyImages := indexImages("drawings")
	legacyByName := map[string][]cardInfo{}
	legacyByNormalizedName := map[string][]cardInfo{}
	legacyByID := map[string]cardInfo{}
	for _, card := range legacy {
		if legacyImages[card.Number] != "" {
			legacyByName[card.Name] = append(legacyByName[card.Name], card)
			legacyByNormalizedName[normalizeName(card.Name)] = append(legacyByNormalizedName[normalizeName(card.Name)], card)
			legacyByID[card.Number] = card
		}
	}

	b, err := os.ReadFile(migrationPath)
	must(err)
	var document map[string]any
	must(json.Unmarshal(b, &document))
	existingCards, ok := document["cards"].([]any)
	if !ok {
		panic("migration cards array missing")
	}
	existingByID := map[string]map[string]any{}
	existingByName := map[string]map[string]any{}
	for _, raw := range existingCards {
		card := raw.(map[string]any)
		existingByID[card["new_id"].(string)] = card
		if name, ok := card["name"].(string); ok && name != "" {
			existingByName[name] = card
		}
	}

	sort.Slice(current, func(i, j int) bool { return current[i].Number < current[j].Number })
	rebuilt := make([]any, 0, len(current))
	matched, missing := 0, 0
	for _, card := range current {
		record := existingByID[card.Number]
		if record == nil || record["name"] != card.Name {
			record = existingByName[card.Name]
		}
		if record == nil {
			record = map[string]any{}
		}
		legacyID, hasLegacy := record["legacy_id"].(string)
		nameMatchConfirmed, _ := record["name_match_confirmed"].(bool)
		legacyCard, legacyExists := legacyByID[legacyID]
		preserveLegacy := hasLegacy && legacyID != "" && legacyExists && (legacyCard.Name == card.Name || nameMatchConfirmed)
		if !preserveLegacy {
			candidates := legacyByName[card.Name]
			status := "legacy_name_match"
			if len(candidates) == 0 {
				candidates = legacyByNormalizedName[normalizeName(card.Name)]
				status = "legacy_normalized_name_match"
			}
			switch len(candidates) {
			case 0:
				record["legacy_id"] = nil
				record["legacy_row"] = nil
				record["legacy_set"] = nil
				record["legacy_sheet"] = nil
				record["legacy_workbook"] = nil
				record["status"] = "new_design_no_legacy_art"
				missing++
			case 1:
				candidate := candidates[0]
				record["legacy_id"] = candidate.Number
				record["legacy_row"] = nil
				record["legacy_set"] = candidate.VersionName
				record["legacy_sheet"] = nil
				record["legacy_workbook"] = nil
				record["status"] = status
				matched++
			default:
				ids := make([]string, 0, len(candidates))
				for _, candidate := range candidates {
					ids = append(ids, candidate.Number)
				}
				panic(fmt.Sprintf("%s %q has ambiguous legacy matches: %s", card.Number, card.Name, strings.Join(ids, ", ")))
			}
		} else {
			matched++
		}
		record["name"] = card.Name
		record["new_id"] = card.Number
		record["rarity"] = nullable(card.Rarity)
		record["sheet"] = card.Source
		record["subtype"] = card.Subtype
		record["type"] = card.Type
		rebuilt = append(rebuilt, record)
	}
	document["cards"] = rebuilt
	updateSources(document)
	b, err = json.MarshalIndent(document, "", "  ")
	must(err)
	b = append(b, '\n')
	temporary := migrationPath + ".tmp"
	must(os.WriteFile(temporary, b, 0644))
	must(os.Rename(temporary, migrationPath))
	fmt.Printf("rebuilt %d mappings: %d legacy drawings, %d without legacy art\n", len(rebuilt), matched, missing)
}

func normalizeName(name string) string {
	name = strings.NewReplacer(" ", "", "·", "", "“", "", "”", "", string(rune(34)), "").Replace(name)
	return strings.TrimSuffix(name, "卷轴")
}

func updateSources(document map[string]any) {
	b, err := os.ReadFile("new_mass_producer_params.json")
	must(err)
	var p params
	must(json.Unmarshal(b, &p))
	if len(p.XlsxPaths) == 0 {
		panic("xlsx_paths is empty")
	}
	workbook, err := os.ReadFile(p.XlsxPaths[0])
	must(err)
	sources, ok := document["sources"].(map[string]any)
	if !ok {
		sources = map[string]any{}
		document["sources"] = sources
	}
	sources["current_workbook"] = filepath.ToSlash(p.XlsxPaths[0])
	sources["current_workbook_sha256"] = fmt.Sprintf("%x", sha256.Sum256(workbook))
}

func readCards(path string) []cardInfo {
	b, err := os.ReadFile(path)
	must(err)
	var cards []cardInfo
	must(json.Unmarshal(b, &cards))
	return cards
}

func indexImages(root string) map[string]string {
	images := map[string]string{}
	must(filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".png", ".jpg", ".jpeg", ".webp", ".tiff", ".jfif":
			images[strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))] = path
		}
		return nil
	}))
	return images
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
