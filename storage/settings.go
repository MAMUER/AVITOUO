package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"AVITOUO/core"
)

const settingsFile = "settings.json"

// LoadSettings загружает настройки из JSON файла
// Если файл не существует - возвращает настройки по умолчанию
func LoadSettings() (*core.Settings, error) {
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return &core.Settings{
			Contacts:  []string{"Мариелена"},
			Phones:    []string{"79268509135"},
			Addresses: []string{"Мытищи, Волковское ш., 21А"},
			Companies: []string{"СтройДерево"},
			Emails:    []string{"stroyderevo-direct@yandex.ru"},
		}, nil
	}

	var s core.Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("ошибка парсинга настроек: %w", err)
	}

	return &s, nil
}

// SaveSettings сохраняет настройки в JSON файл
func SaveSettings(s *core.Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFile, data, 0644)
}
