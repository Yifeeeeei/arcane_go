package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Yifeeeeei/arcane_go/card_maker"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"github.com/xuri/excelize/v2"
	"golang.org/x/image/tiff"
)

const (
	OUTPUT_PATH = "output"
)

type MassProducer struct {
	Params            *MassProducerParams
	OldAllCardsInfos  map[string]card_maker.CardInfo
	AllCardInfos      []card_maker.CardInfo
	Config            *card_maker.Config
	Mu                *sync.Mutex
	Log               []string
	LogMu             *sync.Mutex
	SanityCheckPassed bool
	SanityMu          *sync.Mutex
}

func NewMassProducer(paramPath string) (*MassProducer, error) {
	jsonBytes, err := os.ReadFile(paramPath)
	if err != nil {
		return nil, err
	}
	params := &MassProducerParams{}
	err = json.Unmarshal(jsonBytes, params)
	if err != nil {
		return nil, err
	}

	massProducer := &MassProducer{Params: params}
	cfg := card_maker.NewDefaultConfig(params.SizeRatio, params.GeneralPath, "dummydrawingpath", params.FontPath)
	massProducer.Config = &cfg
	massProducer.Config.GeneralPath = params.GeneralPath
	massProducer.Config.FontPath = params.FontPath

	massProducer.OldAllCardsInfos = nil

	listOldCardInfos := []card_maker.CardInfo{}
	_, err = os.Stat(filepath.Join(OUTPUT_PATH, "all_card_infos.json"))
	if err == nil {
		jsonBytes, err = os.ReadFile(filepath.Join(OUTPUT_PATH, "all_card_infos.json"))
		if err == nil {
			err = json.Unmarshal(jsonBytes, &listOldCardInfos)
			if err == nil {
				massProducer.OldAllCardsInfos = map[string]card_maker.CardInfo{}
				for _, cardInfo := range listOldCardInfos {
					massProducer.OldAllCardsInfos[cardInfo.Number] = cardInfo
				}
			} else {
				log.Printf("unmarshal old card infos failed %v", err)
			}

		}
	}

	if (len(massProducer.Params.XlsxPaths) != len(massProducer.Params.DrawingPaths)) || (len(massProducer.Params.XlsxPaths) != len(massProducer.Params.VersionNames)) || (len(massProducer.Params.XlsxPaths) != len(massProducer.Params.DrawingPaths)) {
		return nil, fmt.Errorf("length of xlsx_paths, drawing_paths, version_names should be the same")
	}

	massProducer.AllCardInfos = []card_maker.CardInfo{}
	massProducer.Mu = &sync.Mutex{}
	massProducer.LogMu = &sync.Mutex{}

	massProducer.SanityCheckPassed = true
	massProducer.SanityMu = &sync.Mutex{}

	return massProducer, nil
}

func (m *MassProducer) makeDir(path string) error {
	// if exist, do nothing, otherwise create it
	_, err := os.Stat(path)
	if err != nil {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		return nil
	}
	if !m.Params.Overwrite {
		return fmt.Errorf("path %s already exists", path)
	}
	return nil
}

func (m *MassProducer) cleanString(s string) string {
	dic := map[string]string{
		"？": "?",
		"，": ",",
		"。": ".",
		"：": ":",
		"；": ";",
		"“": "\"",
		"”": "\"",
		"‘": "'",
		"’": "'",
		"（": "(",
		"）": ")",
		"【": "[",
		"】": "]",
		"！": "!",
	}
	r := []rune(s)
	for i, v := range r {
		if val, ok := dic[string(v)]; ok {
			r[i] = []rune(val)[0]
		}
	}
	return string(r)
}

func (m *MassProducer) elementAnalysis(sentence string) *card_maker.Elements {
	lastIndex := -1
	eles := card_maker.Elements{}
	if sentence == "" {
		return &eles
	}
	r := []rune(sentence)
	for i, v := range r {
		if slices.Contains(card_maker.AllElements, string(v)) {
			num, err := strconv.Atoi(string(r[lastIndex+1 : i]))
			lastIndex = i
			if err != nil {
				continue
			}
			eles[string(v)] = num
		}
	}
	return &eles
}

type Row map[string]string

const (
	ERR_SKIP = iota
)

func (m *MassProducer) getCardInfoFromRow(row Row, versionName string) (*card_maker.CardInfo, error) {
	cardInfo := &card_maker.CardInfo{}
	var err error
	if len(row) == 0 {
		return nil, fmt.Errorf("%v", ERR_SKIP)
	}
	if _, ok := row["编号"]; ok {
		if row["编号"] == "" {
			return nil, fmt.Errorf("%v", ERR_SKIP)
		}
		if len([]rune(row["编号"])) != 7 {
			return nil, fmt.Errorf("number should be 7 digits")
		}
		cardInfo.Number = row["编号"]
	} else {
		return nil, fmt.Errorf("%v", ERR_SKIP)
	}
	if _, ok := row["属性"]; ok {
		cardInfo.Category = m.cleanString(row["属性"])
	} else {
		return nil, fmt.Errorf("category is required")
	}
	if _, ok := row["名称"]; ok {
		cardInfo.Name = m.cleanString(row["名称"])
	} else {
		return nil, fmt.Errorf("name is required")
	}
	if _, ok := row["种类"]; ok {
		cardInfo.Tag = m.cleanString(row["种类"])
	}
	if _, ok := row["标签"]; ok {
		cardInfo.Tag = m.cleanString(row["标签"])
	}
	if _, ok := row["生命"]; ok {
		cardInfo.Life, err = strconv.Atoi(m.cleanString(row["生命"]))
		if err != nil {
			cardInfo.Life = -1
		}
	} else {
		cardInfo.Life = -1
	}
	if _, ok := row["条件"]; ok {
		cardInfo.ElementsCost = *m.elementAnalysis(m.cleanString(row["条件"]))
	} else {
		cardInfo.ElementsCost = card_maker.Elements{}
	}
	if _, ok := row["负载"]; ok {
		cardInfo.ElementsGain = *m.elementAnalysis(m.cleanString(row["负载"]))
	} else {
		cardInfo.ElementsGain = card_maker.Elements{}
	}
	if _, ok := row["效果"]; ok {
		cardInfo.Description = m.cleanString(row["效果"])
	}
	if _, ok := row["引言"]; ok {
		cardInfo.Quote = m.cleanString(row["引言"])
	}
	if _, ok := row["威力"]; ok {
		cardInfo.Power, err = strconv.Atoi(m.cleanString(row["威力"]))
		if err != nil {
			cardInfo.Power = -1
		}
	} else {
		cardInfo.Power = -1
	}
	if _, ok := row["时间"]; ok {
		cardInfo.Duration, err = strconv.Atoi(m.cleanString(row["时间"]))
		if err != nil {
			cardInfo.Duration = -1
		}
	} else {
		cardInfo.Duration = -1
	}
	if _, ok := row["代价"]; ok {
		cardInfo.ElementsExpense = *m.elementAnalysis(m.cleanString(row["代价"]))
	} else {
		cardInfo.ElementsExpense = card_maker.Elements{}
	}
	if _, ok := row["攻击"]; ok {
		cardInfo.Attack, err = strconv.Atoi(m.cleanString(row["攻击"]))
		if err != nil {
			cardInfo.Attack = -1
		}
	} else {
		cardInfo.Attack = -1
	}
	if _, ok := row["衍生"]; ok {
		cleanStr := m.cleanString(row["衍生"])
		//  split by whitespace
		spawns := strings.Split(cleanStr, " ")
		cardInfo.Spawns = []string{}
		for _, spawn := range spawns {
			// if it only contains whitespace, skip it
			if spawn == "" {
				continue
			}

			// if it can be parsed as int
			// or it can be parsed as float

			if _, err := strconv.Atoi(spawn); err == nil {
				cardInfo.Spawns = append(cardInfo.Spawns, spawn)
			} else if parsedSpawn, err := strconv.ParseFloat(spawn, 64); err == nil {
				intSpawn := int(parsedSpawn)
				cardInfo.Spawns = append(cardInfo.Spawns, strconv.Itoa(intSpawn))
			} else {
				log.Printf("%s %s has invalid spawn list %s\n", cardInfo.Number, cardInfo.Name, cleanStr)
				continue
			}
		}
	} else {
		cardInfo.Spawns = []string{}
	}

	cardInfo.VersionNumber = string([]rune(cardInfo.Number)[3:4])
	cardInfo.VersionName = versionName
	tmpCardMaker := &card_maker.CardMaker{}
	cardInfo.Type = tmpCardMaker.GetCardType(cardInfo)
	cardInfo.OutputPath = m.GetOutputPath(cardInfo)
	return cardInfo, nil
}

func (m *MassProducer) GetOutputPath(cardInfo *card_maker.CardInfo) string {
	return filepath.Join(
		OUTPUT_PATH,
		cardInfo.VersionName,
		cardInfo.Type,
		cardInfo.Category,
		cardInfo.Number+".jpg",
	)
}

func (m *MassProducer) parseSheet(rows [][]string) []Row {
	result := []Row{}
	header := rows[0]
	for i, row := range rows {
		if i == 0 {
			continue
		}
		r := Row{}
		for j, cell := range row {
			r[header[j]] = cell
		}
		result = append(result, r)
	}
	return result
}

func (m *MassProducer) rgbaToCmyk(img *image.RGBA) *image.CMYK {
	bounds := img.Bounds()
	cmykImg := image.NewCMYK(bounds)

	// Iterate over every pixel
	for by := bounds.Min.Y; by < bounds.Max.Y; by++ {
		for bx := bounds.Min.X; bx < bounds.Max.X; bx++ {
			r, g, b, _ := img.At(bx, by).RGBA()
			c, m, y, k := color.RGBToCMYK(uint8(r>>8), uint8(g>>8), uint8(b>>8))
			cmykImg.Set(bx, by, color.CMYK{C: c, M: m, Y: y, K: k})
		}
	}

	return cmykImg
}

func (m *MassProducer) saveCmykAsTiff(img *image.CMYK, filename string) error {
	// Create output file
	_, err := os.Stat(filename)
	if err == nil && !m.Params.Overwrite {
		return fmt.Errorf("file %s already exists", filename)
	}
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Encode and save the image as TIFF
	err = tiff.Encode(outFile, img, nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *MassProducer) SaveRgbaAsJpeg(img *image.RGBA, filename string) error {
	// Create output file
	_, err := os.Stat(filename)
	if err == nil && !m.Params.Overwrite {
		return fmt.Errorf("file %s already exists", filename)
	}

	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Set JPEG quality (from 1 to 100, where 100 is the best quality)
	options := &jpeg.Options{Quality: 100}

	// Encode the RGBA image as a JPEG and save to the file
	err = jpeg.Encode(outFile, img, options)
	if err != nil {
		return err
	}

	return nil
}

func (m *MassProducer) addLog(log string) {
	m.LogMu.Lock()
	m.Log = append(m.Log, log)
	m.LogMu.Unlock()
}

func (m *MassProducer) dealWithSheet(xlsxFile *excelize.File, sheetName, versionName, drawingPath string) {
	newConfig := m.Config.Copy()
	newConfig.DrawingPath = drawingPath

	cardMaker := card_maker.NewCardMaker(*newConfig)
	rawRows, err := xlsxFile.GetRows(sheetName)
	if err != nil {
		log.Fatalf("get rows failed when reading sheet %v", sheetName)
		return
	}
	rows := m.parseSheet(rawRows)

	for _, row := range rows {
		cardInfo, err := m.getCardInfoFromRow(row, versionName)
		if err != nil {
			if err.Error() == fmt.Sprintf("%v", ERR_SKIP) {
				continue
			} else {
				m.addLog(fmt.Sprintf("get card info from row failed %v, row: %v\n", err, row))
				continue
			}
		}
		// should I draw this
		if m.OldAllCardsInfos != nil && m.Params.NewCardsOnly {
			_, err := os.Stat(cardInfo.OutputPath)
			if ci, ok := m.OldAllCardsInfos[cardInfo.Number]; ok {
				if ci.Equals(*cardInfo) && err == nil {
					// already exists, add it to all card infos, skip it
					m.Mu.Lock()
					m.AllCardInfos = append(m.AllCardInfos, *cardInfo)
					m.Mu.Unlock()
					continue
				}
			}
		}

		// start drawing
		img, err := cardMaker.MakeCard(cardInfo)
		if err != nil {
			m.addLog(fmt.Sprintf("make card failed %v, card info: %v\n", err, cardInfo))
			continue
		}
		// get the dirs to the output path
		err = m.makeDir(filepath.Dir(cardInfo.OutputPath))
		if err != nil {
			log.Fatalf("make dir failed %v", err)
		}

		if m.Params.IsPrintingVersion {
			// convert to cmyk
			// change the extension to tiff
			cardInfo.OutputPath = strings.Replace(cardInfo.OutputPath, ".jpg", ".tiff", 1)
			cmykImg := m.rgbaToCmyk(img)
			err = m.saveCmykAsTiff(cmykImg, cardInfo.OutputPath)
			if err != nil {
				log.Fatalf("save cmyk as tiff failed %v", err)
				continue
			}
		} else {
			err = m.SaveRgbaAsJpeg(img, cardInfo.OutputPath)
			if err != nil {
				log.Fatalf("save rgba as jpeg failed %v", err)
				continue
			}
		}
		// add to all card infos
		m.Mu.Lock()
		m.AllCardInfos = append(m.AllCardInfos, *cardInfo)
		m.Mu.Unlock()

	}

}

func (m *MassProducer) dealWithXlsx(xlsxPath, drawingPath, versionName string, p *mpb.Progress) {
	xlsxFile, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		log.Fatalf("open file %s failed", xlsxPath)
		return
	}
	sheetList := xlsxFile.GetSheetList()
	bar := p.New(int64(len(sheetList)), // total value
		mpb.BarStyle().Lbound("|").Filler("█").Tip(">").Padding("░").Rbound("|"),
		mpb.PrependDecorators(
			decor.Name(versionName+" "+xlsxPath+": ", decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d/%d"),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)
	wg := sync.WaitGroup{}
	for _, sheetName := range sheetList {
		wg.Add(1)
		go func() {
			m.dealWithSheet(xlsxFile, sheetName, versionName, drawingPath)
			bar.Increment()
			defer wg.Done()
		}()
	}
	wg.Wait()
	bar.SetTotal(int64(len(sheetList)), true)

}

func (m *MassProducer) SanityCheck() {
	// this function checks whether all the images can be found
	fmt.Println("Sanity check ...")
	numberOfSheets := len(m.Params.XlsxPaths)
	wg := sync.WaitGroup{}

	for i := 0; i < numberOfSheets; i++ {
		xlsxPath := m.Params.XlsxPaths[i]
		drawingPath := m.Params.DrawingPaths[i]
		wg.Add(1)

		go func() {
			// m.dealWithXlsx(xlsxPath, drawingPath, versionName, p)
			xlsxFile, err := excelize.OpenFile(xlsxPath)
			if err != nil {
				log.Fatalf("open file %s failed", xlsxPath)
				return
			}

			sheetList := xlsxFile.GetSheetList()

			wg2 := sync.WaitGroup{}
			for _, sheetName := range sheetList {
				wg2.Add(1)
				go func() {
					// m.dealWithSheet(xlsxFile, sheetName, versionName, drawingPath)
					rawRows, err := xlsxFile.GetRows(sheetName)
					if err != nil {
						log.Fatalf("get rows failed when reading sheet %v", sheetName)
						return
					}
					rows := m.parseSheet(rawRows)
					for _, row := range rows {
						num, ok := row["编号"]
						if !ok {
							continue
						}
						if num == "" {
							continue
						}

						possibleExtentions := []string{".webp", ".jpg", ".png", ".jpeg", ".jfif", ".tiff"}
						found := false
						for _, ext := range possibleExtentions {
							// try to find the image in drawing path
							imgPath := filepath.Join(drawingPath, num+ext)
							_, err := os.Stat(imgPath)
							if err == nil {
								found = true
								break
							}
						}
						if !found {
							m.SanityMu.Lock()
							m.SanityCheckPassed = false
							m.SanityMu.Unlock()
							m.addLog(fmt.Sprintf("image not found for %s\n", num))
						}

					}
					defer wg2.Done()
				}()
			}
			wg2.Wait()

			defer wg.Done()
		}()
	}
	wg.Wait()

	if m.SanityCheckPassed {
		fmt.Println("Sanity check passed")
	} else {
		fmt.Println("Sanity check failed:")
		for _, log := range m.Log {
			fmt.Println(log)
		}
	}
}

func (m *MassProducer) Produce() {
	// xlsxFile, err := excelize.OpenFile()
	m.SanityCheck()
	if !m.SanityCheckPassed {
		return
	}
	numberOfSheets := len(m.Params.XlsxPaths)
	wg := sync.WaitGroup{}
	p := mpb.New(mpb.WithWidth(64))

	for i := 0; i < numberOfSheets; i++ {
		xlsxPath := m.Params.XlsxPaths[i]
		drawingPath := m.Params.DrawingPaths[i]
		versionName := m.Params.VersionNames[i]
		wg.Add(1)

		go func() {
			m.dealWithXlsx(xlsxPath, drawingPath, versionName, p)
			defer wg.Done()
		}()
	}
	wg.Wait()
	p.Wait()
	// save all card infos
	jsonBytes, err := json.MarshalIndent(m.AllCardInfos, "", "  ")
	if err != nil {
		log.Fatalf("marshal all card infos failed %v", err)
	}
	err = os.WriteFile(filepath.Join(OUTPUT_PATH, "all_card_infos.json"), jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatalf("write all card infos failed %v", err)
	}
	// create simplified cards infos
	simplifiedCardInfos := map[string]string{}
	for _, cardInfo := range m.AllCardInfos {
		simplifiedCardInfos[cardInfo.Number] = cardInfo.OutputPath
	}
	jsonBytes, err = json.MarshalIndent(simplifiedCardInfos, "", "  ")
	if err != nil {
		log.Fatalf("marshal simplified card infos failed %v", err)
	}
	err = os.WriteFile(filepath.Join(OUTPUT_PATH, "simplified_card_infos.json"), jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatalf("write simplified card infos failed %v", err)
	}
	// print log
	for _, log := range m.Log {
		fmt.Println(log)
	}
}
