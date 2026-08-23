package core

import "math/rand"

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
