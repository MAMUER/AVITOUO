package storage

import (
	"archive/zip"
	"fmt"
	"image"
	"io"
	"math/rand"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/xuri/excelize/v2"
)

const (
	PhotosDir                = "photos"
	TargetActionHeader       = "Настройка цены целевого действия"
	TargetActionManualHeader = "Настройка цены целевого действия: ручная"
	ProductTypeHeader        = "вид товара"
)

func isImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

// PhotoGenerator генерирует уникальные копии изображений
type PhotoGenerator struct{}

// GenerateUniquePhotos создает N наборов фото из папки.
// Для каждого набора берется ВСЕ фото из папки (до 10 штук), и КАЖДОЕ фото уникализируется.
// Возвращает слайс строк, где каждая строка — имена фото для одного объявления через " | ".
func (pg *PhotoGenerator) GenerateUniquePhotos(sourceDir string, count int) ([]string, error) {
	fullDir := filepath.Join(PhotosDir, filepath.Clean(sourceDir))

	entries, err := os.ReadDir(fullDir)
	if err != nil {
		return nil, fmt.Errorf("папка не найдена: %w", err)
	}

	var sourceImages []string
	for _, entry := range entries {
		if !entry.IsDir() && isImage(entry.Name()) {
			sourceImages = append(sourceImages, filepath.Join(fullDir, entry.Name()))
		}
	}

	if len(sourceImages) == 0 {
		return nil, fmt.Errorf("в папке нет изображений")
	}

	var result []string
	globalPhotoIdx := 0

	for adIdx := 0; adIdx < count; adIdx++ {
		var names []string
		maxPhotos := len(sourceImages)
		if maxPhotos > 10 {
			maxPhotos = 10
		}
		for photoIdx := 0; photoIdx < maxPhotos; photoIdx++ {
			srcPath := sourceImages[photoIdx]
			ext := strings.ToLower(filepath.Ext(srcPath))
			baseName := fmt.Sprintf("a%d%s", globalPhotoIdx+1, ext)
			savePath := filepath.Join(fullDir, baseName)

			if err := processOnePhoto(srcPath, savePath, globalPhotoIdx); err != nil {
				return nil, err
			}
			names = append(names, baseName)
			globalPhotoIdx++
		}
		result = append(result, strings.Join(names, " | "))
	}

	return result, nil
}

func processOnePhoto(srcPath, savePath string, index int) error {
	srcData, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения фото: %w", err)
	}

	img, err := imaging.Decode(strings.NewReader(string(srcData)))
	if err != nil {
		return fmt.Errorf("ошибка декодирования изображения: %w", err)
	}

	uniqueImg := applyUniqueTransformations(img, index)
	if err := imaging.Save(uniqueImg, savePath); err != nil {
		return fmt.Errorf("ошибка сохранения уникального фото: %w", err)
	}

	return nil
}

// applyUniqueTransformations применяет уникальные трансформации к изображению
func applyUniqueTransformations(img image.Image, index int) image.Image {
	if img == nil {
		return nil
	}

	brightness := float64((index%20)-10) / 120.0
	contrast := float64((index%15)-7) / 120.0
	return imaging.AdjustBrightness(imaging.AdjustContrast(img, contrast), brightness)
}

func findHeaderRow(rows [][]string) int {
	for i := 0; i < len(rows) && i < 20; i++ {
		if IsHeaderRow(rows[i]) {
			return i
		}
	}
	return 0
}

func findServiceColumns(headerRow []string) (int, int, int) {
	targetActionColIdx := -1
	targetActionManualColIdx := -1
	productTypeColIdx := -1
	for i, h := range headerRow {
		if strings.Contains(strings.ToLower(h), strings.ToLower(TargetActionHeader)) && !strings.Contains(strings.ToLower(h), "ручная") {
			targetActionColIdx = i
		}
		if strings.Contains(strings.ToLower(h), strings.ToLower(TargetActionManualHeader)) {
			targetActionManualColIdx = i
		}
		if strings.Contains(strings.ToLower(h), ProductTypeHeader) {
			productTypeColIdx = i
		}
	}
	return targetActionColIdx, targetActionManualColIdx, productTypeColIdx
}

func ensureServiceColumns(f *excelize.File, sheetName string, headerRowIdx int, headerRow *[]string, targetActionColIdx, targetActionManualColIdx *int) error {
	if *targetActionColIdx < 0 {
		*targetActionColIdx = len(*headerRow)
		*headerRow = append(*headerRow, TargetActionHeader)
		colName, err := excelize.ColumnNumberToName(*targetActionColIdx + 1)
		if err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: %w", err)
		}
		cell := fmt.Sprintf("%s%d", colName, headerRowIdx+1)
		if err := f.SetCellValue(sheetName, cell, TargetActionHeader); err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: %w", err)
		}
	}

	if *targetActionManualColIdx < 0 {
		*targetActionManualColIdx = len(*headerRow)
		colName, err := excelize.ColumnNumberToName(*targetActionManualColIdx + 1)
		if err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: ручная: %w", err)
		}
		cell := fmt.Sprintf("%s%d", colName, headerRowIdx+1)
		if err := f.SetCellValue(sheetName, cell, TargetActionManualHeader); err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: ручная: %w", err)
		}
	}
	return nil
}

func resolveServiceRowValue(productType, value string) string {
	productLower := strings.ToLower(productType)
	if strings.Contains(productLower, "окна") || strings.Contains(productLower, "балкон") ||
		strings.Contains(productLower, "двери") || strings.Contains(productLower, "дверь") ||
		strings.Contains(productLower, "баня") || strings.Contains(productLower, "сауна") || strings.Contains(productLower, "бассейн") {
		return "Москва|8|1000\nМосковская область|8|1000"
	}
	return value
}

func writeServiceCell(f *excelize.File, sheetName string, colIdx, rowIdx int, text string) {
	if colIdx < 0 {
		return
	}
	colName, err := excelize.ColumnNumberToName(colIdx + 1)
	if err != nil {
		return
	}
	_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, rowIdx+1), text)
}

func processSheetForServices(f *excelize.File, sheetName string, rows [][]string, targetActionColIdx, targetActionManualColIdx, productTypeColIdx int, value string) error {
	for rowIdx := 4; rowIdx < len(rows); rowIdx++ {
		productType := ""
		if productTypeColIdx >= 0 && productTypeColIdx < len(rows[rowIdx]) {
			productType = strings.TrimSpace(rows[rowIdx][productTypeColIdx])
		}
		rowValue := resolveServiceRowValue(productType, value)
		writeServiceCell(f, sheetName, targetActionColIdx, rowIdx, "Manual")
		writeServiceCell(f, sheetName, targetActionManualColIdx, rowIdx, rowValue)
	}
	return nil
}

// AddServices создает столбцы услуг (если их нет) и заполняет их значением для всех строк данных (начиная с 5-й)
func AddServices(templatePath, outputPath string, value string) error {
	if value == "" {
		return fmt.Errorf("значение для услуг не указано")
	}

	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("в файле нет листов")
	}

	for _, sheetName := range sheets {
		if strings.EqualFold(sheetName, "Инструкция") {
			continue
		}
		rows, err := f.GetRows(sheetName)
		if err != nil || len(rows) == 0 {
			continue
		}

		headerRowIdx := findHeaderRow(rows)
		headerRow := rows[headerRowIdx]

		targetActionColIdx, targetActionManualColIdx, productTypeColIdx := findServiceColumns(headerRow)
		if err := ensureServiceColumns(f, sheetName, headerRowIdx, &headerRow, &targetActionColIdx, &targetActionManualColIdx); err != nil {
			return err
		}
		if err := processSheetForServices(f, sheetName, rows, targetActionColIdx, targetActionManualColIdx, productTypeColIdx, value); err != nil {
			return err
		}
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("ошибка сохранения файла: %w", err)
	}

	return nil
}

// CreatePhotoZip создает ZIP-архив со всеми фото из папки (без вложенных папок)
func CreatePhotoZip(sourceDir string, zipPath string) (string, error) {
	fullDir := filepath.Join(PhotosDir, filepath.Clean(sourceDir))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := zipFile.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	entries, err := os.ReadDir(fullDir)
	if err != nil {
		return "", err
	}

	var fileNames []string
	for _, entry := range entries {
		if entry.IsDir() || !isImage(entry.Name()) {
			continue
		}
		name, err := addImageToZip(zipWriter, fullDir, entry.Name())
		if err == nil {
			fileNames = append(fileNames, name)
		}
	}

	return strings.Join(fileNames, "|"), nil
}

func addImageToZip(zipWriter *zip.Writer, fullDir, entryName string) (string, error) {
	srcPath := filepath.Join(fullDir, entryName)
	fileToZip, err := os.Open(srcPath)
	if err != nil {
		return "", nil
	}
	defer func() { _ = fileToZip.Close() }()

	w, err := zipWriter.Create(entryName)
	if err != nil {
		return "", nil
	}

	if _, err := io.Copy(w, fileToZip); err != nil {
		return "", nil
	}
	return entryName, nil
}

// CheckTotalSize проверяет общий размер файлов
func CheckTotalSize(zipPath, xlsxPath string) error {
	const maxSize = 100 * 1024 * 1024
	var totalSize int64
	if info, err := os.Stat(zipPath); err == nil {
		totalSize += info.Size()
	}
	if info, err := os.Stat(xlsxPath); err == nil {
		totalSize += info.Size()
	}
	if totalSize > maxSize {
		return fmt.Errorf("общий размер файлов (%.2f МБ) превышает лимит в 100 МБ", float64(totalSize)/(1024*1024))
	}
	return nil
}

// CheckSizeLimit проверяет, не превысит ли итоговый размер лимит
func CheckSizeLimit(photoCount int, estimatePhotoSize int64) error {
	const maxSize = 100 * 1024 * 1024
	estimatedSize := int64(photoCount) * estimatePhotoSize
	estimatedSize += int64(photoCount) * 1024

	if estimatedSize > maxSize {
		return fmt.Errorf("превышен лимит в 100 МБ (примерный размер: %.2f МБ). Уменьшите количество вариантов", float64(estimatedSize)/(1024*1024))
	}
	return nil
}

// SaveExcelParams содержит параметры для SaveExcelWithNewRows
type SaveExcelParams struct {
	TemplatePath string
	OutputPath   string
	SheetName    string

	TitleColIdx                   int
	DescColIdx                    int
	ImageNamesIdx                 int
	ContactColIdx                 int
	PhoneColIdx                   int
	AddressColIdx                 int
	CompanyColIdx                 int
	EmailColIdx                   int
	NewTitles                     []string
	NewDescriptions               []string
	NewImageNames                 []string
	NewContacts                   []string
	NewPhones                     []string
	NewAddresses                  []string
	NewCompanies                  []string
	NewEmails                     []string
	IDColIdx                      int
	PlacementColIdx               int
	ContactMethodColIdx           int
	CategoryColIdx                int
	ProductTypeColIdx             int
	SubProductTypeColIdx          int
	PriceUnitColIdx               int
	ConditionColIdx               int
	AvailabilityColIdx            int
	AdTypeColIdx                  int
	SalesTypeColIdx               int
	ConnectColIdx                 int
	ProcessingColIdx              int
	PurposeColIdx                 int
	GOSTColIdx                    int
	NewIDs                        []string
	NewPlacements                 []string
	NewContactMethods             []string
	NewCategories                 []string
	NewProductTypes               []string
	NewSubProductTypes            []string
	NewPriceUnits                 []string
	NewConditions                 []string
	NewAvailabilities             []string
	NewAdTypes                    []string
	NewSalesTypes                 []string
	NewConnects                   []string
	NewProcessing                 []string
	NewPurpose                    []string
	NewLumberTypes                []string
	NewWoodTypes                  []string
	NewEdges                      []string
	NewGrades                     []string
	NewMoistures                  []string
	NewProfiles                   []string
	NewStructures                 []string
	NewLumberProfiles             []string
	NewThicknesses                []string
	NewWidths                     []string
	NewLengths                    []string
	NewHeights                    []string
	NewWidthDs                    []string
	NewLengthDs                   []string
	NewGOSTValues                 []string
	TargetActionColIdx            int
	TargetActionManualColIdx      int
	NewTargetActionManual         []string
	NewTargetActionManualSettings []string
	DiameterColIdx                int
	NewDiameters                  []string
	PriceValueColIdx              int
	NewPriceValues                []string
}

// SaveExcelWithNewRows добавляет новые строки в Excel файл и сохраняет

func buildColumnFinder(headerRow []string, usedIdxs map[int]bool) func(names ...string) int {
	return func(names ...string) int {
		for i, h := range headerRow {
			if usedIdxs[i] {
				continue
			}
			lower := strings.ToLower(h)
			for _, name := range names {
				if strings.Contains(lower, strings.ToLower(name)) {
					usedIdxs[i] = true
					return i
				}
			}
		}
		return -1
	}
}

func markUsedColumns(p *SaveExcelParams, markUsed func(int)) {
	markUsed(p.TitleColIdx)
	markUsed(p.DescColIdx)
	markUsed(p.ImageNamesIdx)
	markUsed(p.ContactColIdx)
	markUsed(p.PhoneColIdx)
	markUsed(p.AddressColIdx)
	markUsed(p.CompanyColIdx)
	markUsed(p.EmailColIdx)
	markUsed(p.IDColIdx)
	markUsed(p.PlacementColIdx)
	markUsed(p.ContactMethodColIdx)
	markUsed(p.CategoryColIdx)
	markUsed(p.ProductTypeColIdx)
	markUsed(p.SubProductTypeColIdx)
	markUsed(p.PriceUnitColIdx)
	markUsed(p.PriceValueColIdx)
	markUsed(p.ConditionColIdx)
	markUsed(p.AvailabilityColIdx)
	markUsed(p.AdTypeColIdx)
	markUsed(p.SalesTypeColIdx)
	markUsed(p.ConnectColIdx)
	markUsed(p.ProcessingColIdx)
	markUsed(p.PurposeColIdx)
	markUsed(p.GOSTColIdx)
	markUsed(p.TargetActionColIdx)
	markUsed(p.TargetActionManualColIdx)
	markUsed(p.DiameterColIdx)
}

func fallbackScanColumns(p *SaveExcelParams, findUnique func(names ...string) int, headerRowIdx int, headerRow []string) {
	fmt.Printf("[DEBUG] Fallback scan on header row %d: %v\n", headerRowIdx, headerRow)
	if p.TitleColIdx < 0 {
		p.TitleColIdx = findUnique("title", "название", "заголовок")
	}
	if p.DescColIdx < 0 {
		p.DescColIdx = findUnique("description", "описание", "текст")
	}
	if p.ImageNamesIdx < 0 {
		p.ImageNamesIdx = findUnique("image", "фото", "изображение")
	}
	if p.ContactColIdx < 0 {
		p.ContactColIdx = findUnique("контактное лицо", "контакт", "contact")
	}
	if p.PhoneColIdx < 0 {
		p.PhoneColIdx = findUnique("номер телефона", "телефон", "phone")
	}
	if p.AddressColIdx < 0 {
		p.AddressColIdx = findUnique("адрес", "address")
	}
	if p.CompanyColIdx < 0 {
		p.CompanyColIdx = findUnique("название компании", "компания", "организация", "company")
	}
	if p.EmailColIdx < 0 {
		p.EmailColIdx = findUnique("почта", "email", "e-mail", "электронная почта")
	}
	if p.IDColIdx < 0 {
		p.IDColIdx = findUnique("уникальный идентификатор", "id")
	}
	if p.PlacementColIdx < 0 {
		p.PlacementColIdx = findUnique("способ размещения", "размещения")
	}
	if p.ContactMethodColIdx < 0 {
		p.ContactMethodColIdx = findUnique("способ связи", "связи")
	}
	if p.CategoryColIdx < 0 {
		p.CategoryColIdx = findUnique("категория")
	}
	if p.ProductTypeColIdx < 0 {
		p.ProductTypeColIdx = findUnique(ProductTypeHeader, "товара")
	}
	if p.SubProductTypeColIdx < 0 {
		p.SubProductTypeColIdx = findUnique("подвид товара", "подвид")
	}
	if p.PriceUnitColIdx < 0 {
		p.PriceUnitColIdx = findUnique("цена за", "единица")
	}
	if p.PriceValueColIdx < 0 {
		p.PriceValueColIdx = findUnique("цена")
	}
	if p.ConditionColIdx < 0 {
		p.ConditionColIdx = findUnique("состояние")
	}
	if p.AvailabilityColIdx < 0 {
		p.AvailabilityColIdx = findUnique("доступность")
	}
	if p.AdTypeColIdx < 0 {
		p.AdTypeColIdx = findUnique("вид объявления")
	}
	if p.SalesTypeColIdx < 0 {
		p.SalesTypeColIdx = findUnique("вид продажи", "продажи")
	}
	if p.ConnectColIdx < 0 {
		p.ConnectColIdx = findUnique("соединять")
	}
	if p.ProcessingColIdx < 0 {
		p.ProcessingColIdx = findUnique("обработка")
	}
	if p.PurposeColIdx < 0 {
		p.PurposeColIdx = findUnique("назначение")
	}
	fmt.Printf("[DEBUG] Fallback result: titleIdx=%d descIdx=%d imageIdx=%d contactIdx=%d phoneIdx=%d addressIdx=%d companyIdx=%d emailIdx=%d idIdx=%d placementIdx=%d methodIdx=%d categoryIdx=%d productIdx=%d subProductIdx=%d priceUnitIdx=%d priceValueIdx=%d conditionIdx=%d availabilityIdx=%d adTypeIdx=%d salesTypeIdx=%d connectIdx=%d processingIdx=%d purposeIdx=%d\n", p.TitleColIdx, p.DescColIdx, p.ImageNamesIdx, p.ContactColIdx, p.PhoneColIdx, p.AddressColIdx, p.CompanyColIdx, p.EmailColIdx, p.IDColIdx, p.PlacementColIdx, p.ContactMethodColIdx, p.CategoryColIdx, p.ProductTypeColIdx, p.SubProductTypeColIdx, p.PriceUnitColIdx, p.PriceValueColIdx, p.ConditionColIdx, p.AvailabilityColIdx, p.AdTypeColIdx, p.SalesTypeColIdx, p.ConnectColIdx, p.ProcessingColIdx, p.PurposeColIdx)
}

func setDefaultColumnIndices(p *SaveExcelParams, rows [][]string) {
	if p.TitleColIdx < 0 {
		p.TitleColIdx = 0
	}
	if p.DescColIdx < 0 {
		p.DescColIdx = 1
	}
	if p.ImageNamesIdx < 0 {
		if len(rows) > 0 && len(rows[0]) > 2 {
			p.ImageNamesIdx = 2
		} else {
			p.ImageNamesIdx = -1
		}
	}
	if p.IDColIdx < 0 {
		p.IDColIdx = -1
	}
	if p.PlacementColIdx < 0 {
		p.PlacementColIdx = -1
	}
	if p.ContactMethodColIdx < 0 {
		p.ContactMethodColIdx = -1
	}
	if p.CategoryColIdx < 0 {
		p.CategoryColIdx = -1
	}
	if p.ProductTypeColIdx < 0 {
		p.ProductTypeColIdx = -1
	}
	if p.SubProductTypeColIdx < 0 {
		p.SubProductTypeColIdx = -1
	}
	if p.PriceUnitColIdx < 0 {
		p.PriceUnitColIdx = -1
	}
	if p.PriceValueColIdx < 0 {
		p.PriceValueColIdx = -1
	}
	if p.ConditionColIdx < 0 {
		p.ConditionColIdx = -1
	}
	if p.AvailabilityColIdx < 0 {
		p.AvailabilityColIdx = -1
	}
	if p.AdTypeColIdx < 0 {
		p.AdTypeColIdx = -1
	}
	if p.SalesTypeColIdx < 0 {
		p.SalesTypeColIdx = -1
	}
	if p.ConnectColIdx < 0 {
		p.ConnectColIdx = -1
	}
	if p.ProcessingColIdx < 0 {
		p.ProcessingColIdx = -1
	}
	if p.PurposeColIdx < 0 {
		p.PurposeColIdx = -1
	}
}

func SaveExcelWithNewRows(p *SaveExcelParams) error {
	f, err := excelize.OpenFile(p.TemplatePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия шаблона: %w", err)
	}

	rows, err := f.GetRows(p.SheetName)
	if err != nil {
		return fmt.Errorf("ошибка чтения листа: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("лист пустой")
	}

	headerRowIdx := -1
	for i := 0; i < len(rows) && i < 20; i++ {
		if IsHeaderRow(rows[i]) {
			headerRowIdx = i
			break
		}
	}
	if headerRowIdx < 0 {
		headerRowIdx = 0
	}
	headerRow := rows[headerRowIdx]

	usedIdxs := make(map[int]bool)
	markUsed := func(idx int) {
		if idx >= 0 {
			usedIdxs[idx] = true
		}
	}
	findUnique := func(names ...string) int {
		for i, h := range headerRow {
			if usedIdxs[i] {
				continue
			}
			lower := strings.ToLower(h)
			for _, name := range names {
				if strings.Contains(lower, strings.ToLower(name)) {
					usedIdxs[i] = true
					return i
				}
			}
		}
		return -1
	}

	markUsed(p.TitleColIdx)
	markUsed(p.DescColIdx)
	markUsed(p.ImageNamesIdx)
	markUsed(p.ContactColIdx)
	markUsed(p.PhoneColIdx)
	markUsed(p.AddressColIdx)
	markUsed(p.CompanyColIdx)
	markUsed(p.EmailColIdx)
	markUsed(p.IDColIdx)
	markUsed(p.PlacementColIdx)
	markUsed(p.ContactMethodColIdx)
	markUsed(p.CategoryColIdx)
	markUsed(p.ProductTypeColIdx)
	markUsed(p.SubProductTypeColIdx)
	markUsed(p.PriceUnitColIdx)
	markUsed(p.PriceValueColIdx)
	markUsed(p.ConditionColIdx)
	markUsed(p.AvailabilityColIdx)
	markUsed(p.AdTypeColIdx)
	markUsed(p.SalesTypeColIdx)
	markUsed(p.ConnectColIdx)
	markUsed(p.ProcessingColIdx)
	markUsed(p.PurposeColIdx)
	markUsed(p.GOSTColIdx)
	markUsed(p.TargetActionColIdx)
	markUsed(p.TargetActionManualColIdx)
	markUsed(p.DiameterColIdx)

	if p.TitleColIdx < 0 || p.DescColIdx < 0 || p.ImageNamesIdx < 0 || p.ContactColIdx < 0 || p.PhoneColIdx < 0 || p.AddressColIdx < 0 || p.CompanyColIdx < 0 || p.EmailColIdx < 0 || p.IDColIdx < 0 || p.PlacementColIdx < 0 || p.ContactMethodColIdx < 0 || p.CategoryColIdx < 0 || p.ProductTypeColIdx < 0 || p.SubProductTypeColIdx < 0 {
		fmt.Printf("[DEBUG] Fallback scan on header row %d: %v\n", headerRowIdx, headerRow)
		if p.TitleColIdx < 0 {
			p.TitleColIdx = findUnique("title", "название", "заголовок")
		}
		if p.DescColIdx < 0 {
			p.DescColIdx = findUnique("description", "описание", "текст")
		}
		if p.ImageNamesIdx < 0 {
			p.ImageNamesIdx = findUnique("image", "фото", "изображение")
		}
		if p.ContactColIdx < 0 {
			p.ContactColIdx = findUnique("контактное лицо", "контакт", "contact")
		}
		if p.PhoneColIdx < 0 {
			p.PhoneColIdx = findUnique("номер телефона", "телефон", "phone")
		}
		if p.AddressColIdx < 0 {
			p.AddressColIdx = findUnique("адрес", "address")
		}
		if p.CompanyColIdx < 0 {
			p.CompanyColIdx = findUnique("название компании", "компания", "организация", "company")
		}
		if p.EmailColIdx < 0 {
			p.EmailColIdx = findUnique("почта", "email", "e-mail", "электронная почта")
		}
		if p.IDColIdx < 0 {
			p.IDColIdx = findUnique("уникальный идентификатор", "id")
		}
		if p.PlacementColIdx < 0 {
			p.PlacementColIdx = findUnique("способ размещения", "размещения")
		}
		if p.ContactMethodColIdx < 0 {
			p.ContactMethodColIdx = findUnique("способ связи", "связи")
		}
		if p.CategoryColIdx < 0 {
			p.CategoryColIdx = findUnique("категория")
		}
		if p.ProductTypeColIdx < 0 {
			p.ProductTypeColIdx = findUnique(ProductTypeHeader, "товара")
		}
		if p.SubProductTypeColIdx < 0 {
			p.SubProductTypeColIdx = findUnique("подвид товара", "подвид")
		}
		if p.PriceUnitColIdx < 0 {
			p.PriceUnitColIdx = findUnique("цена за", "единица")
		}
		if p.PriceValueColIdx < 0 {
			p.PriceValueColIdx = findUnique("цена")
		}
		if p.ConditionColIdx < 0 {
			p.ConditionColIdx = findUnique("состояние")
		}
		if p.AvailabilityColIdx < 0 {
			p.AvailabilityColIdx = findUnique("доступность")
		}
		if p.AdTypeColIdx < 0 {
			p.AdTypeColIdx = findUnique("вид объявления")
		}
		if p.SalesTypeColIdx < 0 {
			p.SalesTypeColIdx = findUnique("вид продажи", "продажи")
		}
		if p.ConnectColIdx < 0 {
			p.ConnectColIdx = findUnique("соединять")
		}
		if p.ProcessingColIdx < 0 {
			p.ProcessingColIdx = findUnique("обработка")
		}
		if p.PurposeColIdx < 0 {
			p.PurposeColIdx = findUnique("назначение")
		}
		fmt.Printf("[DEBUG] Fallback result: titleIdx=%d descIdx=%d imageIdx=%d contactIdx=%d phoneIdx=%d addressIdx=%d companyIdx=%d emailIdx=%d idIdx=%d placementIdx=%d methodIdx=%d categoryIdx=%d productIdx=%d subProductIdx=%d priceUnitIdx=%d priceValueIdx=%d conditionIdx=%d availabilityIdx=%d adTypeIdx=%d salesTypeIdx=%d connectIdx=%d processingIdx=%d purposeIdx=%d\n", p.TitleColIdx, p.DescColIdx, p.ImageNamesIdx, p.ContactColIdx, p.PhoneColIdx, p.AddressColIdx, p.CompanyColIdx, p.EmailColIdx, p.IDColIdx, p.PlacementColIdx, p.ContactMethodColIdx, p.CategoryColIdx, p.ProductTypeColIdx, p.SubProductTypeColIdx, p.PriceUnitColIdx, p.PriceValueColIdx, p.ConditionColIdx, p.AvailabilityColIdx, p.AdTypeColIdx, p.SalesTypeColIdx, p.ConnectColIdx, p.ProcessingColIdx, p.PurposeColIdx)
	}

	if p.TitleColIdx < 0 {
		p.TitleColIdx = 0
	}
	if p.DescColIdx < 0 {
		p.DescColIdx = 1
	}
	if p.ImageNamesIdx < 0 {
		if len(rows) > 0 && len(rows[0]) > 2 {
			p.ImageNamesIdx = 2
		} else {
			p.ImageNamesIdx = -1
		}
	}
	if p.IDColIdx < 0 {
		p.IDColIdx = -1
	}
	if p.PlacementColIdx < 0 {
		p.PlacementColIdx = -1
	}
	if p.ContactMethodColIdx < 0 {
		p.ContactMethodColIdx = -1
	}
	if p.CategoryColIdx < 0 {
		p.CategoryColIdx = -1
	}
	if p.ProductTypeColIdx < 0 {
		p.ProductTypeColIdx = -1
	}
	if p.SubProductTypeColIdx < 0 {
		p.SubProductTypeColIdx = -1
	}
	if p.PriceUnitColIdx < 0 {
		p.PriceUnitColIdx = -1
	}
	if p.PriceValueColIdx < 0 {
		p.PriceValueColIdx = -1
	}
	if p.ConditionColIdx < 0 {
		p.ConditionColIdx = -1
	}
	if p.AvailabilityColIdx < 0 {
		p.AvailabilityColIdx = -1
	}
	if p.AdTypeColIdx < 0 {
		p.AdTypeColIdx = -1
	}
	if p.SalesTypeColIdx < 0 {
		p.SalesTypeColIdx = -1
	}
	if p.ConnectColIdx < 0 {
		p.ConnectColIdx = -1
	}
	if p.ProcessingColIdx < 0 {
		p.ProcessingColIdx = -1
	}
	if p.PurposeColIdx < 0 {
		p.PurposeColIdx = -1
	}

	lumberTypeColIdx := -1
	woodTypeColIdx := -1
	edgeColIdx := -1
	gradeColIdx := -1
	moistureColIdx := -1
	profileColIdx := -1
	structureColIdx := -1
	lumberProfileColIdx := -1
	thicknessColIdx := -1
	widthColIdx := -1
	lengthColIdx := -1
	heightColIdx := -1
	widthDColIdx := -1
	lengthDColIdx := -1

	if len(headerRow) > 0 {
		if lumberTypeColIdx < 0 {
			lumberTypeColIdx = findUnique("тип пиломатериала")
		}
		if woodTypeColIdx < 0 {
			woodTypeColIdx = findUnique("вид древесины")
		}
		if edgeColIdx < 0 {
			edgeColIdx = findUnique("кромка")
		}
		if gradeColIdx < 0 {
			gradeColIdx = findUnique("сорт древесины")
		}
		if moistureColIdx < 0 {
			moistureColIdx = findUnique("степень влажности")
		}
		if profileColIdx < 0 {
			profileColIdx = findUnique("профилированный")
		}
		if structureColIdx < 0 {
			structureColIdx = findUnique("структура")
		}
		if lumberProfileColIdx < 0 {
			lumberProfileColIdx = findUnique("профиль")
		}
		if thicknessColIdx < 0 {
			thicknessColIdx = findUnique("толщина пиломатериала", "толщина")
		}
		if widthColIdx < 0 {
			widthColIdx = findUnique("ширина пиломатериала", "ширина бруса")
		}
		if lengthColIdx < 0 {
			lengthColIdx = findUnique("длина пиломатериала", "длина бруса")
		}
		if heightColIdx < 0 {
			heightColIdx = findUnique("высота")
		}
		if widthDColIdx < 0 {
			widthDColIdx = findUnique("ширина")
		}
		if lengthDColIdx < 0 {
			lengthDColIdx = findUnique("длина")
		}
		if p.TargetActionManualColIdx < 0 {
			p.TargetActionManualColIdx = findUnique(TargetActionManualHeader)
		}
		if p.TargetActionColIdx < 0 {
			p.TargetActionColIdx = findUnique(TargetActionHeader)
		}
		if p.DiameterColIdx < 0 {
			p.DiameterColIdx = findUnique("диаметр")
		}
		if p.PriceValueColIdx < 0 {
			p.PriceValueColIdx = findUnique("цена")
		}
	}

	if p.TargetActionColIdx < 0 {
		p.TargetActionColIdx = len(headerRow)
		headerRow = append(headerRow, TargetActionHeader)
		colName, err := excelize.ColumnNumberToName(p.TargetActionColIdx + 1)
		if err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: %w", err)
		}
		cell := fmt.Sprintf("%s%d", colName, headerRowIdx+1)
		if err := f.SetCellValue(p.SheetName, cell, TargetActionHeader); err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: %w", err)
		}
	}

	if p.TargetActionManualColIdx < 0 {
		p.TargetActionManualColIdx = len(headerRow)
		headerRow = append(headerRow, TargetActionManualHeader)
		colName, err := excelize.ColumnNumberToName(p.TargetActionManualColIdx + 1)
		if err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: ручная: %w", err)
		}
		cell := fmt.Sprintf("%s%d", colName, headerRowIdx+1)
		if err := f.SetCellValue(p.SheetName, cell, TargetActionManualHeader); err != nil {
			return fmt.Errorf("ошибка создания колонки целевого действия: ручная: %w", err)
		}
	}

	startRow := len(rows) + 1
	wrote := 0

	writeCol := func(colIdx int, value string) {
		if colIdx < 0 {
			return
		}
		colName, err := excelize.ColumnNumberToName(colIdx + 1)
		if err != nil {
			fmt.Printf("[DEBUG] Invalid column index %d: %v\n", colIdx, err)
			return
		}
		cell := fmt.Sprintf("%s%d", colName, startRow+wrote)
		if err := f.SetCellValue(p.SheetName, cell, value); err != nil {
			fmt.Printf("[DEBUG] SetCellValue error at %s: %v\n", cell, err)
		}
	}

	for i := 0; i < len(p.NewTitles); i++ {
		writeCol(p.TitleColIdx, p.NewTitles[i])
		writeCol(p.DescColIdx, p.NewDescriptions[i])
		if p.ImageNamesIdx >= 0 && i < len(p.NewImageNames) {
			writeCol(p.ImageNamesIdx, p.NewImageNames[i])
		}
		if p.ContactColIdx >= 0 && i < len(p.NewContacts) {
			writeCol(p.ContactColIdx, p.NewContacts[i])
		}
		if p.PhoneColIdx >= 0 && i < len(p.NewPhones) {
			writeCol(p.PhoneColIdx, p.NewPhones[i])
		}
		if p.AddressColIdx >= 0 && i < len(p.NewAddresses) {
			writeCol(p.AddressColIdx, p.NewAddresses[i])
		}
		if p.CompanyColIdx >= 0 && i < len(p.NewCompanies) {
			writeCol(p.CompanyColIdx, p.NewCompanies[i])
		}
		if p.EmailColIdx >= 0 && i < len(p.NewEmails) {
			writeCol(p.EmailColIdx, p.NewEmails[i])
		}
		if p.IDColIdx >= 0 && i < len(p.NewIDs) {
			writeCol(p.IDColIdx, p.NewIDs[i])
		}
		if p.PlacementColIdx >= 0 && i < len(p.NewPlacements) {
			writeCol(p.PlacementColIdx, p.NewPlacements[i])
		}
		if p.ContactMethodColIdx >= 0 && i < len(p.NewContactMethods) {
			writeCol(p.ContactMethodColIdx, p.NewContactMethods[i])
		}
		if p.CategoryColIdx >= 0 && i < len(p.NewCategories) {
			writeCol(p.CategoryColIdx, p.NewCategories[i])
		}
		if p.ProductTypeColIdx >= 0 && i < len(p.NewProductTypes) {
			writeCol(p.ProductTypeColIdx, p.NewProductTypes[i])
		}
		if p.SubProductTypeColIdx >= 0 && i < len(p.NewSubProductTypes) {
			writeCol(p.SubProductTypeColIdx, p.NewSubProductTypes[i])
		}
		if p.PriceUnitColIdx >= 0 && i < len(p.NewPriceUnits) {
			writeCol(p.PriceUnitColIdx, p.NewPriceUnits[i])
		}
		if p.ConditionColIdx >= 0 && i < len(p.NewConditions) {
			writeCol(p.ConditionColIdx, p.NewConditions[i])
		}
		if p.AvailabilityColIdx >= 0 && i < len(p.NewAvailabilities) {
			writeCol(p.AvailabilityColIdx, p.NewAvailabilities[i])
		}
		if p.AdTypeColIdx >= 0 && i < len(p.NewAdTypes) {
			writeCol(p.AdTypeColIdx, p.NewAdTypes[i])
		}
		if p.SalesTypeColIdx >= 0 && i < len(p.NewSalesTypes) {
			writeCol(p.SalesTypeColIdx, p.NewSalesTypes[i])
		}
		if p.ConnectColIdx >= 0 && i < len(p.NewConnects) {
			writeCol(p.ConnectColIdx, p.NewConnects[i])
		}
		if p.ProcessingColIdx >= 0 && i < len(p.NewProcessing) {
			writeCol(p.ProcessingColIdx, p.NewProcessing[i])
		}
		if p.PurposeColIdx >= 0 && i < len(p.NewPurpose) {
			writeCol(p.PurposeColIdx, p.NewPurpose[i])
		}
		if lumberTypeColIdx >= 0 && i < len(p.NewLumberTypes) {
			writeCol(lumberTypeColIdx, p.NewLumberTypes[i])
		}
		if woodTypeColIdx >= 0 && i < len(p.NewWoodTypes) {
			writeCol(woodTypeColIdx, p.NewWoodTypes[i])
		}
		if edgeColIdx >= 0 && i < len(p.NewEdges) {
			writeCol(edgeColIdx, p.NewEdges[i])
		}
		if gradeColIdx >= 0 && i < len(p.NewGrades) {
			writeCol(gradeColIdx, p.NewGrades[i])
		}
		if moistureColIdx >= 0 && i < len(p.NewMoistures) {
			writeCol(moistureColIdx, p.NewMoistures[i])
		}
		if profileColIdx >= 0 && i < len(p.NewProfiles) {
			writeCol(profileColIdx, p.NewProfiles[i])
		}
		if structureColIdx >= 0 && i < len(p.NewStructures) {
			writeCol(structureColIdx, p.NewStructures[i])
		}
		if lumberProfileColIdx >= 0 && i < len(p.NewLumberProfiles) {
			writeCol(lumberProfileColIdx, p.NewLumberProfiles[i])
		}
		if thicknessColIdx >= 0 && i < len(p.NewThicknesses) {
			writeCol(thicknessColIdx, p.NewThicknesses[i])
		}
		if widthColIdx >= 0 && i < len(p.NewWidths) {
			writeCol(widthColIdx, p.NewWidths[i])
		}
		if lengthColIdx >= 0 && i < len(p.NewLengths) {
			writeCol(lengthColIdx, p.NewLengths[i])
		}
		if heightColIdx >= 0 && i < len(p.NewHeights) {
			writeCol(heightColIdx, p.NewHeights[i])
		}
		if widthDColIdx >= 0 && i < len(p.NewWidthDs) {
			writeCol(widthDColIdx, p.NewWidthDs[i])
		}
		if lengthDColIdx >= 0 && i < len(p.NewLengthDs) {
			writeCol(lengthDColIdx, p.NewLengthDs[i])
		}
		if p.GOSTColIdx >= 0 && i < len(p.NewGOSTValues) {
			writeCol(p.GOSTColIdx, p.NewGOSTValues[i])
		}
		if p.TargetActionColIdx >= 0 && i < len(p.NewTargetActionManual) {
			writeCol(p.TargetActionColIdx, p.NewTargetActionManual[i])
		}
		if p.TargetActionManualColIdx >= 0 && i < len(p.NewTargetActionManualSettings) {
			writeCol(p.TargetActionManualColIdx, p.NewTargetActionManualSettings[i])
		}
		if p.DiameterColIdx >= 0 && i < len(p.NewDiameters) {
			writeCol(p.DiameterColIdx, p.NewDiameters[i])
		}
		if p.PriceValueColIdx >= 0 && i < len(p.NewPriceValues) {
			writeCol(p.PriceValueColIdx, p.NewPriceValues[i])
		}
		wrote++
	}

	fmt.Printf("[DEBUG] SaveExcelWithNewRows: wrote=%d starting_at_row=%d\n", wrote, startRow)

	if err := f.SaveAs(p.OutputPath); err != nil {
		return fmt.Errorf("ошибка сохранения файла: %w", err)
	}

	info, err := os.Stat(p.OutputPath)
	if err != nil {
		return fmt.Errorf("файл не создан: %w", err)
	}
	fmt.Printf("[DEBUG] SaveExcelWithNewRows: saved size=%d bytes\n", info.Size())

	return nil
}

// FindColumnIndex находит индекс колонки по имени (частичное совпадение, регистронезависимо)
func FindColumnIndex(headers []string, name string) int {
	for i, h := range headers {
		if strings.Contains(strings.ToLower(h), strings.ToLower(name)) {
			return i
		}
	}
	return -1
}

// GetMimeType возвращает MIME тип файла
func GetMimeType(filename string) string {
	mime := mime.TypeByExtension(filepath.Ext(filename))
	if mime == "" {
		return "application/octet-stream"
	}
	return mime
}

func buildRegionMap(addresses []string) (map[string][]string, []string) {
	regionMap := make(map[string][]string)
	for _, addr := range addresses {
		region := extractRegion(addr)
		regionMap[region] = append(regionMap[region], addr)
	}
	regionKeys := make([]string, 0, len(regionMap))
	for k := range regionMap {
		regionKeys = append(regionKeys, k)
	}
	return regionMap, regionKeys
}

func findAddressColumn(headerRow []string) int {
	for i, h := range headerRow {
		if strings.Contains(strings.ToLower(h), "адрес") {
			return i
		}
	}
	return -1
}

func pickAndWriteNewAddress(f *excelize.File, sheetName string, rows [][]string, rowIdx, addressColIdx int, regionMap map[string][]string, regionKeys []string, rnd *rand.Rand) {
	if addressColIdx >= len(rows[rowIdx]) {
		return
	}
	currentAddr := strings.TrimSpace(rows[rowIdx][addressColIdx])
	currentRegion := extractRegion(currentAddr)

	var candidates []string
	for _, key := range regionKeys {
		if key != "" && key != currentRegion {
			candidates = append(candidates, regionMap[key]...)
		}
	}
	if len(candidates) == 0 {
		return
	}

	newAddr := candidates[rnd.Intn(len(candidates))]
	cellName, err := excelize.CoordinatesToCellName(addressColIdx+1, rowIdx+1)
	if err != nil {
		return
	}
	_ = f.SetCellValue(sheetName, cellName, newAddr)
}

func shuffleSheetAddresses(f *excelize.File, sheetName string, rows [][]string, addressColIdx int, regionMap map[string][]string, regionKeys []string, rnd *rand.Rand) error {
	for rowIdx := 4; rowIdx < len(rows); rowIdx++ {
		pickAndWriteNewAddress(f, sheetName, rows, rowIdx, addressColIdx, regionMap, regionKeys, rnd)
	}
	return nil
}

// ShuffleAddresses перемешивает адреса в столбце "Адрес" для всех строк данных (начиная с 5-й)
func ShuffleAddresses(templatePath, outputPath string, addresses []string) error {
	if len(addresses) == 0 {
		return fmt.Errorf("список адресов пуст")
	}

	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("в файле нет листов")
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	regionMap, regionKeys := buildRegionMap(addresses)

	for _, sheetName := range sheets {
		if strings.EqualFold(sheetName, "Инструкция") {
			continue
		}
		rows, err := f.GetRows(sheetName)
		if err != nil || len(rows) == 0 {
			continue
		}

		headerRowIdx := findHeaderRow(rows)
		addressColIdx := findAddressColumn(rows[headerRowIdx])
		if addressColIdx < 0 {
			continue
		}

		if err := shuffleSheetAddresses(f, sheetName, rows, addressColIdx, regionMap, regionKeys, rnd); err != nil {
			return err
		}
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("ошибка сохранения файла: %w", err)
	}

	return nil
}

func extractRegion(addr string) string {
	lower := strings.ToLower(addr)
	if strings.Contains(lower, "москва") {
		return "москва"
	}
	parts := strings.Split(addr, ",")
	for _, part := range parts {
		p := strings.ToLower(strings.TrimSpace(part))
		if strings.Contains(p, "область") || strings.Contains(p, "край") || strings.Contains(p, "округ") {
			return strings.TrimSpace(part)
		}
	}
	return ""
}
