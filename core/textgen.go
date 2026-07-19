package core

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// TextGenerator генерирует уникальные вариации текста
type TextGenerator struct {
	used map[string]bool
	rnd  *rand.Rand
}

// NewTextGenerator создаёт новый генератор
func NewTextGenerator() *TextGenerator {
	return &TextGenerator{
		used: make(map[string]bool),
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateVariations генерирует N уникальных вариаций из шаблона
// Поддерживаемые конструкции: {вариант1|вариант2|вариант3}
func (tg *TextGenerator) GenerateVariations(template string, count int) ([]string, error) {
	if strings.TrimSpace(template) == "" {
		return nil, fmt.Errorf("шаблон пустой")
	}

	groups := extractGroups(template)
	totalPossible := calculateCombinations(groups)
	if totalPossible < count {
		return nil, fmt.Errorf("недостаточно уникальных комбинаций: возможно %d, запрошено %d", totalPossible, count)
	}

	results := make([]string, 0, count)
	attempts := 0
	maxAttempts := count * 10

	for len(results) < count && attempts < maxAttempts {
		attempts++
		variant := replaceGroups(template, groups, tg.rnd)
		if !tg.used[variant] {
			tg.used[variant] = true
			results = append(results, variant)
		}
	}

	if len(results) < count {
		return nil, fmt.Errorf("не удалось сгенерировать %d уникальных вариантов, получено %d", count, len(results))
	}

	return results, nil
}

// GenerateUniqueTexts генерирует уникальные тексты на основе базового заголовка и описания
// Без спантекса - использует синонимы и перестановки слов
func (tg *TextGenerator) GenerateUniqueTexts(baseTitle, baseDescription string, count int) ([]string, []string, error) {
	titles := make([]string, 0, count)
	descriptions := make([]string, 0, count)

	for i := 0; i < count; i++ {
		title := applyTextTransformations(baseTitle, tg.rnd)
		desc := applyTextTransformations(baseDescription, tg.rnd)

		// Убеждаемся в уникальности
		attempts := 0
		for (tg.used["title_"+title] || tg.used["desc_"+desc]) && attempts < 100 {
			title = applyTextTransformations(baseTitle, tg.rnd)
			desc = applyTextTransformations(baseDescription, tg.rnd)
			attempts++
		}

		tg.used["title_"+title] = true
		tg.used["desc_"+desc] = true
		titles = append(titles, title)
		descriptions = append(descriptions, desc)
	}

	return titles, descriptions, nil
}

// applyTextTransformations применяет трансформации к тексту для уникальности
func applyTextTransformations(text string, rnd *rand.Rand) string {
	// Синонимы и варианты замены
	synonyms := map[string][]string{
		"купить":    {"приобрести", "закупить", "заказать"},
		"продать":   {"продать", "продажа", "отдать"},
		"новый":     {"новый", "свежий", "новинка"},
		"качество":  {"качество", "качественный материал", "отличное качество"},
		"дерево":    {"дерево", "древесина", "лесоматериал"},
		"доска":     {"доска", "доска хвойная", "стружка"},
		"брус":      {"брус", "брус целой древесины", "строительный брус"},
		"горячо":    {"горячо", "спешка", "быстро"},
		"доставка":  {"доставка", "доставляем", "привозим", "доставим"},
		"цена":      {"цена", "стоимость", "стоимость товара"},
		"отличная":  {"отличная", "замечательная", "прекрасная"},
		"хорошая":   {"хорошая", "неплохая", "достойная"},
		"большая":   {"большая", "значительная", "великая"},
		"маленькая": {"маленькая", "небольшая", "компактная"},
	}

	result := text

	// Применяем случайные синонимы
	for word, replacements := range synonyms {
		if strings.Contains(strings.ToLower(result), word) && len(replacements) > 1 {
			choice := replacements[rnd.Intn(len(replacements))]
			// Замена только в нижнем регистре (чтобы не ломать начало предложения)
			result = replaceWord(result, word, choice)
		}
	}

	// Перестановка слов для коротких фраз
	words := strings.Fields(result)
	if len(words) >= 3 {
		// Иногда меняем порядок слов
		if rnd.Intn(3) == 0 {
			// Перемешиваем 2 случайных слова местами
			if len(words) >= 2 {
				i, j := rnd.Intn(len(words)), rnd.Intn(len(words))
				if i != j {
					words[i], words[j] = words[j], words[i]
					result = strings.Join(words, " ")
				}
			}
		}
	}

	return result
}

func replaceWord(text, old, new string) string {
	lowerText := strings.ToLower(text)
	lowerOld := strings.ToLower(old)
	result := text
	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerOld)
		if idx == -1 {
			break
		}
		actualIdx := start + idx
		result = result[:actualIdx] + new + result[actualIdx+len(old):]
		lowerText = strings.ToLower(result)
		start = actualIdx + len(new)
	}
	return result
}

type optionGroup struct {
	options []string
}

func extractGroups(template string) []optionGroup {
	var groups []optionGroup
	remaining := template

	for strings.Contains(remaining, "{") {
		start := strings.Index(remaining, "{")
		end := strings.Index(remaining[start:], "}")
		if end == -1 {
			break
		}
		end += start

		content := remaining[start+1 : end]
		parts := strings.Split(content, "|")
		options := make([]string, 0, len(parts))
		for _, p := range parts {
			options = append(options, strings.TrimSpace(p))
		}

		groups = append(groups, optionGroup{options: options})
		remaining = remaining[end+1:]
	}

	return groups
}

func replaceGroups(template string, groups []optionGroup, rnd *rand.Rand) string {
	result := template
	offset := 0

	for _, g := range groups {
		if len(g.options) == 0 {
			continue
		}
		choice := g.options[rnd.Intn(len(g.options))]

		openIdx := strings.Index(result[offset:], "{")
		if openIdx == -1 {
			continue
		}
		openIdx += offset
		closeIdx := strings.Index(result[openIdx:], "}")
		if closeIdx == -1 {
			continue
		}
		closeIdx += openIdx

		result = result[:openIdx] + choice + result[closeIdx+1:]
		offset = openIdx + len(choice)
	}

	return result
}

func calculateCombinations(groups []optionGroup) int {
	total := 1
	for _, g := range groups {
		if len(g.options) > 0 {
			total *= len(g.options)
		}
	}
	return total
}
