package storage

import (
	"fmt"

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

	// Проверка наличия листа "Инструкция"
	if _, err := f.GetSheetIndex("Инструкция"); err != nil {
		return nil, fmt.Errorf("лист 'Инструкция' отсутствует или переименован")
	}

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

	// Находим реальную шапку (строку с первой буквы А)
	headers := rows[0]
	fmt.Printf("[DEBUG] Headers: %v\n", headers)

	data := make([][]string, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := make([]string, len(headers))
		for j := 0; j < len(headers) && j < len(rows[i]); j++ {
			row[j] = rows[i][j]
		}
		// Проверяем, что строка содержит данные (хотя бы один непустой столбец)
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
		data = append(data, row)
	}

	fmt.Printf("[DEBUG] Total data rows (excluding empty): %d\n", len(data))

	return headers, data, nil
}
