package core

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
)

const (
	Birch                   = "Берёза"
	Karagach                = "Карагач"
	RedWood                 = "Красное дерево"
	Larch                   = "Лиственница"
	Magnolia                = "Магнолия"
	PinkWood                = "Розовое дерево"
	LumberBar               = "Брусок"
	Amaranth                = "Амарант"
	Balsa                   = "Бальса"
	Grenadilla              = "Гренадил"
	Other                   = "Другой"
	Merbau                  = "Мербау"
	Rosewood                = "Палисандр"
	Vagonka                 = "Вагонка"
	Acacia                  = "Акация"
	Bamboo                  = "Бамбук"
	Walnut                  = "Грецкий орех"
	Zebrawood               = "Зебрано"
	Chestnut                = "Каштан"
	Sandalwood              = "Самшит"
	Poplar                  = "Тополь"
	Cirico                  = "Цирикоте"
	ImitationBeamBlockHouse = "Имитация бревна, блок-хаус"
	ImitationBeamRauHaus    = "Имитация бруса, рау-хаус"
	Nalichnik               = "Наличник"
	Nashchelnik             = "Нащельник"
	Planken                 = "Планкен"
	Plinthus                = "Плинтус"
	Raskladka               = "Раскладка"
	Ugolok                  = "Уголок"
)

var LumberTypeWoods = map[string][]string{
	"Брус":                  {"Ель", "Липа", "Сосна"},
	LumberBar:               {"Ель", "Липа", "Сосна"},
	Vagonka:                 {"Абаш", "Ель", "Липа", "Сосна"},
	"Горбыль":               {"Ель", "Липа", "Сосна"},
	"Доска":                 {"Ель", "Липа", "Сосна"},
	"Дрова":                 {"Ель", "Липа", "Сосна"},
	Other:                   {Other, "Ель", "Липа", "Сосна"},
	ImitationBeamBlockHouse: {Other, "Ель", "Липа", "Сосна"},
	ImitationBeamRauHaus:    {Other, "Ель", "Липа", "Сосна"},
	"Лес-кругляк":           {"Ель", "Липа", "Сосна"},
	"Мебельный щит":         {"Абаш", "Ель", "Липа", "Сосна"},
	Nalichnik:               {"Липа", "Сосна"},
	"Настил":                {"Сосна"},
	Nashchelnik:             {"Липа", "Сосна"},
	"Оцилиндрованное бревно": {"Ель", "Сосна"},
	Planken:  {"Ель", "Липа", "Сосна"},
	Plinthus: {"Липа", "Сосна"},
	"Поддон": {"Сосна"},
	"Полок":  {"Ель", "Липа", "Сосна"},
	"Потолочный плинтус, галтель": {"Липа", "Сосна"},
	Raskladka:          {"Липа", "Сосна"},
	"Рейка":            {"Ель", "Липа", "Сосна"},
	"Слэб":             {"Липа", "Сосна"},
	"Столб для забора": {"Ель", "Липа", "Сосна"},
	Ugolok:             {"Липа", "Сосна"},
	"Шкант":            {"Ель", "Липа", "Сосна"},
	"Штапик":           {"Абаш", "Ель", "Липа", "Сосна"},
}

var PanelProfiles = map[string][]string{
	Vagonka: {"Евровагонка", "Софтлайн", "Штиль"},
	Planken: {"Прямой", "Скошенный"},
}

var EdgeLumberTypes = map[string]bool{
	"Доска": true,
}

var EdgeOptions = []string{"Обрезанная", "Необрезанная", "Шпунтованная", "Завальцованная"}

var GradeLumberTypes = map[string]bool{
	"Брус": true, LumberBar: true, Vagonka: true, "Горбыль": true, "Доска": true,
	Other: true, ImitationBeamBlockHouse: true, ImitationBeamRauHaus: true,
	"Мебельный щит": true, Nalichnik: true, Nashchelnik: true, Plinthus: true,
	"Полок": true, "Потолочный плинтус, галтель": true, Raskladka: true,
	"Рейка": true, Ugolok: true, "Штапик": true,
}

var GradeWoodTypes = map[string]bool{
	"Абаш": true, Other: true, "Ель": true,
	"Липа": true, Larch: true, "Сосна": true,
}

var GradeOptions = []string{"Отборный, экстра", "1 (A)", "1–2 (AB)", "1–3 (ABC)", "2 (B)", "2–3 (BC)", "3 (C)", "3–4 (CD)", "4 (D)"}

var MoistureLumberTypes = map[string]bool{
	"Доска": true, "Брус": true, ImitationBeamRauHaus: true,
	LumberBar: true, Planken: true, "Дрова": true, "Рейка": true, "Слэб": true,
}

var MoistureWoodTypes = map[string]bool{
	"Абаш": true, Other: true, "Ель": true,
	"Липа": true, Larch: true, "Сосна": true,
}

var MoistureOptions = []string{"Сухая", "Естественная"}

var ProfileOnlyLumberType = "Брус"
var ProfileOptions = []string{"Да", "Нет"}

var StructureLumberTypes = map[string]bool{
	"Брус": true, Vagonka: true, ImitationBeamRauHaus: true,
	LumberBar: true, Nalichnik: true, Nashchelnik: true, Plinthus: true,
	Raskladka: true, "Рейка": true, Ugolok: true,
}

var StructureOptions = []string{"Цельная", "Клеёная"}

var DefaultDimensions = map[string][]string{
	"thickness": {"20 мм", "30 мм", "40 мм", "50 мм"},
	"width":     {"100 мм", "150 мм", "200 мм"},
	"length":    {"2000 мм", "3000 мм", "4000 мм", "6000 мм"},
	"height":    {"40 мм", "50 мм", "60 мм"},
	"widthD":    {"100 мм", "150 мм", "200 мм"},
	"lengthD":   {"2000 мм", "3000 мм", "4000 мм"},
}

var DependentDimensions = map[string]map[string]map[string][]string{
	"thickness": {
		"В наличии": {
			"Доска":  {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "23 мм", "24 мм", "25 мм", "26 мм", "27 мм", "28 мм", "30 мм", "32 мм", "34 мм", "35 мм", "36 мм", "38 мм", "40 мм", "42 мм", "44 мм", "45 мм", "50 мм", "60 мм", "75 мм", "250 мм"},
			"Брус":   {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "23 мм", "24 мм", "25 мм", "26 мм", "27 мм", "28 мм", "30 мм", "32 мм", "34 мм", "35 мм", "36 мм", "38 мм", "40 мм", "42 мм", "44 мм", "45 мм", "50 мм", "60 мм", "75 мм", "250 мм"},
			"Брусок": {"10 мм", "16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Имитация бруса, рау-хаус": {"15 мм", "16 мм", "17 мм", "18 мм", "19 мм", "20 мм", "21 мм", "22 мм", "23 мм", "25 мм", "27 мм", "28 мм", "34 мм", "35 мм", "36 мм", "37 мм"},
			"Вагонка": {"12,5 мм", "13 мм", "13,5 мм", "14 мм", "15 мм", "16 мм", "18,5 мм", "19 мм", "21 мм", "22,5 мм", "25 мм", "26 мм", "28 мм"},
		},
	},
	"width": {
		"В наличии": {
			"Доска":  {"10 мм", "15 мм", "20 мм", "25 мм", "30 мм", "35 мм", "40 мм", "45 мм", "50 мм", "65 мм", "70 мм", "75 мм", "80 мм", "85 мм", "90 мм", "95 мм", "100 мм", "120 мм", "125 мм", "127 мм", "130 мм", "135 мм", "140 мм", "141 мм", "142 мм", "143 мм", "145 мм", "146 мм", "150 мм", "160 мм", "170 мм", "180 мм", "190 мм", "195 мм", "200 мм", "250 мм", "300 мм"},
			"Брус":   {"40 мм", "50 мм", "60 мм", "75 мм", "90 мм", "96 мм", "100 мм", "110 мм", "120 мм", "125 мм", "127 мм", "130 мм", "135 мм", "140 мм", "142 мм", "143 мм", "145 мм", "146 мм", "150 мм", "160 мм", "170 мм", "180 мм", "190 мм", "195 мм", "196 мм", "200 мм", "220 мм", "250 мм", "300 мм"},
			"Брусок": {"10 мм", "16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм", "110 мм", "120 мм", "128 мм", "130 мм", "140 мм", "146 мм", "150 мм", "160 мм", "170 мм", "180 мм", "190 мм", "196 мм", "200 мм"},
			"Имитация бруса, рау-хаус": {"120 мм", "121 мм", "127 мм", "135 мм", "138 мм", "140 мм", "143 мм", "145 мм", "146 мм", "150 мм", "158 мм", "160 мм", "170 мм", "173 мм", "175 мм", "176 мм", "185 мм", "190 мм", "192 мм", "193 мм", "195 мм", "196 мм"},
			"Вагонка": {"70 мм", "75 мм", "80 мм", "85 мм", "90 мм", "93 мм", "95 мм", "96 мм", "100 мм", "105 мм", "110 мм", "115 мм", "120 мм", "121 мм", "125 мм", "130 мм", "138 мм", "140 мм", "146 мм", "150 мм", "160 мм", "165 мм", "170 мм", "180 мм", "195 мм"},
		},
	},
	"length": {
		"В наличии": {
			"Доска":  {"300 мм", "600 мм", "800 мм", "900 мм", "1000 мм", "1100 мм", "1200 мм", "1400 мм", "1500 мм", "2000 мм", "2500 мм", "2700 мм", "2900 мм", "3000 мм", "3500 мм", "4000 мм", "4500 мм", "5000 мм", "5500 мм", "6000 мм", "9000 мм", "12000 мм", "15000 мм"},
			"Брус":   {"1000 мм", "1500 мм", "2000 мм", "2100 мм", "2400 мм", "2500 мм", "3000 мм", "3500 мм", "4000 мм", "4500 мм", "5000 мм", "5500 мм", "6000 мм", "9000 мм", "12000 мм", "15000 мм"},
			"Брусок": {"1000 мм", "2100 мм", "2200 мм", "2400 мм", "3000 мм", "4000 мм", "5000 мм", "6000 мм"},
			"Имитация бруса, рау-хаус": {"1000 мм", "1500 мм", "2000 мм", "2100 мм", "2400 мм", "2500 мм", "2700 мм", "3000 мм", "3300 мм", "3600 мм", "3900 мм", "4000 мм", "4200 мм", "4500 мм", "4800 мм", "5000 мм", "5100 мм", "5400 мм", "5700 мм", "6000 мм"},
			"Вагонка": {"500 мм", "700 мм", "800 мм", "900 мм", "1000 мм", "1500 мм", "2000 мм", "2200 мм", "2300 мм", "2400 мм", "2500 мм", "2700 мм", "2900 мм", "3000 мм", "3500 мм", "4000 мм", "6000 мм"},
		},
	},
	"height": {
		"": {
			"Имитация бревна, блок-хаус": {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "23 мм", "24 мм", "25 мм", "26 мм", "27 мм", "28 мм", "30 мм", "32 мм", "34 мм", "35 мм", "36 мм", "38 мм", "40 мм", "42 мм", "44 мм", "45 мм", "50 мм", "60 мм", "75 мм", "250 мм"},
			"Планкен":       {"12,5 мм", "13 мм", "13,5 мм", "14 мм", "15 мм", "16 мм", "18,5 мм", "19 мм", "21 мм", "22,5 мм", "25 мм", "26 мм", "28 мм"},
			"Мебельный щит": {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Наличник":      {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Нащельник":     {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Настил":        {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Полок":         {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Плинтус":       {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Поддон":        {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Потолочный плинтус, галтель": {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Раскладка": {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Рейка":     {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Слэб":      {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Уголок":    {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
			"Штапик":    {"16 мм", "18 мм", "19 мм", "20 мм", "22 мм", "25 мм", "30 мм", "32 мм", "35 мм", "40 мм", "44 мм", "45 мм", "50 мм", "60 мм", "64 мм", "70 мм", "75 мм", "80 мм", "88 мм", "90 мм", "95 мм", "96 мм", "100 мм"},
		},
	},
}

var AllLumberTypes = func() []string {
	keys := make([]string, 0, len(LumberTypeWoods))
	for k := range LumberTypeWoods {
		keys = append(keys, k)
	}
	return keys
}()

var AllWoodTypes = func() []string {
	set := make(map[string]bool)
	for _, woods := range LumberTypeWoods {
		for _, w := range woods {
			set[w] = true
		}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}()

func IsValidEdge(lumberType string) bool {
	return EdgeLumberTypes[lumberType]
}

func IsValidGrade(lumberType, woodType string) bool {
	return GradeLumberTypes[lumberType] && GradeWoodTypes[woodType]
}

func IsValidMoisture(lumberType, woodType string) bool {
	return MoistureLumberTypes[lumberType] && MoistureWoodTypes[woodType]
}

func IsValidProfile(lumberType string) bool {
	return lumberType == ProfileOnlyLumberType
}

func IsValidStructure(lumberType string) bool {
	return StructureLumberTypes[lumberType]
}

func GetValidLumberProfiles(lumberType string) []string {
	return PanelProfiles[lumberType]
}

func GetValidWoodTypes(lumberType string) []string {
	return LumberTypeWoods[lumberType]
}

func FilterValid(options []string, validFn func(string) bool) []string {
	var out []string
	for _, opt := range options {
		if validFn(opt) {
			out = append(out, opt)
		}
	}
	return out
}

func PickRandom(r *rand.Rand, options []string) string {
	if len(options) == 0 {
		return ""
	}
	return options[r.Intn(len(options))]
}

func InArray(v string, arr []string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}

func GetDimensionValues(dimension string) []string {
	if vals, ok := DefaultDimensions[dimension]; ok {
		return vals
	}
	return nil
}

func GetDependentDimensions(dimension, availability, lumberType string) []string {
	if availMap, ok := DependentDimensions[dimension]; ok {
		if ltMap, ok := availMap[availability]; ok {
			if vals, ok := ltMap[lumberType]; ok {
				return vals
			}
		}
		if ltMap, ok := availMap[""]; ok {
			if vals, ok := ltMap[lumberType]; ok {
				return vals
			}
		}
	}
	return nil
}

func ParseDimensionNumber(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "по запросу") || strings.Contains(lower, "нет") || strings.Contains(lower, "не указан") {
		return 0
	}

	numEnd := 0
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			numEnd = i + 1
			continue
		}
		if ch == '.' || ch == ',' {
			continue
		}
		break
	}
	if numEnd == 0 {
		return 0
	}

	num, err := strconv.ParseFloat(strings.ReplaceAll(value[:numEnd], ",", "."), 64)
	if err != nil || num <= 0 {
		return 0
	}
	return num
}

func DimensionUnit(value string) string {
	value = strings.TrimSpace(value)
	numEnd := 0
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			numEnd = i + 1
			continue
		}
		if ch == '.' || ch == ',' {
			continue
		}
		break
	}
	return strings.TrimSpace(value[numEnd:])
}

// SelectNearestDimension возвращает ближайшее значение из каталога к оригиналу.
// Если каталог пустой — возвращает оригинал.
func SelectNearestDimension(original string, options []string) string {
	if len(options) == 0 {
		return original
	}

	origNum := ParseDimensionNumber(original)
	if origNum == 0 {
		return original
	}

	unit := DimensionUnit(original)
	best := options[0]
	bestDiff := math.MaxFloat64
	for _, opt := range options {
		optNum := ParseDimensionNumber(opt)
		if optNum == 0 {
			continue
		}
		optUnit := DimensionUnit(opt)
		if unit != "" && optUnit != "" && unit != optUnit {
			continue
		}
		diff := math.Abs(origNum - optNum)
		if diff < bestDiff {
			bestDiff = diff
			best = opt
		}
	}
	return best
}

// IsConfigurationUnique проверяет, уникальна ли конфигурация глобально.
func IsConfigurationUnique(sig string) bool {
	globalUsedMu.Lock()
	defer globalUsedMu.Unlock()
	if globalUsedSigs[sig] {
		return false
	}
	globalUsedSigs[sig] = true
	return true
}
