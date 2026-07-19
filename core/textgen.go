package core

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// TextGenerator генерирует уникальные вариации текста из шаблона
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
