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

	"AVITOUO/core"

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
// Для каждого набора берется ВСЕ фото из папки, и КАЖДОЕ фото уникализируется.
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

	fmt.Printf("[DEBUG] Found %d source images in folder\n", len(sourceImages))

	id := core.GenerateUniqueID()[:8]
	var result []string

	for adIdx := 0; adIdx < count; adIdx++ {
		var names []string
		for photoIdx, srcPath := range sourceImages {
			ext := strings.ToLower(filepath.Ext(srcPath))
			baseName := fmt.Sprintf("%s_ad%d_photo%d%s", id, adIdx+1, photoIdx+1, ext)
			savePath := filepath.Join(fullDir, baseName)

			srcData, err := os.ReadFile(srcPath)
			if err != nil {
				return nil, fmt.Errorf("ошибка чтения фото: %w", err)
			}

			img, err := imaging.Decode(strings.NewReader(string(srcData)))
			if err != nil {
				return nil, fmt.Errorf("ошибка декодирования изображения: %w", err)
			}

			uniqueImg := applyUniqueTransformations(img, adIdx*len(sourceImages)+photoIdx)
			if err := imaging.Save(uniqueImg, savePath); err != nil {
				return nil, fmt.Errorf("ошибка сохранения уникального фото: %w", err)
			}
			names = append(names, baseName)
			fmt.Printf("[DEBUG] Ad %d: unique photo %d/%d: %s\n", adIdx+1, photoIdx+1, len(sourceImages), baseName)
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
func SaveExcelWithNewRows(templatePath, outputPath string, sheetName string, titleColIdx, descColIdx, imageNamesIdx, contactColIdx, phoneColIdx, addressColIdx, companyColIdx, emailColIdx int, newTitles, newDescriptions, newImageNames []string, newContacts, newPhones, newAddresses, newCompanies, newEmails []string, idColIdx, placementColIdx, contactMethodColIdx, categoryColIdx, productTypeColIdx, subProductTypeColIdx, priceUnitColIdx, conditionColIdx, availabilityColIdx, adTypeColIdx, salesTypeColIdx, connectColIdx, processingColIdx, purposeColIdx int, newIDs, newPlacements, newContactMethods, newCategories, newProductTypes, newSubProductTypes, newPriceUnits, newConditions, newAvailabilities, newAdTypes, newSalesTypes, newConnects, newProcessing, newPurpose []string) error {
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия шаблона: %w", err)
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("ошибка чтения листа: %w", err)
	}

	fmt.Printf("[DEBUG] SaveExcelWithNewRows: sheet=%q rows_in=%d newTitles=%d titleIdx=%d descIdx=%d imageIdx=%d contactIdx=%d phoneIdx=%d addressIdx=%d companyIdx=%d emailIdx=%d idIdx=%d placementIdx=%d methodIdx=%d categoryIdx=%d productIdx=%d subProductIdx=%d priceUnitIdx=%d conditionIdx=%d availabilityIdx=%d adTypeIdx=%d salesTypeIdx=%d connectIdx=%d processingIdx=%d purposeIdx=%d output=%q\n", sheetName, len(rows), len(newTitles), titleColIdx, descColIdx, imageNamesIdx, contactColIdx, phoneColIdx, addressColIdx, companyColIdx, emailColIdx, idColIdx, placementColIdx, contactMethodColIdx, categoryColIdx, productTypeColIdx, subProductTypeColIdx, priceUnitColIdx, conditionColIdx, availabilityColIdx, adTypeColIdx, salesTypeColIdx, connectColIdx, processingColIdx, purposeColIdx, outputPath)

	if len(rows) == 0 {
		return fmt.Errorf("лист пустой")
	}

	if titleColIdx < 0 || descColIdx < 0 || imageNamesIdx < 0 || contactColIdx < 0 || phoneColIdx < 0 || addressColIdx < 0 || companyColIdx < 0 || emailColIdx < 0 || idColIdx < 0 || placementColIdx < 0 || contactMethodColIdx < 0 || categoryColIdx < 0 || productTypeColIdx < 0 || subProductTypeColIdx < 0 {
		if len(rows) > 0 {
			firstRow := rows[0]
			fmt.Printf("[DEBUG] Fallback scan on first row: %v\n", firstRow)
			if titleColIdx < 0 {
				for i, h := range firstRow {
					if strings.Contains(strings.ToLower(h), "title") || strings.Contains(strings.ToLower(h), "название") || strings.Contains(strings.ToLower(h), "заголовок") {
						titleColIdx = i
						break
					}
				}
			}
			if descColIdx < 0 {
				for i, h := range firstRow {
					if strings.Contains(strings.ToLower(h), "description") || strings.Contains(strings.ToLower(h), "описание") || strings.Contains(strings.ToLower(h), "текст") {
						descColIdx = i
						break
					}
				}
			}
			if imageNamesIdx < 0 {
				for i, h := range firstRow {
					if strings.Contains(strings.ToLower(h), "image") || strings.Contains(strings.ToLower(h), "фото") || strings.Contains(strings.ToLower(h), "изображение") {
						imageNamesIdx = i
						break
					}
				}
			}
			if contactColIdx < 0 {
				contactColIdx = findColumnInFirstRow(firstRow, "контактное лицо", "контакт", "contact")
			}
			if phoneColIdx < 0 {
				phoneColIdx = findColumnInFirstRow(firstRow, "номер телефона", "телефон", "phone")
			}
			if addressColIdx < 0 {
				addressColIdx = findColumnInFirstRow(firstRow, "адрес", "address")
			}
			if companyColIdx < 0 {
				companyColIdx = findColumnInFirstRow(firstRow, "название компании", "компания", "организация", "company")
			}
			if emailColIdx < 0 {
				emailColIdx = findColumnInFirstRow(firstRow, "почта", "email", "e-mail", "электронная почта")
			}
			if idColIdx < 0 {
				idColIdx = findColumnInFirstRow(firstRow, "уникальный идентификатор", "id")
			}
			if placementColIdx < 0 {
				placementColIdx = findColumnInFirstRow(firstRow, "способ размещения", "размещения")
			}
			if contactMethodColIdx < 0 {
				contactMethodColIdx = findColumnInFirstRow(firstRow, "способ связи", "связи")
			}
			if categoryColIdx < 0 {
				categoryColIdx = findColumnInFirstRow(firstRow, "категория")
			}
			if productTypeColIdx < 0 {
				productTypeColIdx = findColumnInFirstRow(firstRow, "вид товара", "товара")
			}
			if subProductTypeColIdx < 0 {
				subProductTypeColIdx = findColumnInFirstRow(firstRow, "подвид товара", "подвид")
			}
			if priceUnitColIdx < 0 {
				priceUnitColIdx = findColumnInFirstRow(firstRow, "цена за", "единица")
			}
			if conditionColIdx < 0 {
				conditionColIdx = findColumnInFirstRow(firstRow, "состояние")
			}
			if availabilityColIdx < 0 {
				availabilityColIdx = findColumnInFirstRow(firstRow, "доступность")
			}
			if adTypeColIdx < 0 {
				adTypeColIdx = findColumnInFirstRow(firstRow, "вид объявления")
			}
			if salesTypeColIdx < 0 {
				salesTypeColIdx = findColumnInFirstRow(firstRow, "вид продажи", "продажи")
			}
			if connectColIdx < 0 {
				connectColIdx = findColumnInFirstRow(firstRow, "соединять")
			}
			if processingColIdx < 0 {
				processingColIdx = findColumnInFirstRow(firstRow, "обработка")
			}
			if purposeColIdx < 0 {
				purposeColIdx = findColumnInFirstRow(firstRow, "назначение")
			}
			fmt.Printf("[DEBUG] Fallback result: titleIdx=%d descIdx=%d imageIdx=%d contactIdx=%d phoneIdx=%d addressIdx=%d companyIdx=%d emailIdx=%d idIdx=%d placementIdx=%d methodIdx=%d categoryIdx=%d productIdx=%d subProductIdx=%d priceUnitIdx=%d conditionIdx=%d availabilityIdx=%d adTypeIdx=%d salesTypeIdx=%d connectIdx=%d processingIdx=%d purposeIdx=%d\n", titleColIdx, descColIdx, imageNamesIdx, contactColIdx, phoneColIdx, addressColIdx, companyColIdx, emailColIdx, idColIdx, placementColIdx, contactMethodColIdx, categoryColIdx, productTypeColIdx, subProductTypeColIdx, priceUnitColIdx, conditionColIdx, availabilityColIdx, adTypeColIdx, salesTypeColIdx, connectColIdx, processingColIdx, purposeColIdx)
		}
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
