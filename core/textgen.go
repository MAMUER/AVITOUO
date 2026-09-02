package core

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	globalUsedMu     sync.Mutex
	globalUsedTitles = make(map[string]bool)
	globalUsedDescs  = make(map[string]bool)
	globalUsedSigs   = make(map[string]bool)
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

// GenerateUniqueTitle создаёт уникальное название на основе базового текста
func (tg *TextGenerator) GenerateUniqueTitle(baseTitle string, index int, existingData [][]string, titleColIdx int) string {
	baseTitle = strings.TrimSpace(baseTitle)
	if baseTitle == "" {
		baseTitle = "Объявление"
	}

	candidates := []string{
		baseTitle,
		baseTitle + " (фото, характеристики)",
		strings.ToUpper(baseTitle),
	}

	for _, c := range candidates {
		if c == "" {
			continue
		}
		c = toUpperFirst(c)
		globalUsedMu.Lock()
		exists := globalUsedTitles["title_"+c]
		globalUsedMu.Unlock()
		if !exists {
			globalUsedMu.Lock()
			globalUsedTitles["title_"+c] = true
			tg.used["title_"+c] = true
			globalUsedMu.Unlock()
			return c
		}
	}

	suffixes := []string{
		" премиум-класса",
		" высокого качества",
		" строительный",
		" обрезной",
		" необрезной",
		" строганый",
		" сухой",
		" точные размеры",
		" фабричный",
		" в наличии",
		" под заказ",
		" стандарт качества",
		" напрямую от производителя",
		" без посредников",
		" с доставкой",
		" торг уместен",
		" для строительства",
		" для отделки",
		" для кровли",
		" для фасада",
		" для ремонта",
		" размерный ряд",
		" новый",
		" сертифицированный",
		" натуральный",
		" проверенное качество",
	}

	for attempt := 0; attempt < 500; attempt++ {
		suffix := suffixes[tg.rnd.Intn(len(suffixes))]
		title := baseTitle + suffix
		title = toUpperFirst(title)
		globalUsedMu.Lock()
		exists := globalUsedTitles["title_"+title]
		globalUsedMu.Unlock()
		if !exists {
			globalUsedMu.Lock()
			globalUsedTitles["title_"+title] = true
			tg.used["title_"+title] = true
			globalUsedMu.Unlock()
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
func (tg *TextGenerator) GenerateUniqueDescription(baseDescription string, index int, params GenerateDescriptionParams) string {
	_ = baseDescription

	lumberType := strings.TrimSpace(params.LumberType)
	if lumberType == "" {
		lumberType = "пиломатериал"
	}
	woodType := strings.TrimSpace(params.WoodType)
	if woodType == "" {
		woodType = "сосна/ель"
	}
	grade := strings.TrimSpace(params.Grade)
	if grade == "" {
		grade = "Экстра"
	}
	priceUnit := strings.TrimSpace(params.PriceUnit)
	if priceUnit == "" {
		priceUnit = "шт."
	}
	height := strings.TrimSpace(params.Height)
	width := strings.TrimSpace(params.Width)
	length := strings.TrimSpace(params.Length)
	if height == "" {
		height = "50"
	}
	if width == "" {
		width = "150"
	}
	if length == "" {
		length = "3000"
	}

	descriptiveWords := []string{
		"премиум", "высококачественный", "качественный", "отличный", "натуральный", "фабричный", "проверенный",
	}
	descWord := descriptiveWords[index%len(descriptiveWords)]

	alsoAvailablePool := []string{
		"брус строганная доска обрезная доска; имитация бруса хвойная планкен хвоя и лиственница вагонка террасная доска доска пола крепеж",
		"брус; доска строганная; доска обрезная; имитация бруса; планкен хвойный; вагонка; террасная доска; доска пола; крепеж",
		"брус, строганная доска, обрезная доска, имитация бруса, хвойная планкен, вагонка, террасная доска, доска пола, крепеж",
	}
	alsoAvailable := alsoAvailablePool[index%len(alsoAvailablePool)]

	ctaPool := []string{
		"📞 Звоните или пишите в чат Авито! Бесплатно проконсультируем, поможем с расчётом количества под ваш проект.",
		"📞 ЗВОНИТЕ ИЛИ ПИШИТЕ ПРЯМО СЕЙЧАС! Отправим актуальные фото/видео, рассчитаем стоимость, забронируем объём.",
		"📞 Пишите или звоните! Рассчитаем объём, подберём материал под проект, забронируем нужную партию.",
		"📞 Звоните! Бесплатно рассчитаем объём, подберём оптимальный вариант, оформим доставку.",
	}
	cta := ctaPool[index%len(ctaPool)]

	build := func(descWord, alsoAvailable, cta string) string {
		var b strings.Builder
		fmt.Fprintf(&b, "<p>👉<strong>ЧТОБЫ ОЗНАКОМИТЬСЯ ПОДРОБНЕЕ</strong> с нашим ассортиментом, напишите в чат слово<strong>: \"КАТАЛОГ\"</strong></p>")
		fmt.Fprintf(&b, "<p><strong>%s %s %s×%s×%s мм, хвоя (%s), сорт %s, камерной сушки. В НАЛИЧИИ на складе в МЫТИЩАХ!</strong></p>", lumberType, descWord, height, width, length, woodType, grade)
		fmt.Fprintf(&b, "<p><strong>Цена указана за %s!</strong></p>", priceUnit)
		b.WriteString("<p><strong>Характеристики:</strong></p>")
		fmt.Fprintf(&b, "<p>• <strong>Порода</strong>: хвоя (%s)</p>", woodType)
		fmt.Fprintf(&b, "<p>• <strong>Сорт</strong>: %s</p>", grade)
		fmt.Fprintf(&b, "<p>• <strong>Размер</strong>: %s мм × %s мм × %s мм</p>", height, width, length)
		b.WriteString("<p>• <strong>Влажность</strong>: 8–15% (камерная сушка — не коробится, не трескается)</p>")
		b.WriteString("<p>• <strong>Обработка</strong>: строганная с 4 сторон — готова к монтажу</p>")
		b.WriteString("<ul>")
		b.WriteString("<li>Камерная сушка</li>")
		b.WriteString("<li>Строганная с четырех сторон</li>")
		b.WriteString("<li>Гладкая поверхность без дополнительной обработки</li>")
		b.WriteString("<li>Подходит для внутренних и наружных работ</li>")
		b.WriteString("<li>Реальные фотографии товара со склада</li>")
		b.WriteString("</ul>")
		fmt.Fprintf(&b, "<p><strong>В НАЛИЧИИ ТАКЖЕ: </strong>%s</p>", alsoAvailable)
		b.WriteString("<p>📍 <strong>Самовывоз — 2 точки в г. Мытищи:</strong></p>")
		b.WriteString("<p>• Осташковское ш., 1Б, стр. 7, ангар №15 (у въезда под аркой «Стройдвор Яуза»)</p>")
		b.WriteString("<p>• Волковское ш., стр. 21А</p>")
		b.WriteString("<p><strong>Ежедневно 9:00–18:00 (без выходных)</strong></p>")
		b.WriteString("<p><strong>Доставка по Москве и МО</strong> | Отправка в регионы транспортными компаниями</p>")
		b.WriteString("<p><strong>Оплата:</strong> наличные, банковская карта, перевод, QR-код, безнал с/без НДС</p>")
		fmt.Fprintf(&b, "<p>📞 <strong>%s</strong> Бесплатно проконсультируем, поможем с расчётом количества под ваш проект.</p>", cta)
		b.WriteString("<p><strong>Добавьте в избранное</strong> — всегда в курсе свежих поставок и спецпредложений!</p>")
		b.WriteString("<p>Работаем с физ. и юр. лицами. Гарантия качества на всю продукцию!</p>")
		return b.String()
	}

	desc := build(descWord, alsoAvailable, cta)
	key := "desc_" + desc
	globalUsedMu.Lock()
	exists := globalUsedDescs[key]
	globalUsedMu.Unlock()
	if !exists {
		globalUsedMu.Lock()
		globalUsedDescs[key] = true
		tg.used[key] = true
		globalUsedMu.Unlock()
		return desc
	}

	for attempt := 0; attempt < 50; attempt++ {
		descWord = descriptiveWords[(index+attempt)%len(descriptiveWords)]
		alsoAvailable = alsoAvailablePool[(index+attempt)%len(alsoAvailablePool)]
		cta = ctaPool[(index+attempt)%len(ctaPool)]

		desc := build(descWord, alsoAvailable, cta)
		key := "desc_" + desc
		globalUsedMu.Lock()
		exists := globalUsedDescs[key]
		globalUsedMu.Unlock()
		if !exists {
			globalUsedMu.Lock()
			globalUsedDescs[key] = true
			tg.used[key] = true
			globalUsedMu.Unlock()
			return desc
		}
	}

	return build(descWord, alsoAvailable, cta)
}

// GenerateDescriptionParams содержит параметры для генерации описания
type GenerateDescriptionParams struct {
	LumberType string
	WoodType   string
	Grade      string
	Height     string
	Width      string
	Length     string
	PriceUnit  string
}

// GenerateVariations генерирует N уникальных вариаций из шаблона
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

// ResolvePriceUnit возвращает единицу измерения для типа товара
func ResolvePriceUnit(productType, defaultUnit string, index int) string {
	pt := strings.ToLower(strings.TrimSpace(productType))
	switch pt {
	case "брусок", "брус", "доска":
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

// VaryDimension возвращает вариацию размера, близкую к оригиналу, только в мм.
// Значения в других единицах не конвертируются и возвращаются как есть.
func VaryDimension(value string, r *rand.Rand) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "по запросу") || strings.Contains(lower, "нет") || strings.Contains(lower, "не указан") {
		return value
	}

	numEnd := 0
	hasDecimal := false
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			numEnd = i + 1
			continue
		}
		if ch == '.' || ch == ',' {
			hasDecimal = true
			continue
		}
		break
	}
	if numEnd == 0 {
		return value
	}

	num, err := strconv.ParseFloat(strings.ReplaceAll(value[:numEnd], ",", "."), 64)
	if err != nil || num <= 0 {
		return value
	}

	unit := strings.TrimSpace(value[numEnd:])
	unitLower := strings.ToLower(unit)
	if unitLower == "" || strings.Contains(unitLower, "мм") || strings.Contains(unitLower, "mm") {
		unit = "мм"
	} else {
		return value
	}

	delta := num * 0.1
	if delta < 1 {
		delta = 1
	}
	change := r.Float64()*2*delta - delta
	newNum := num + change
	if newNum < 1 {
		newNum = 1
	}
	if hasDecimal {
		newNum = math.Round(newNum*10) / 10
	} else {
		newNum = math.Round(newNum)
	}
	newNumStr := strconv.FormatFloat(newNum, 'f', -1, 64)
	if hasDecimal {
		newNumStr = strings.TrimRight(newNumStr, "0")
		newNumStr = strings.TrimRight(newNumStr, ".")
	}
	return newNumStr + " " + unit
}
