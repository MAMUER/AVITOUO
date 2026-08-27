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
			Contacts:    []string{"Мариелена"},
			Phones:      []string{"79268509135"},
			Addresses:   []string{"Мытищи, Волковское ш., 21А"},
			Companies:   []string{"СтройДерево"},
			Emails:      []string{"stroyderevo-direct@yandex.ru"},
			ProductType: "Доска",
		}, nil
	}

	var s core.Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("ошибка парсинга настроек: %w", err)
	}

	if s.Contacts == nil {
		s.Contacts = []string{"Мариелена"}
	}
	if s.Phones == nil {
		s.Phones = []string{"79268509135"}
	}
	if s.Addresses == nil {
		s.Addresses = []string{"Мытищи, Волковское ш., 21А"}
	}
	if s.Companies == nil {
		s.Companies = []string{"СтройДерево"}
	}
	if s.Emails == nil {
		s.Emails = []string{"stroyderevo-direct@yandex.ru"}
	}
	if s.Placement == "" {
		s.Placement = "Package"
	}
	if s.ContactMethod == "" {
		s.ContactMethod = "По телефону и в сообщениям"
	}
	if s.AdType == "" {
		s.AdType = "Товар от производителя"
	}
	if s.Condition == "" {
		s.Condition = "Новое"
	}
	if s.Availability == "" {
		s.Availability = "В наличии"
	}
	if s.PriceUnit == "" {
		s.PriceUnit = "Штуку"
	}
	if s.ProductType == "" {
		s.ProductType = "Доска"
	}
	if s.Connect == "" {
		s.Connect = "Да"
	}
	if s.Proxy == "" {
		s.Proxy = ""
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
