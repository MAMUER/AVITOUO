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
		" Склад в Мытищах",
		" купить в Мытищах",
		" Доставка по Москве",
		" камерной сушки",
		" сухой строганный",
		" Сорт АВ",
		" Сорт Экстра",
		" Хвоя",
		" Лиственница",
		" для строительства",
		" для отделки",
		" Цена за м³",
		" Цена за шт.",
		" от производителя",
		" Гарантия качества",
		" Работаем по всей России",
		" Бесплатная консультация",
		" Собственное производство",
		" Сорт Прима",
		" ГОСТ",
		" 3 м и 6 м",
		" в наличии",
		" под заказ",
		" Торг уместен",
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

	var buildDesc func() string
	buildDesc = func() string {
		r := tg.rnd
		var result strings.Builder

		result.WriteString("<p>")
		result.WriteString(catalogIntros[r.Intn(len(catalogIntros))])
		result.WriteString("</p>\n")

		para := cleanParagraphs[r.Intn(len(cleanParagraphs))]
		result.WriteString("<p>")
		result.WriteString(productLeads[r.Intn(len(productLeads))])
		result.WriteString("<strong>")
		result.WriteString(para)
		result.WriteString("</strong>")
		if r.Intn(2) == 0 {
			result.WriteString(" — ")
			result.WriteString(descriptionLeads[r.Intn(len(descriptionLeads))])
		}
		result.WriteString("</p>\n")

		if r.Intn(2) == 0 {
			result.WriteString("<p><strong>✅ ПОЧЕМУ ЭТО ЛУЧШИЙ ВЫБОР:</strong></p>\n")
			benefits := []string{
				"✨ Стабильная геометрия — точные размеры по всем сторонам, готов к монтажу.",
				"✨ Камерная сушка 8–12% — минимальная усадка, устойчивость к перепадам влажности.",
				"✨ Чистовая строжка с 4-х сторон — гладкая поверхность, готова к покраске.",
				"✨ Сорт АВ/Экстра — без выпадающих сучков, однородная текстура.",
				"✨ Прямые поставки с производства — цены ниже рынка без посредников.",
			}
			for i := 0; i < 3; i++ {
				result.WriteString("<p>")
				result.WriteString(benefits[r.Intn(len(benefits))])
				result.WriteString("</p>\n")
			}
		}

		result.WriteString("<p><strong>✅ ХАРАКТЕРИСТИКИ:</strong></p>\n")
		chars := []string{
			"• Размер: по запросу в наличии (ширина × толщина × длина)",
			"• Материал: хвоя (сосна/ель), лиственница — под запас",
			"• Обработка: камерная сушка 8-12%, чистовая строжка с 4-х сторон",
			"• Сорт: АВ, Экстра, Прима — без выпадающих сучков и черноты",
			"• Поверхность: гладкая, ровная — готова к покраске и монтажу",
		}
		for _, c := range chars {
			if r.Intn(4) != 0 {
				result.WriteString("<p>")
				result.WriteString(c)
				result.WriteString("</p>\n")
			}
		}

		result.WriteString("<p><strong>📍 Самовывоз в г. Мытищи (2 точки):</strong></p>\n")
		result.WriteString("<p>1️⃣ Осташковское ш., 1Б, стр. 7, ангар №15 (под аркой «Стройдвор Яуза»)</p>\n")
		result.WriteString("<p>2️⃣ Волковское ш., стр. 21А</p>\n")
		result.WriteString("<p>🕒 Ежедневно 9:00–18:00 (без выходных)</p>\n")

		deliveryOptions := []string{
			"🚚 Доставка по Москве и МО. Отправка в регионы через ТК",
			"🚛 Доставка по Москве и области. Регионы — транспортными компаниями",
		}
		result.WriteString("<p>")
		result.WriteString(deliveryOptions[r.Intn(len(deliveryOptions))])
		result.WriteString("</p>\n")

		paymentOptions := []string{
			"💳 Оплата: наличные, карта, перевод, QR, безнал с НДС / без НДС",
			"💳 Принимаем: наличные, банковская карта, перевод, безналичный расчёт (с НДС/без НДС)",
		}
		result.WriteString("<p>")
		result.WriteString(paymentOptions[r.Intn(len(paymentOptions))])
		result.WriteString("</p>\n")

		if r.Intn(2) == 0 {
			result.WriteString("<p>")
			result.WriteString(suffixPool[r.Intn(len(suffixPool))])
			result.WriteString("</p>\n")
		}

		result.WriteString("<p>")
		result.WriteString(ctaPool[r.Intn(len(ctaPool))])
		result.WriteString("</p>\n")

		result.WriteString("<p>🔹 Добавьте объявление в избранное — всегда в курсе свежих поступлений и спецпредложений!</p>\n")

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
