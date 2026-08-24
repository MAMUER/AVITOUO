package core

// Settings хранит настройки по умолчанию для автозаполнения
type Settings struct {
	Contacts               []string `json:"contacts"`
	Phones                 []string `json:"phones"`
	Addresses              []string `json:"addresses"`
	Companies              []string `json:"companies"`
	Emails                 []string `json:"emails"`
	DisableAddressAutoFill bool     `json:"disable_address_auto_fill"`
	ProductType            string   `json:"product_type"`
	Placement              string   `json:"placement"`
	ContactMethod          string   `json:"contact_method"`
	AdType                 string   `json:"ad_type"`
	Condition              string   `json:"condition"`
	Availability           string   `json:"availability"`
	SalesType              string   `json:"sales_type"`
	PriceUnit              string   `json:"price_unit"`
	Connect                string   `json:"connect"`
}

// AdRow представляет строку объявления для редактирования
type AdRow struct {
	RowNum         int
	ID             string
	Title          string
	Description    string
	ContactPerson  string
	Phone          string
	Address        string
	Company        string
	Email          string
	Category       string
	ProductType    string
	SubProductType string
	Price          string
	PhotoNames     string // Формат: a.jpg|b.jpg|...
	// Дополнительные поля автозаполнения
	Condition     string // Новое, Б/у
	Availability  string // В наличии, Под заказ
	ContactMethod string // По телефону, В сообщениях
	AdType        string // Товар от производителя, Товар приобретен на продажу
	SalesType     string // Товар куплен на продажу
	Placement     string // Всегда "Package"
	// Поля для освещения
	CeilingType   string // Светильник, Люстра
	LED           string // Нет, Да
	LightingParts string // Тип комплектующих освещения
	// Поля для пиломатериалов
	LumberType string // Брус, Вагонка, Доска...
	WoodType   string // Липа, Сосна, Дуб...
	WoodGrade  string // 1 (A), 2 (B)...
	Moisture   string // Сухая, Естественная
	Thickness  string // мм
	Width      string // мм
	Length     string // мм
	// Поля для дверей
	DoorType     string // Межкомнатные, Входные
	DoorMaterial string // Материал
	DoorColor    string // Цвет
	// Жёстко заданные значения
	Processing string // Строгание | Шлифование | Камерная сушка
	Purpose    string // Баня | Дверь | Дом | Забор | Кровля | Лестница | Мебель | Окна | Опалубка | Поддоны | Пол | Полка | Потолок | Стена | Стропила | Терраса | Фасад
	// Пустые/игнорируемые столбцы
	AvitoAdNumber string
	IncludeVAT    string
	AvitoDateEnd  string
	AvitoStatus   string
	// Ссылки на фото
	PhotoURLs string
}
