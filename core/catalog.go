package core

import "math/rand"

const (
	Birch = "Берёза"
	Karagach = "Карагач"
	RedWood = "Красное дерево"
	Larch = "Лиственница"
	Magnolia = "Магнолия"
	PinkWood = "Розовое дерево"
	LumberBar = "Брусок"
	Amaranth = "Амарант"
	Balsa = "Бальса"
	Grenadilla = "Гренадил"
	Other = "Другой"
	Merbau = "Мербау"
	Rosewood = "Палисандр"
	Vagonka = "Вагонка"
	Acacia = "Акация"
	Bamboo = "Бамбук"
	Walnut = "Грецкий орех"
	Zebrawood = "Зебрано"
	Chestnut = "Каштан"
	Sandalwood = "Самшит"
	Poplar = "Тополь"
	Cirico = "Цирикоте"
	ImitationBeamBlockHouse = "Имитация бревна, блок-хаус"
	ImitationBeamRauHaus = "Имитация бруса, рау-хаус"
	Nalichnik = "Наличник"
	Nashchelnik = "Нащельник"
	Planken = "Планкен"
	Plinthus = "Плинтус"
	Raskladka = "Раскладка"
	Ugolok = "Уголок"
)

var LumberTypeWoods = map[string][]string{
	"Брус":    {Birch, "Бук", "Дуб", "Ель", Karagach, "Кедр", "Клён", RedWood, "Липа", Larch, Magnolia, "Орех", "Осина", "Падук", "Пихта", PinkWood, "Сосна", "Тик", "Ясень"},
	LumberBar:  {Amaranth, Balsa, Birch, "Бук", "Венге", "Граб", Grenadilla, Other, "Дуб", "Ель", Karagach, "Кедр", RedWood, "Липа", Larch, Magnolia, Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, PinkWood, "Сосна", "Тик", "Ясень"},
	Vagonka: {"Абаш", Birch, "Бук", Other, "Дуб", "Ель", "Кедр", "Липа", Larch, Magnolia, "Ольха", "Орех", "Осина", "Падук", "Пихта", "Сосна", "Тик", "Ясень"},
	"Горбыль": {Birch, Other, "Дуб", "Ель", "Кедр", "Клён", "Липа", Larch, Magnolia, "Ольха", "Орех", "Осина", "Падук", "Пихта", "Сосна", "Тик", "Ясень"},
	"Доска":   {Other, "Ель", "Липа", "Сосна"},
	"Дрова":   {Acacia, Birch, "Бук", "Груша", Other, "Дуб", "Ель", Karagach, "Кедр", "Клён", "Липа", Larch, "Ольха", "Орех", "Осина", "Падук", "Пихта", "Сосна", "Тик", "Ясень"},
	Other:  {Acacia, Amaranth, Bamboo, Birch, "Бук", "Венге", "Вишня", "Граб", Walnut, "Груша", "ДПК", Other, "Дуб", "Ель", Zebrawood, Karagach, Chestnut, "Кедр", "Клён", RedWood, "Липа", Larch, Magnolia, Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, "Пихта", Sandalwood, "Сосна", "Тик", Poplar, Cirico, "Ясень"},
	ImitationBeamBlockHouse:  {Other, "Дуб", "Ель", "Кедр", "Липа", Larch, "Ольха", "Сосна", "Ясень"},
	ImitationBeamRauHaus:    {Other, "Дуб", "Ель", "Кедр", "Липа", Larch, "Ольха", "Осина", "Сосна", "Ясень"},
	"Лес-кругляк":                 {Acacia, Birch, Other, "Дуб", "Ель", Karagach, "Кедр", "Клён", "Липа", Larch, "Ольха", "Осина", "Пихта", Sandalwood, "Сосна", "Ясень"},
	"Мебельный щит":               {"Абаш", Acacia, Amaranth, "Бакаут", Balsa, Bamboo, Birch, "Бук", "Венге", "Вишня", "Граб", Grenadilla, Walnut, "Груша", "ДПК", Other, "Дуб", "Ель", Zebrawood, "Ироко", Karagach, Chestnut, "Кедр", "Клён", RedWood, "Липа", Larch, Magnolia, "Меранти", Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, "Пихта", PinkWood, Sandalwood, "Сосна", "Тик", Poplar, Cirico, "Чёрное дерево", "Ясень"},
	Nalichnik:                    {"Бук", Other, "Дуб", "Липа", Larch, "Осина", "Падук", "Сосна", "Тик"},
	"Настил":                      {Birch, Other, "Дуб", Larch, "Осина", "Сосна", "Тик"},
	Nashchelnik:                   {Other, "Липа", "Ольха", "Сосна", "Тик"},
	"Оцилиндрованное бревно":      {Other, "Дуб", "Ель", "Кедр", Larch, "Орех", "Сосна", "Тик", "Ясень"},
	Planken:                     {"ДПК", Other, "Дуб", "Ель", "Кедр", "Липа", Larch, Magnolia, Merbau, "Сосна", "Тик", "Ясень"},
	Plinthus:                     {"Бук", "Венге", "Вишня", Other, "Дуб", "Кедр", "Клён", RedWood, "Липа", Larch, Merbau, "Ольха", "Орех", "Осина", "Сосна", "Тик", "Ясень"},
	"Поддон":                      {Birch, Other, "Дуб", Larch, "Осина", "Сосна", "Тик"},
	"Полок":                       {Acacia, Amaranth, Bamboo, Birch, "Бук", "Венге", "Вишня", "Граб", Walnut, "Груша", "ДПК", Other, "Дуб", "Ель", Zebrawood, Karagach, Chestnut, "Кедр", "Клён", RedWood, "Липа", Larch, Magnolia, Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, "Пихта", Sandalwood, "Сосна", "Тик", Poplar, Cirico, "Ясень"},
	"Потолочный плинтус, галтель": {Other, "Кедр", "Липа", Larch, "Осина", "Сосна"},
	Raskladka:                   {Other, "Липа", "Ольха", "Сосна", "Тик"},
	"Рейка":                       {Amaranth, Balsa, Birch, "Бук", "Венге", "Граб", Grenadilla, Other, "Дуб", "Ель", Karagach, "Кедр", RedWood, "Липа", Larch, Magnolia, Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, PinkWood, "Сосна", "Тик", "Ясень"},
	"Слэб":                        {Acacia, Birch, "Бук", "Вишня", "Граб", Walnut, "Груша", Other, "Дуб", Karagach, Chestnut, "Кедр", "Клён", RedWood, "Липа", Larch, Merbau, "Ольха", "Орех", "Осина", "Сосна", Poplar, "Ясень"},
	"Столб для забора":            {Acacia, Amaranth, Bamboo, Birch, "Бук", "Венге", "Вишня", "Граб", Walnut, "Груша", "ДПК", Other, "Дуб", "Ель", Zebrawood, Karagach, Chestnut, "Кедр", "Клён", RedWood, "Липа", Larch, Magnolia, Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, "Пихта", Sandalwood, "Сосна", "Тик", Poplar, Cirico, "Ясень"},
	Ugolok:                      {"Бук", "ДПК", Other, "Дуб", "Кедр", "Липа", Larch, "Ольха", "Осина", "Сосна"},
	"Шкант":                       {Acacia, Amaranth, Bamboo, Birch, "Бук", "Венге", "Вишня", "Граб", Walnut, "Груша", "ДПК", Other, "Дуб", "Ель", Zebrawood, Karagach, Chestnut, "Кедр", "Клён", RedWood, "Липа", Larch, Magnolia, Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, "Пихта", Sandalwood, "Сосна", "Тик", Poplar, Cirico, "Ясень"},
	"Штапик":                      {"Абаш", Acacia, Amaranth, "Бакаут", Balsa, Bamboo, Birch, "Бук", "Венге", "Вишня", "Граб", Grenadilla, Walnut, "Груша", "ДПК", Other, "Дуб", "Ель", Zebrawood, "Ироко", Karagach, Chestnut, "Кедр", "Клён", RedWood, "Липа", Larch, Magnolia, "Меранти", Merbau, "Ольха", "Орех", "Осина", "Падук", Rosewood, "Пихта", PinkWood, Sandalwood, "Сосна", "Тик", Poplar, Cirico, "Чёрное дерево", "Ясень"},
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
