package storage

import (
	"encoding/json"
	"os"

	"AVITOUO/core"
)

const settingsFile = "settings.json"

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
	err = json.Unmarshal(data, &s)
	return &s, err
}

func SaveSettings(s *core.Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFile, data, 0644)
}
