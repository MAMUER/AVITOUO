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

// GenerateUniquePhotos создает N уникальных копий фото с изменением EXIF, пикселей и размером
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

	fmt.Printf("[DEBUG] Found %d source images for duplication\n", len(sourceImages))

	var generatedNames []string
	id := core.GenerateUniqueID()[:8]

	for i := 0; i < count; i++ {
		srcPath := sourceImages[i%len(sourceImages)]
		ext := strings.ToLower(filepath.Ext(srcPath))

		photoName := fmt.Sprintf("%s_%d%s", id, i+1, ext)
		savePath := filepath.Join(fullDir, photoName)

		srcData, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения исходного фото: %w", err)
		}

		img, err := imaging.Decode(strings.NewReader(string(srcData)))
		if err != nil {
			return nil, fmt.Errorf("ошибка декодирования изображения: %w", err)
		}

		uniqueImg := applyUniqueTransformations(img, i)

		if err := imaging.Save(uniqueImg, savePath); err != nil {
			return nil, fmt.Errorf("ошибка сохранения уникального фото: %w", err)
		}

		generatedNames = append(generatedNames, photoName)
		fmt.Printf("[DEBUG] Generated unique photo: %s\n", photoName)
	}

	return generatedNames, nil
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
			if index%3 == 0 {
				r = clampColor(r + uint32(index%10))
			} else if index%3 == 1 {
				g = clampColor(g + uint32(index%10))
			} else {
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

		w, createErr := zipWriter.Create(entry.Name())
		if createErr != nil {
			fileToZip.Close()
			continue
		}

		if _, copyErr := io.Copy(w, fileToZip); copyErr != nil {
			fileToZip.Close()
			continue
		}
		fileToZip.Close()
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
func SaveExcelWithNewRows(templatePath, outputPath string, sheetName string, titleColIdx, descColIdx, imageNamesIdx int, newTitles, newDescriptions, newImageNames []string) error {
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

	startRow := len(rows) + 1

	for i := 0; i < len(newTitles); i++ {
		rowNum := startRow + i
		if titleColIdx >= 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'A'+titleColIdx, rowNum), newTitles[i])
		}
		if descColIdx >= 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'A'+descColIdx, rowNum), newDescriptions[i])
		}
		if imageNamesIdx >= 0 && i < len(newImageNames) {
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'A'+imageNamesIdx, rowNum), newImageNames[i])
		}
	}

	return f.SaveAs(outputPath)
}

// FindColumnIndex находит индекс колонки по имени (регистронезависимо)
func FindColumnIndex(headers []string, name string) int {
	for i, h := range headers {
		if strings.EqualFold(h, name) {
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
