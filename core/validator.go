package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func ValidatePhone(phone string) bool {
	re := regexp.MustCompile(`^(?:\+7|8)[\s\-\(\)\d]{9,14}$`)
	return re.MatchString(phone)
}

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

func isUpper(s string) bool {
	for _, r := range s {
		return !unicode.IsLower(r)
	}
	return true
}

func ValidatePrice(price string) bool {
	_, err := strconv.Atoi(price)
	return err == nil
}

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
	}
	return result
}
