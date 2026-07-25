package core

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"
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

// GenerateUniqueTitle создаёт уникальное название на основе шаблона
func (tg *TextGenerator) GenerateUniqueTitle(baseTitle string, index int, existingData [][]string, titleColIdx int) string {
	baseTitle = strings.TrimSpace(baseTitle)
	if baseTitle == "" {
		baseTitle = "Объявление"
	}

	candidates := []string{
		baseTitle,
		baseTitle + " (фото, характеристики)",
		baseTitle + " — доступен по предзаказу",
		strings.ReplaceAll(baseTitle, "брус", "стройбрус") + " для строительства",
		strings.ReplaceAll(baseTitle, "доска", "пиломатериал") + " сухая",
		baseTitle + " от производителя",
	}

	if index < len(candidates) {
		candidate := candidates[index]
		if len(candidate) > 0 {
			candidate = toUpperFirst(candidate)
		}
		if !tg.used["title_"+candidate] {
			tg.used["title_"+candidate] = true
			return candidate
		}
	}

	for attempt := 0; attempt < 200; attempt++ {
		words := strings.Fields(baseTitle)
		if len(words) >= 2 && attempt < len(candidates) {
			title := candidates[attempt%len(candidates)]
			if len(title) > 0 {
				title = toUpperFirst(title)
			}
			if !tg.used["title_"+title] {
				tg.used["title_"+title] = true
				return title
			}
		}
		suffixes := []string{" premium", " elite", " extra", " plus", " pro", ""}
		suffix := suffixes[attempt%len(suffixes)]
		title := baseTitle + suffix
		if len(title) > 0 {
			title = toUpperFirst(title)
		}
		if !tg.used["title_"+title] {
			tg.used["title_"+title] = true
			return title
		}
	}

	title := baseTitle
	if len(title) > 0 {
		title = toUpperFirst(title)
	}
	return title
}

// GenerateUniqueDescription создаёт уникальное описание в HTML формате
func (tg *TextGenerator) GenerateUniqueDescription(baseDescription string, index int) string {
	baseDescription = strings.TrimSpace(baseDescription)
	if baseDescription == "" {
		return "<p>Качественные материалы для строительства и отделки. Звоните для консультации.</p>"
	}

	paragraphs := strings.Split(baseDescription, "\n")
	var cleanParagraphs []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			cleanParagraphs = append(cleanParagraphs, p)
		}
	}

	if len(cleanParagraphs) == 0 {
		cleanParagraphs = []string{baseDescription}
	}

	variantSuffixes := []string{
		" Только качественные материалы. Звоните!",
		" Гарантия качества. Доставка по Москве и МО.",
		" Работаем с физ. и юр. лицами. НДС.",
		" От производителя. Свое производство.",
		" Бесплатная консультация. Работаем без выходных.",
	}

	uniqueHints := []string{
		" Индивидуальный подход к каждому клиенту.",
		" Профессиональная консультация бесплатно.",
		" Надежный поставщик с 2010 года.",
		" Работаем по всей России.",
		" Собственный автопарк для доставки.",
	}

	var result strings.Builder
	for i, p := range cleanParagraphs {
		if i == 1 && index < len(variantSuffixes) {
			p += variantSuffixes[index]
		}
		if strings.Contains(p, "–") || strings.Contains(p, "- ") || strings.Contains(p, ":") {
			parts := strings.Split(p, ":")
			if len(parts) > 1 {
				result.WriteString("<p><strong>")
				result.WriteString(parts[0])
				result.WriteString("</strong>:")
				for j := 1; j < len(parts); j++ {
					result.WriteString(parts[j])
					if j < len(parts)-1 {
						result.WriteString(":")
					}
				}
				result.WriteString("</p>\n")
				continue
			}
		}
		if strings.HasPrefix(p, "–") || strings.HasPrefix(p, "- ") {
			result.WriteString("<p><strong>")
			result.WriteString(strings.TrimPrefix(strings.TrimPrefix(p, "–"), "- "))
			result.WriteString("</strong></p>\n")
			continue
		}
		result.WriteString("<p>")
		result.WriteString(p)
		if len(cleanParagraphs) == 1 && index < len(uniqueHints) {
			result.WriteString(uniqueHints[index])
		}
		result.WriteString("</p>\n")
	}

	desc := strings.TrimSpace(result.String())
	if len(desc) > 7500 {
		desc = desc[:7497] + "..."
	}

	return desc
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

func toUpperFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
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

func ResolvePriceUnit(productType, defaultUnit string, index int) string {
	pt := strings.ToLower(strings.TrimSpace(productType))
	switch pt {
	case "брусок", "брус":
		units := []string{"Штуку", "м³"}
		return units[index%len(units)]
	case "доска":
		units := []string{"Штуку", "м³"}
		return units[index%len(units)]
	case "планкен", "вагонка":
		units := []string{"Штуку", "м²"}
		return units[index%len(units)]
	case "биг-бэг":
		units := []string{"Биг-бэг", "м³"}
		return units[index%len(units)]
	case "мешок":
		return "Мешок"
	case "россыпью":
		return "м³"
	default:
		units := []string{"Штуку", "Биг-бэг", "Мешок", "м²", "м³"}
		return units[index%len(units)]
	}
}
