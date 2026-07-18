package storage

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func LoadTemplate(path string) (*excelize.File, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	if _, err := f.GetSheetIndex("Инструкция"); err != nil {
		return nil, fmt.Errorf("лист 'Инструкция' отсутствует или переименован")
	}
	return f, nil
}

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
		if len(rows) > 50005 {
			return fmt.Errorf("превышен лимит в 50 000 объявлений на листе '%s'", sheet)
		}
	}
	return f.SaveAs(path)
}
