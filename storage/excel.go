package storage

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// LoadTemplate загружает XLSX файл и проверяет базовые правила
func LoadTemplate(path string) (*excelize.File, error) {
	f, err := excelize.OpenFile(path, excelize.Options{
		Password: "",
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}

	fmt.Printf("[DEBUG] Excel file opened successfully: %s\n", path)

	return f, nil
}

// SaveTemplate сохраняет XLSX файл и проверяет лимиты
func SaveTemplate(f *excelize.File, path string) error {
	sheets := f.GetSheetList()
	for _, sheet := range sheets {
		if sheet == "Инструкция" {
			continue
		}
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		if len(rows) > 50005 { // 4 защищенные строки + 50000 данных
			return fmt.Errorf("превышен лимит в 50 000 объявлений на листе '%s'", sheet)
		}
	}
	return f.SaveAs(path)
}

// isHeaderRow проверяет, похожа ли строка на заголовок таблицы
func isHeaderRow(row []string) bool {
	if len(row) < 3 {
		return false
	}

	headerKeywords := []string{
		"название", "описание", "цена", "фото", "адрес", "категория",
		"состояние", "вид товара", "номер", "контакт", "телефон",
		"продажа", "состояние", "длительность", "площадь", "ширина",
		"длина", "высота", "профиль", "кромка", "сорт", "wood",
		"title", "description", "price", "photo", "address", "category",
	}

	rowLower := make([]string, len(row))
	for i, cell := range row {
		rowLower[i] = strings.ToLower(strings.TrimSpace(cell))
	}

	matchCount := 0
	for _, cell := range rowLower {
		for _, keyword := range headerKeywords {
			if strings.Contains(cell, keyword) {
				matchCount++
				break
			}
		}
	}

	return matchCount >= 3
}

// GetSheetData возвращает данные из листа
func GetSheetData(f *excelize.File, sheetName string) ([]string, [][]string, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("[DEBUG] GetSheetData: sheet '%s' has %d rows\n", sheetName, len(rows))

	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("лист пустой")
	}

	headerRowIdx := 0
	for i := 0; i < len(rows) && i < 10; i++ {
		if isHeaderRow(rows[i]) {
			headerRowIdx = i
			break
		}
	}

	headers := rows[headerRowIdx]
	fmt.Printf("[DEBUG] Using header row %d: %v\n", headerRowIdx, headers)

	data := make([][]string, 0, len(rows)-headerRowIdx-1)
	for i := headerRowIdx + 1; i < len(rows); i++ {
		row := make([]string, len(headers))
		for j := 0; j < len(headers) && j < len(rows[i]); j++ {
			row[j] = rows[i][j]
		}
		hasData := false
		for j := 0; j < len(row); j++ {
			if row[j] != "" {
				hasData = true
				break
			}
		}
		if !hasData {
			fmt.Printf("[DEBUG] Skipping empty row at index %d\n", i)
			continue
		}

		firstCell := strings.ToLower(strings.TrimSpace(row[0]))
		if strings.Contains(firstCell, "подробнее о параметре") ||
			strings.Contains(firstCell, "обязательный") ||
			strings.Contains(firstCell, "необязательный") ||
			strings.Contains(firstCell, "способ размещения") {
			fmt.Printf("[DEBUG] Skipping template row at index %d: %v\n", i, row[0])
			continue
		}

		data = append(data, row)
	}

	fmt.Printf("[DEBUG] Total data rows (excluding empty): %d\n", len(data))

	return headers, data, nil
}
