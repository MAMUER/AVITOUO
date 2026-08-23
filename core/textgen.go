package core

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

const pClose = "</p>\n"

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
		if !tg.used["title_"+c] {
			tg.used["title_"+c] = true
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
		baseDescription = "Качественный пиломатериал премиум-класса для строительства и отделки."
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

	catalogIntros := []string{
		"👉 Чтобы увидеть весь ассортимент, напишите в чат: «КАТАЛОГ»",
		"✍️ Напишите «КАТАЛОГ» — вышлем полный прайс-лист в личные сообщения!",
		"📋 Актуальный каталог по запросу. Пишите слово «КАТАЛОГ»!",
		"📑 Чтобы получить полный каталог, напишите «КАТАЛОГ» в чат.",
		"✉️ Присылайте запрос — вышлем полный прайс-лист с ценами на всю продукцию.",
		"💬 Жмите «КАТАЛОГ» — вышлем актуальные остатки и цены прямо сейчас!",
		"📩 Запросите «КАТАЛОГ» — отправим актуальный прайс в личные сообщения.",
	}

	productLeads := []string{
		"🔥 ",
		"🌲 ",
		"💎 ",
		"⭐ ",
		"🏆 ",
	}

	descriptionLeads := []string{
		"Высококачественный пиломатериал для строительства и отделки. Камерная сушка и чистовая строжка гарантируют точную геометрию и долговечность.",
		"Идеальное решение для строительных и отделочных работ. Стабильная геометрия, точные размеры и устойчивость к влаге.",
		"Премиальный пиломатериал для надёжных конструкций. Обработка на современном оборудовании, контроль влажности 8–12%.",
		"Готовый материал к монтажу и покраске. Гладкая поверхность, ровные кромки, точный размер по всем сторонам.",
		"Проверенное качество для частного и коммерческого строительства. Выдерживает нагрузку, не коробится со временем.",
	}

	suffixPool := []string{
		" Только натуральные материалы. звоните!",
		" Гарантия качества. Доставка по Москве и МО.",
		" Работаем с физ. и юр. лицами. НДС.",
		" От производителя. Собственное производство.",
		" Бесплатная консультация. Работаем без выходных.",
		" Прямые поставки без посредников. Отгружаем сегодня.",
		" Сертифицированный материал. Соответствует ГОСТ.",
		" Фото по запросу. Скидки за объём.",
		" Бесплатный расчёт под ваш проект. Уточняйте актуальные цены.",
	}

	ctaPool := []string{
		"📞 Звоните или пишите в чат! Бесплатно проконсультируем, рассчитаем объём и подберём материал под вашу задачу.",
		"📞 ЗВОНИТЕ ИЛИ ПИШИТЕ ПРЯМО СЕЙЧАС! Отправим актуальные фото/видео, рассчитаем стоимость, забронируем объём.",
		"📞 Пишите или звоните! Рассчитаем объём, подберём материал под проект, забронируем нужную партию.",
		"📞 Звоните! Бесплатно рассчитаем объём, подберём оптимальный вариант, оформим доставку.",
	}

	var buildDesc = func() string {
		r := tg.rnd
		var result strings.Builder

		writeRandomParagraph(&result, r, catalogIntros)

		para := pickRandom(r, cleanParagraphs)
		result.WriteString("<p>")
		result.WriteString(pickRandom(r, productLeads))
		result.WriteString("<strong>")
		result.WriteString(para)
		result.WriteString("</strong>")
		if r.Intn(2) == 0 {
			result.WriteString(" — ")
			result.WriteString(pickRandom(r, descriptionLeads))
		}
		result.WriteString(pClose)

		if r.Intn(2) == 0 {
			writeBenefits(&result, r)
		}
		writeCharacteristics(&result, r)

		writeAddress(&result)
		writeDelivery(&result, r)
		writePayment(&result, r)
		writeFooter(&result, r, suffixPool, ctaPool)

		return strings.TrimSpace(result.String())
	}

	for attempt := 0; attempt < 100; attempt++ {
		desc := buildDesc()
		key := "desc_" + desc
		if !tg.used[key] {
			tg.used[key] = true
			return desc
		}
	}

	desc := buildDesc()
	if desc != "" && !tg.used["desc_"+desc] {
		tg.used["desc_"+desc] = true
	}
	return desc
}

func writeBenefits(result *strings.Builder, r *rand.Rand) {
	result.WriteString("<p><strong>✅ ПОЧЕМУ ЭТО ЛУЧШИЙ ВЫБОР:</strong>")
	result.WriteString(pClose)
	benefits := []string{
		"✨ Стабильная геометрия — точные размеры по всем сторонам, готов к монтажу.",
		"✨ Камерная сушка 8–12% — минимальная усадка, устойчивость к перепадам влажности.",
		"✨ Чистовая строжка с 4-х сторон — гладкая поверхность, готова к покраске.",
		"✨ Сорт АВ/Экстра — без выпадающих сучков, однородная текстура.",
		"✨ Прямые поставки с производства — цены ниже рынка без посредников.",
	}
	for i := 0; i < 3; i++ {
		writeParagraph(result, benefits[r.Intn(len(benefits))])
	}
}

func writeCharacteristics(result *strings.Builder, r *rand.Rand) {
	result.WriteString("<p><strong>✅ ХАРАКТЕРИСТИКИ:</strong>")
	result.WriteString(pClose)
	chars := []string{
		"• Размер: по запросу в наличии (ширина × толщина × длина)",
		"• Материал: хвоя (сосна/ель), лиственница — под запас",
		"• Обработка: камерная сушка 8-12%, чистовая строжка с 4-х сторон",
		"• Сорт: АВ, Экстра, Прима — без выпадающих сучков и черноты",
		"• Поверхность: гладкая, ровная — готова к покраске и монтажу",
	}
	for _, c := range chars {
		if r.Intn(4) != 0 {
			writeParagraph(result, c)
		}
	}
}

func writeAddress(result *strings.Builder) {
	result.WriteString("<p><strong>📍 Самовывоз в г. Мытищи (2 точки):</strong>")
	result.WriteString(pClose)
	result.WriteString("<p>1️⃣ Осташковское ш., 1Б, стр. 7, ангар №15 (под аркой «Стройдвор Яуза»)")
	result.WriteString(pClose)
	result.WriteString("<p>2️⃣ Волковское ш., стр. 21А")
	result.WriteString(pClose)
	result.WriteString("<p>🕒 Ежедневно 9:00–18:00 (без выходных)")
	result.WriteString(pClose)
}

func writeDelivery(result *strings.Builder, r *rand.Rand) {
	deliveryOptions := []string{
		"🚚 Доставка по Москве и МО. Отправка в регионы через ТК",
		"🚛 Доставка по Москве и области. Регионы — транспортными компаниями",
	}
	writeParagraph(result, deliveryOptions[r.Intn(len(deliveryOptions))])
}

func writePayment(result *strings.Builder, r *rand.Rand) {
	paymentOptions := []string{
		"💳 Оплата: наличные, карта, перевод, QR, безнал с НДС / без НДС",
		"💳 Принимаем: наличные, банковская карта, перевод, безналичный расчёт (с НДС/без НДС)",
	}
	writeParagraph(result, paymentOptions[r.Intn(len(paymentOptions))])
}

func writeFooter(result *strings.Builder, r *rand.Rand, suffixPool, ctaPool []string) {
	if r.Intn(2) == 0 {
		writeRandomParagraph(result, r, suffixPool)
	}

	writeRandomParagraph(result, r, ctaPool)

	result.WriteString("<p>🔹 Добавьте объявление в избранное — всегда в курсе свежих поступлений и спецпредложений!")
	result.WriteString(pClose)
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

func pickRandom(r *rand.Rand, options []string) string {
	if len(options) == 0 {
		return ""
	}
	return options[r.Intn(len(options))]
}

func writeParagraph(result *strings.Builder, content string) {
	result.WriteString("<p>")
	result.WriteString(content)
	result.WriteString(pClose)
}

func writeRandomParagraph(result *strings.Builder, r *rand.Rand, options []string) {
	if len(options) == 0 {
		return
	}
	writeParagraph(result, options[r.Intn(len(options))])
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
