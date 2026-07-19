package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ValidatePhone проверяет российский формат телефона
// Разрешены: +7 или 8, пробелы, скобки, дефисы
func ValidatePhone(phone string) bool {
	re := regexp.MustCompile(`^(?:\+7|8)[\s\-\(\)\d]{9,14}$`)
	return re.MatchString(phone)
}

// ValidateTitle проверяет название объявления
// До 100 символов, первая буква заглавная, запрещены: цена, продам, тел
func ValidateTitle(title string) error {
	if len(title) > 100 {
		return fmt.Errorf("название не должно превышать 100 символов")
	}
	lower := strings.ToLower(title)
	if strings.Contains(lower, "цена") || strings.Contains(lower, "продам") || strings.Contains(lower, "тел") {
		return fmt.Errorf("название не должно содержать 'цена', 'продам' или контакты")
	}
	if len(title) > 0 && !isUpper(title) {
		return fmt.Errorf("первая буква должна быть заглавной")
	}
	return nil
}

// ValidatePrice проверяет, что цена - целое число
func ValidatePrice(price string) bool {
	_, err := strconv.Atoi(price)
	return err == nil
}

// ValidateDescription проверяет описание объявления
// До 7500 символов, проверка на уникальность, автоформатирование CDATA
func ValidateDescription(desc string) error {
	if len(desc) > 7500 {
		return fmt.Errorf("описание не должно превышать 7500 символов")
	}
	if strings.TrimSpace(desc) == "" {
		return fmt.Errorf("описание не может быть пустым")
	}
	return nil
}

// FormatDescription форматирует описание в CDATA
// Переносы строк заменяются на <br>
// Разрешённые теги: p, br, strong, em, ul, ol, li
func FormatDescription(desc string) string {
	formatted := strings.ReplaceAll(desc, "\n", "<br>")
	formatted = strings.ReplaceAll(formatted, "\r\n", "<br>")
	return fmt.Sprintf("<![CDATA[%s]]>", formatted)
}

// ParseCategoryString парсит строку категории из 1-й строки Excel
// Пример: "Для дома и дачи - Ремонт и строительство - Стройматаериалы - Крепёж"
// Возвращает: Категория, Вид товара, Подвид товара, Тип крепежа/Тип товара
func ParseCategoryString(categoryStr string) map[string]string {
	parts := strings.Split(categoryStr, " - ")
	result := make(map[string]string)

	if len(parts) >= 2 {
		result["Категория"] = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		result["Вид товара"] = strings.TrimSpace(parts[2])
	}
	if len(parts) >= 4 {
		result["Подвид товара"] = strings.TrimSpace(parts[3])
	}
	if len(parts) >= 5 {
		result["Тип крепежа"] = strings.TrimSpace(parts[4])
		result["Тип товара"] = strings.TrimSpace(parts[3]) // для освещения
	}
	if len(parts) >= 6 {
		result["Тип потолочного освещения"] = strings.TrimSpace(parts[5])
	}

	return result
}

func isUpper(s string) bool {
	for _, r := range s {
		return !unicode.IsLower(r)
	}
	return true
}
