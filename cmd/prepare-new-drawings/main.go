package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type mappingDocument struct {
	Cards []mappingCard `json:"cards"`
}
type mappingCard struct {
	NewID    string  `json:"new_id"`
	LegacyID *string `json:"legacy_id"`
}

func main() {
	b, err := os.ReadFile("new_fork/arcane_artwork_migration.json")
	must(err)
	var document mappingDocument
	must(json.Unmarshal(b, &document))
	assets := map[string]string{}
	must(filepath.WalkDir("drawings", func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp", ".tiff", ".jfif":
			assets[strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))] = path
		}
		return nil
	}))
	must(os.MkdirAll("new_drawings", 0755))
	copied, missing := 0, []string{}
	for _, card := range document.Cards {
		if card.LegacyID == nil {
			missing = append(missing, card.NewID)
			continue
		}
		source := assets[*card.LegacyID]
		if source == "" {
			missing = append(missing, card.NewID)
			continue
		}
		destination := filepath.Join("new_drawings", card.NewID+strings.ToLower(filepath.Ext(source)))
		must(copyFile(source, destination))
		copied++
	}
	fmt.Printf("copied %d drawings to new_drawings; %d unavailable\n", copied, len(missing))
	if len(missing) > 0 {
		fmt.Println("missing:", strings.Join(missing, ", "))
	}
}

func copyFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
