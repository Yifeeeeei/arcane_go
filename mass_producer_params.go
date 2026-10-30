package main

type MassProducerParams struct {
	XlsxPaths            []string `json:"xlsx_paths"`
	DrawingPaths         []string `json:"drawing_paths"`
	VersionNames         []string `json:"version_names"`
	ArtworkMigrationPath string   `json:"artwork_migration_path"`
	OutputPath           string   `json:"output_path"`
	SizeRatio            int      `json:"size_ratio"`
	IsPrintingVersion    bool     `json:"is_printing_version"`
	Overwrite            bool     `json:"overwrite"`
	NewCardsOnly         bool     `json:"new_cards_only"`
	GeneralPath          string   `json:"general_path"`
	FontPath             string   `json:"font_path"`
}
