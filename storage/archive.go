package storage

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/xuri/excelize/v2"
)

const PhotosDir = "photos"

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

			srcData, err := os.ReadFile(srcPath)
			if err != nil {
				return nil, fmt.Errorf("ошибка чтения фото: %w", err)
			}

			img, err := imaging.Decode(strings.NewReader(string(srcData)))
			if err != nil {
				return nil, fmt.Errorf("ошибка декодирования изображения: %w", err)
			}

			uniqueImg := applyUniqueTransformations(img, globalPhotoIdx)
			if err := imaging.Save(uniqueImg, savePath); err != nil {
				return nil, fmt.Errorf("ошибка сохранения уникального фото: %w", err)
			}
			names = append(names, baseName)
			globalPhotoIdx++
		}
		result = append(result, strings.Join(names, " | "))
	}

	return result, nil
}

// applyUniqueTransformations применяет уникальные трансформации к изображению
func applyUniqueTransformations(img image.Image, index int) image.Image {
	bounds := img.Bounds()

	shiftX := (index * 7) % bounds.Dx()
	shiftY := (index * 11) % bounds.Dy()

	newW := bounds.Dx() + 1
	newH := bounds.Dy() + 1
	dst := imaging.New(newW, newH, getColorForIndex(index))

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			pxX := (x + shiftX) % bounds.Dx()
			pxY := (y + shiftY) % bounds.Dy()
			c := colorAt(img, pxX, pxY)
			r, g, b, a := c.RGBA()
			switch index % 3 {
			case 0:
				r = clampColor(r + uint32(index%10))
			case 1:
				g = clampColor(g + uint32(index%10))
			default:
				b = clampColor(b + uint32(index%10))
			}
			dst.Set(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)})
		}
	}

	return dst
}

func getColorForIndex(index int) color.RGBA {
	colors := []color.RGBA{
		{255, 255, 255, 0},
		{250, 250, 250, 0},
		{245, 245, 245, 0},
	}
	return colors[index%len(colors)]
}

func clampColor(v uint32) uint32 {
	if v > 65535 {
		return 65535
	}
	return v
}

func colorAt(img image.Image, x, y int) color.RGBA {
	c := img.At(x, y)
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
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

		srcPath := filepath.Join(fullDir, entry.Name())
		fileToZip, openErr := os.Open(srcPath)
		if openErr != nil {
			continue
		}
		defer func() { _ = fileToZip.Close() }()

		w, createErr := zipWriter.Create(entry.Name())
		if createErr != nil {
			continue
		}

		if _, copyErr := io.Copy(w, fileToZip); copyErr != nil {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}

	return strings.Join(fileNames, "|"), nil
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

// SaveExcelWithNewRows добавляет новые строки в Excel файл и сохраняет
func SaveExcelWithNewRows(templatePath, outputPath string, sheetName string, titleColIdx, descColIdx, imageNamesIdx, contactColIdx, phoneColIdx, addressColIdx, companyColIdx, emailColIdx int, newTitles, newDescriptions, newImageNames []string, newContacts, newPhones, newAddresses, newCompanies, newEmails []string, idColIdx, placementColIdx, contactMethodColIdx, categoryColIdx, productTypeColIdx, subProductTypeColIdx, priceUnitColIdx, conditionColIdx, availabilityColIdx, adTypeColIdx, salesTypeColIdx, connectColIdx, processingColIdx, purposeColIdx, gostColIdx int, newIDs, newPlacements, newContactMethods, newCategories, newProductTypes, newSubProductTypes, newPriceUnits, newConditions, newAvailabilities, newAdTypes, newSalesTypes, newConnects, newProcessing, newPurpose, newLumberTypes, newWoodTypes, newEdges, newGrades, newMoistures, newProfiles, newStructures, newLumberProfiles, newThicknesses, newWidths, newLengths, newHeights, newWidthDs, newLengthDs, newGOSTValues []string, targetActionColIdx, targetActionManualColIdx int, newTargetActionManual, newTargetActionManualSettings []string, diameterColIdx int, newDiameters []string, priceValueColIdx int, newPriceValues []string) error {
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия шаблона: %w", err)
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("ошибка чтения листа: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("лист пустой")
	}

	headerRowIdx := -1
	for i := 0; i < len(rows) && i < 20; i++ {
		if isHeaderRow(rows[i]) {
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

	markUsed(titleColIdx)
	markUsed(descColIdx)
	markUsed(imageNamesIdx)
	markUsed(contactColIdx)
	markUsed(phoneColIdx)
	markUsed(addressColIdx)
	markUsed(companyColIdx)
	markUsed(emailColIdx)
	markUsed(idColIdx)
	markUsed(placementColIdx)
	markUsed(contactMethodColIdx)
	markUsed(categoryColIdx)
	markUsed(productTypeColIdx)
	markUsed(subProductTypeColIdx)
	markUsed(priceUnitColIdx)
	markUsed(priceValueColIdx)
	markUsed(conditionColIdx)
	markUsed(availabilityColIdx)
	markUsed(adTypeColIdx)
	markUsed(salesTypeColIdx)
	markUsed(connectColIdx)
	markUsed(processingColIdx)
	markUsed(purposeColIdx)
	markUsed(gostColIdx)
	markUsed(targetActionColIdx)
	markUsed(targetActionManualColIdx)
	markUsed(diameterColIdx)

	if titleColIdx < 0 || descColIdx < 0 || imageNamesIdx < 0 || contactColIdx < 0 || phoneColIdx < 0 || addressColIdx < 0 || companyColIdx < 0 || emailColIdx < 0 || idColIdx < 0 || placementColIdx < 0 || contactMethodColIdx < 0 || categoryColIdx < 0 || productTypeColIdx < 0 || subProductTypeColIdx < 0 {
		fmt.Printf("[DEBUG] Fallback scan on header row %d: %v\n", headerRowIdx, headerRow)
		if titleColIdx < 0 {
			titleColIdx = findUnique("title", "название", "заголовок")
		}
		if descColIdx < 0 {
			descColIdx = findUnique("description", "описание", "текст")
		}
		if imageNamesIdx < 0 {
			imageNamesIdx = findUnique("image", "фото", "изображение")
		}
		if contactColIdx < 0 {
			contactColIdx = findUnique("контактное лицо", "контакт", "contact")
		}
		if phoneColIdx < 0 {
			phoneColIdx = findUnique("номер телефона", "телефон", "phone")
		}
		if addressColIdx < 0 {
			addressColIdx = findUnique("адрес", "address")
		}
		if companyColIdx < 0 {
			companyColIdx = findUnique("название компании", "компания", "организация", "company")
		}
		if emailColIdx < 0 {
			emailColIdx = findUnique("почта", "email", "e-mail", "электронная почта")
		}
		if idColIdx < 0 {
			idColIdx = findUnique("уникальный идентификатор", "id")
		}
		if placementColIdx < 0 {
			placementColIdx = findUnique("способ размещения", "размещения")
		}
		if contactMethodColIdx < 0 {
			contactMethodColIdx = findUnique("способ связи", "связи")
		}
		if categoryColIdx < 0 {
			categoryColIdx = findUnique("категория")
		}
		if productTypeColIdx < 0 {
			productTypeColIdx = findUnique("вид товара", "товара")
		}
		if subProductTypeColIdx < 0 {
			subProductTypeColIdx = findUnique("подвид товара", "подвид")
		}
		if priceUnitColIdx < 0 {
			priceUnitColIdx = findUnique("цена за", "единица")
		}
		if priceValueColIdx < 0 {
			priceValueColIdx = findUnique("цена")
		}
		if conditionColIdx < 0 {
			conditionColIdx = findUnique("состояние")
		}
		if availabilityColIdx < 0 {
			availabilityColIdx = findUnique("доступность")
		}
		if adTypeColIdx < 0 {
			adTypeColIdx = findUnique("вид объявления")
		}
		if salesTypeColIdx < 0 {
			salesTypeColIdx = findUnique("вид продажи", "продажи")
		}
		if connectColIdx < 0 {
			connectColIdx = findUnique("соединять")
		}
		if processingColIdx < 0 {
			processingColIdx = findUnique("обработка")
		}
		if purposeColIdx < 0 {
			purposeColIdx = findUnique("назначение")
		}
		fmt.Printf("[DEBUG] Fallback result: titleIdx=%d descIdx=%d imageIdx=%d contactIdx=%d phoneIdx=%d addressIdx=%d companyIdx=%d emailIdx=%d idIdx=%d placementIdx=%d methodIdx=%d categoryIdx=%d productIdx=%d subProductIdx=%d priceUnitIdx=%d priceValueIdx=%d conditionIdx=%d availabilityIdx=%d adTypeIdx=%d salesTypeIdx=%d connectIdx=%d processingIdx=%d purposeIdx=%d\n", titleColIdx, descColIdx, imageNamesIdx, contactColIdx, phoneColIdx, addressColIdx, companyColIdx, emailColIdx, idColIdx, placementColIdx, contactMethodColIdx, categoryColIdx, productTypeColIdx, subProductTypeColIdx, priceUnitColIdx, priceValueColIdx, conditionColIdx, availabilityColIdx, adTypeColIdx, salesTypeColIdx, connectColIdx, processingColIdx, purposeColIdx)
	}

	if titleColIdx < 0 {
		titleColIdx = 0
	}
	if descColIdx < 0 {
		descColIdx = 1
	}
	if imageNamesIdx < 0 {
		if len(rows) > 0 && len(rows[0]) > 2 {
			imageNamesIdx = 2
		} else {
			imageNamesIdx = -1
		}
	}
	if idColIdx < 0 {
		idColIdx = -1
	}
	if placementColIdx < 0 {
		placementColIdx = -1
	}
	if contactMethodColIdx < 0 {
		contactMethodColIdx = -1
	}
	if categoryColIdx < 0 {
		categoryColIdx = -1
	}
	if productTypeColIdx < 0 {
		productTypeColIdx = -1
	}
	if subProductTypeColIdx < 0 {
		subProductTypeColIdx = -1
	}
	if priceUnitColIdx < 0 {
		priceUnitColIdx = -1
	}
	if priceValueColIdx < 0 {
		priceValueColIdx = -1
	}
	if conditionColIdx < 0 {
		conditionColIdx = -1
	}
	if availabilityColIdx < 0 {
		availabilityColIdx = -1
	}
	if adTypeColIdx < 0 {
		adTypeColIdx = -1
	}
	if salesTypeColIdx < 0 {
		salesTypeColIdx = -1
	}
	if connectColIdx < 0 {
		connectColIdx = -1
	}
	if processingColIdx < 0 {
		processingColIdx = -1
	}
	if purposeColIdx < 0 {
		purposeColIdx = -1
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
		if targetActionColIdx < 0 {
			targetActionColIdx = findUnique("настройка цены целевого действия")
		}
		if targetActionManualColIdx < 0 {
			targetActionManualColIdx = findUnique("настройка цены целевого действия: ручная")
		}
		if diameterColIdx < 0 {
			diameterColIdx = findUnique("диаметр")
		}
		if priceValueColIdx < 0 {
			priceValueColIdx = findUnique("цена")
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
		if err := f.SetCellValue(sheetName, cell, value); err != nil {
			fmt.Printf("[DEBUG] SetCellValue error at %s: %v\n", cell, err)
		}
	}

	for i := 0; i < len(newTitles); i++ {
		writeCol(titleColIdx, newTitles[i])
		writeCol(descColIdx, newDescriptions[i])
		if imageNamesIdx >= 0 && i < len(newImageNames) {
			writeCol(imageNamesIdx, newImageNames[i])
		}
		if contactColIdx >= 0 && i < len(newContacts) {
			writeCol(contactColIdx, newContacts[i])
		}
		if phoneColIdx >= 0 && i < len(newPhones) {
			writeCol(phoneColIdx, newPhones[i])
		}
		if addressColIdx >= 0 && i < len(newAddresses) {
			writeCol(addressColIdx, newAddresses[i])
		}
		if companyColIdx >= 0 && i < len(newCompanies) {
			writeCol(companyColIdx, newCompanies[i])
		}
		if emailColIdx >= 0 && i < len(newEmails) {
			writeCol(emailColIdx, newEmails[i])
		}
		if idColIdx >= 0 && i < len(newIDs) {
			writeCol(idColIdx, newIDs[i])
		}
		if placementColIdx >= 0 && i < len(newPlacements) {
			writeCol(placementColIdx, newPlacements[i])
		}
		if contactMethodColIdx >= 0 && i < len(newContactMethods) {
			writeCol(contactMethodColIdx, newContactMethods[i])
		}
		if categoryColIdx >= 0 && i < len(newCategories) {
			writeCol(categoryColIdx, newCategories[i])
		}
		if productTypeColIdx >= 0 && i < len(newProductTypes) {
			writeCol(productTypeColIdx, newProductTypes[i])
		}
		if subProductTypeColIdx >= 0 && i < len(newSubProductTypes) {
			writeCol(subProductTypeColIdx, newSubProductTypes[i])
		}
		if priceUnitColIdx >= 0 && i < len(newPriceUnits) {
			writeCol(priceUnitColIdx, newPriceUnits[i])
		}
		if conditionColIdx >= 0 && i < len(newConditions) {
			writeCol(conditionColIdx, newConditions[i])
		}
		if availabilityColIdx >= 0 && i < len(newAvailabilities) {
			writeCol(availabilityColIdx, newAvailabilities[i])
		}
		if adTypeColIdx >= 0 && i < len(newAdTypes) {
			writeCol(adTypeColIdx, newAdTypes[i])
		}
		if salesTypeColIdx >= 0 && i < len(newSalesTypes) {
			writeCol(salesTypeColIdx, newSalesTypes[i])
		}
		if connectColIdx >= 0 && i < len(newConnects) {
			writeCol(connectColIdx, newConnects[i])
		}
		if processingColIdx >= 0 && i < len(newProcessing) {
			writeCol(processingColIdx, newProcessing[i])
		}
		if purposeColIdx >= 0 && i < len(newPurpose) {
			writeCol(purposeColIdx, newPurpose[i])
		}
		if lumberTypeColIdx >= 0 && i < len(newLumberTypes) {
			writeCol(lumberTypeColIdx, newLumberTypes[i])
		}
		if woodTypeColIdx >= 0 && i < len(newWoodTypes) {
			writeCol(woodTypeColIdx, newWoodTypes[i])
		}
		if edgeColIdx >= 0 && i < len(newEdges) {
			writeCol(edgeColIdx, newEdges[i])
		}
		if gradeColIdx >= 0 && i < len(newGrades) {
			writeCol(gradeColIdx, newGrades[i])
		}
		if moistureColIdx >= 0 && i < len(newMoistures) {
			writeCol(moistureColIdx, newMoistures[i])
		}
		if profileColIdx >= 0 && i < len(newProfiles) {
			writeCol(profileColIdx, newProfiles[i])
		}
		if structureColIdx >= 0 && i < len(newStructures) {
			writeCol(structureColIdx, newStructures[i])
		}
		if lumberProfileColIdx >= 0 && i < len(newLumberProfiles) {
			writeCol(lumberProfileColIdx, newLumberProfiles[i])
		}
		if thicknessColIdx >= 0 && i < len(newThicknesses) {
			writeCol(thicknessColIdx, newThicknesses[i])
		}
		if widthColIdx >= 0 && i < len(newWidths) {
			writeCol(widthColIdx, newWidths[i])
		}
		if lengthColIdx >= 0 && i < len(newLengths) {
			writeCol(lengthColIdx, newLengths[i])
		}
		if heightColIdx >= 0 && i < len(newHeights) {
			writeCol(heightColIdx, newHeights[i])
		}
		if widthDColIdx >= 0 && i < len(newWidthDs) {
			writeCol(widthDColIdx, newWidthDs[i])
		}
		if lengthDColIdx >= 0 && i < len(newLengthDs) {
			writeCol(lengthDColIdx, newLengthDs[i])
		}
		if gostColIdx >= 0 && i < len(newGOSTValues) {
			writeCol(gostColIdx, newGOSTValues[i])
		}
		if targetActionColIdx >= 0 && i < len(newTargetActionManual) {
			writeCol(targetActionColIdx, newTargetActionManual[i])
		}
		if targetActionManualColIdx >= 0 && i < len(newTargetActionManualSettings) {
			writeCol(targetActionManualColIdx, newTargetActionManualSettings[i])
		}
		if diameterColIdx >= 0 && i < len(newDiameters) {
			writeCol(diameterColIdx, newDiameters[i])
		}
		if priceValueColIdx >= 0 && i < len(newPriceValues) {
			writeCol(priceValueColIdx, newPriceValues[i])
		}
		wrote++
	}

	fmt.Printf("[DEBUG] SaveExcelWithNewRows: wrote=%d starting_at_row=%d\n", wrote, startRow)

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("ошибка сохранения файла: %w", err)
	}

	info, err := os.Stat(outputPath)
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

// findColumnInFirstRow ищет колонку в первой строке по нескольким синонимам
func findColumnInFirstRow(row []string, names ...string) int {
	for i, h := range row {
		lower := strings.ToLower(h)
		for _, name := range names {
			if strings.Contains(lower, strings.ToLower(name)) {
				return i
			}
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
