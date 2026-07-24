package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"AVITOUO/core"
	"AVITOUO/storage"
	"github.com/xuri/excelize/v2"
)

const PhotosDir = "photos"

type App struct {
	server         *http.Server
	port           string
	usedIDs        map[string]bool
	usedTitles     map[string]bool
	mu             sync.RWMutex
	uploadPath     string
	sheetNameMap   map[string]string
	activeSheet    string
	currentData    [][]string
	currentHeaders []string
}

func NewApp() *App {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	staticDir, _ := filepath.Abs("ui/static")

	mux := http.NewServeMux()
	app := &App{
		server:       &http.Server{Addr: ":" + port, Handler: mux},
		port:         port,
		usedIDs:      make(map[string]bool),
		usedTitles:   make(map[string]bool),
		sheetNameMap: make(map[string]string),
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/api/settings", app.handleSettings)
	mux.HandleFunc("/api/upload", app.handleUpload)
	mux.HandleFunc("/api/upload-folder", app.handleUploadFolder)
	mux.HandleFunc("/api/sheet", app.handleSheet)
	mux.HandleFunc("/api/generate-and-export", app.handleGenerateAndExport)
	mux.HandleFunc("/api/download", app.handleDownloadFile)

	return app
}

func (app *App) Run() {
	fmt.Printf("Сервер запущен: http://localhost:%s\n", app.port)
	fmt.Printf("Статические файлы: %s\n", filepath.Join("ui", "static"))
	if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := os.ReadFile("ui/static/index.html")
	if err != nil {
		http.Error(w, "index.html not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (app *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, _ := storage.LoadSettings()
		app.jsonResponse(w, map[string]interface{}{
			"contacts":                  settings.Contacts,
			"phones":                    settings.Phones,
			"addresses":                 settings.Addresses,
			"companies":                 settings.Companies,
			"emails":                    settings.Emails,
			"disable_address_auto_fill": settings.DisableAddressAutoFill,
			"placement":                 settings.Placement,
			"contact_method":            settings.ContactMethod,
			"ad_type":                   settings.AdType,
			"condition":                 settings.Condition,
			"availability":              settings.Availability,
			"sales_type":                settings.SalesType,
			"price_unit":                settings.PriceUnit,
			"connect":                   settings.Connect,
		})
	case http.MethodPost:
		var req struct {
			Contacts               []string `json:"contacts"`
			Phones                 []string `json:"phones"`
			Addresses              []string `json:"addresses"`
			Companies              string   `json:"companies"`
			Emails                 string   `json:"emails"`
			DisableAddressAutoFill bool     `json:"disable_address_auto_fill"`
			Placement              string   `json:"placement"`
			ContactMethod          string   `json:"contact_method"`
			AdType                 string   `json:"ad_type"`
			Condition              string   `json:"condition"`
			Availability           string   `json:"availability"`
			SalesType              string   `json:"sales_type"`
			PriceUnit              string   `json:"price_unit"`
			Connect                string   `json:"connect"`
		}
		if err := app.decodeJSON(r, &req); err != nil {
			app.jsonError(w, http.StatusBadRequest, "Неверный JSON")
			return
		}
		settings, _ := storage.LoadSettings()
		settings.Contacts = req.Contacts
		settings.Phones = req.Phones
		settings.Addresses = req.Addresses
		settings.Companies = strings.Split(req.Companies, "\n")
		settings.Emails = strings.Split(req.Emails, "\n")
		settings.DisableAddressAutoFill = req.DisableAddressAutoFill
		if req.Placement != "" {
			settings.Placement = req.Placement
		}
		if req.ContactMethod != "" {
			settings.ContactMethod = req.ContactMethod
		}
		if req.AdType != "" {
			settings.AdType = req.AdType
		}
		if req.Condition != "" {
			settings.Condition = req.Condition
		}
		if req.Availability != "" {
			settings.Availability = req.Availability
		}
		if req.SalesType != "" {
			settings.SalesType = req.SalesType
		}
		if req.PriceUnit != "" {
			settings.PriceUnit = req.PriceUnit
		}
		if req.Connect != "" {
			settings.Connect = req.Connect
		}
		if err := storage.SaveSettings(settings); err != nil {
			app.jsonError(w, http.StatusInternalServerError, "Ошибка сохранения")
			return
		}
		app.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func normalizeSheetName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "Спр-")
	name = strings.TrimPrefix(name, "Спр")
	name = strings.TrimPrefix(name, "_xlnm.")
	name = strings.TrimPrefix(name, "Print_Titles")
	return strings.TrimSpace(name)
}

func (app *App) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка парсинга формы: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Файл не найден: "+err.Error())
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия файла: %v\n", closeErr)
		}
	}()

	safeName := filepath.Base(header.Filename)
	for _, ch := range []string{":", "?", "*", "<", ">", "|", "\""} {
		safeName = strings.ReplaceAll(safeName, ch, "_")
	}
	tmpPath := filepath.Join(os.TempDir(), "upload_"+safeName)
	out, err := os.Create(tmpPath)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания временного файла: "+err.Error())
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия временного файла: %v\n", closeErr)
		}
		app.jsonError(w, http.StatusInternalServerError, "Ошибка сохранения файла: "+err.Error())
		return
	}
	if closeErr := out.Close(); closeErr != nil {
		fmt.Printf("Ошибка закрытия временного файла: %v\n", closeErr)
	}

	app.mu.Lock()
	if old := app.uploadPath; old != "" {
		_ = os.Remove(old)
	}
	app.uploadPath = tmpPath
	app.sheetNameMap = make(map[string]string)
	app.mu.Unlock()

	f, err := storage.LoadTemplate(tmpPath)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка открытия файла: "+err.Error())
		return
	}

	sheets := f.GetSheetList()
	fmt.Printf("[DEBUG] All sheets in file: %v\n", sheets)
	if len(sheets) == 0 {
		app.jsonError(w, http.StatusBadRequest, "Файл не содержит листов")
		return
	}

	// Ищем лист "Стройматериалы-Пиломатериалы" или похожий
	categorySheets := make([]string, 0, len(sheets))
	originalSheets := make([]string, 0, len(sheets))
	seen := make(map[string]bool)

	for _, s := range sheets {
		fmt.Printf("[DEBUG] Checking sheet: '%s'\n", s)
		if strings.EqualFold(s, "Инструкция") {
			fmt.Printf("[DEBUG] Skipping 'Инструкция' sheet\n")
			continue
		}
		if strings.HasPrefix(s, "Спр-") || strings.HasPrefix(s, "Спр") {
			fmt.Printf("[DEBUG] Skipping reference sheet: '%s'\n", s)
			continue
		}
		if strings.HasPrefix(s, "_xlnm.") || strings.HasPrefix(s, "Print_Titles") {
			fmt.Printf("[DEBUG] Skipping hidden sheet: '%s'\n", s)
			continue
		}

		rows, _ := f.GetRows(s)
		fmt.Printf("[DEBUG] Sheet '%s' has %d rows\n", s, len(rows))

		if len(rows) <= 1 {
			fmt.Printf("[DEBUG] Skipping sheet '%s' - only header or empty\n", s)
			continue
		}

		normalized := normalizeSheetName(s)
		if normalized == "" {
			fmt.Printf("[DEBUG] Skipping sheet '%s' - normalized to empty\n", s)
			continue
		}
		if seen[normalized] {
			fmt.Printf("[DEBUG] Skipping duplicate sheet: '%s'\n", s)
			continue
		}
		seen[normalized] = true

		categorySheets = append(categorySheets, normalized)
		originalSheets = append(originalSheets, s)
	}

	fmt.Printf("[DEBUG] Category sheets found: %v (unique: %d)\n", categorySheets, len(categorySheets))

	var activeSheetIdx int
	for i, s := range originalSheets {
		if strings.Contains(strings.ToLower(s), "пиломатериалы") || strings.Contains(strings.ToLower(s), "стройматериалы") {
			activeSheetIdx = i
			break
		}
	}

	totalRows := 0
	if len(originalSheets) > activeSheetIdx {
		allRows, _ := f.GetRows(originalSheets[activeSheetIdx])
		totalRows = len(allRows)
	}

	if len(categorySheets) == 0 {
		app.jsonError(w, http.StatusBadRequest, "В файле нет категорийных листов")
		return
	}

	activeSheet := originalSheets[activeSheetIdx]
	activeSheetNormalized := categorySheets[activeSheetIdx]
	fmt.Printf("[DEBUG] Active sheet selected: '%s' (original: '%s')\n", activeSheetNormalized, activeSheet)

	headers, data, err := storage.GetSheetData(f, activeSheet)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка чтения листа: "+err.Error())
		return
	}

	fmt.Printf("[DEBUG] Final stats - Headers: %d, Data rows: %d\n", len(headers), len(data))

	app.mu.Lock()
	for i := range categorySheets {
		app.sheetNameMap[categorySheets[i]] = originalSheets[i]
	}
	app.activeSheet = activeSheet
	app.mu.Unlock()

	app.jsonResponse(w, map[string]interface{}{
		"headers":      headers,
		"rows":         data,
		"sheets":       categorySheets,
		"active_sheet": activeSheetNormalized,
		"total_rows":   totalRows,
		"data_rows":    len(data),
		"categories":   len(categorySheets),
	})

	app.mu.Lock()
	app.currentHeaders = headers
	app.currentData = data
	app.mu.Unlock()
}

func (app *App) handleSheet(w http.ResponseWriter, r *http.Request) {
	sheetName := r.URL.Query().Get("name")
	if sheetName == "" {
		app.jsonError(w, http.StatusBadRequest, "Лист не указан")
		return
	}

	app.mu.RLock()
	path := app.uploadPath
	originalName, ok := app.sheetNameMap[sheetName]
	app.mu.RUnlock()

	if path == "" {
		app.jsonError(w, http.StatusBadRequest, "Файл не загружен")
		return
	}

	if !ok {
		originalName = sheetName
	}

	f, err := storage.LoadTemplate(path)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	headers, data, err := storage.GetSheetData(f, originalName)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка чтения листа: "+err.Error())
		return
	}

	app.mu.Lock()
	app.currentHeaders = headers
	app.currentData = data
	app.activeSheet = originalName
	app.mu.Unlock()

	fmt.Printf("[DEBUG] handleSheet response - headers: %d, rows: %d\n", len(headers), len(data))

	app.jsonResponse(w, map[string]interface{}{
		"headers": headers,
		"rows":    data,
	})
}

func (app *App) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		fmt.Printf("Ошибка сериализации JSON: %v\n", err)
	}
}

func (app *App) jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		fmt.Printf("Ошибка сериализации JSON: %v\n", err)
	}
}

func (app *App) decodeJSON(r *http.Request, v interface{}) error {
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия тела запроса: %v\n", closeErr)
		}
	}()
	return json.NewDecoder(r.Body).Decode(v)
}

func (app *App) handleGenerateAndExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BaseTitle       string `json:"base_title"`
		BaseDescription string `json:"base_description"`
		PhotoFolder     string `json:"photo_folder"`
		VariantCount    int    `json:"variant_count"`
	}
	if err := app.decodeJSON(r, &req); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Неверный JSON")
		return
	}

	if req.VariantCount <= 0 {
		req.VariantCount = 10
	}
	if req.VariantCount > 50000 {
		req.VariantCount = 50000
	}

	if req.BaseTitle == "" && req.BaseDescription == "" {
		app.jsonError(w, http.StatusBadRequest, "Укажите базовое название или описание")
		return
	}

	app.mu.RLock()
	path := app.uploadPath
	activeSheetOriginal := app.activeSheet
	dataCopy := make([][]string, len(app.currentData))
	copy(dataCopy, app.currentData)
	headersCopy := make([]string, len(app.currentHeaders))
	copy(headersCopy, app.currentHeaders)
	app.mu.RUnlock()

	if path == "" {
		app.jsonError(w, http.StatusBadRequest, "Excel файл не загружен")
		return
	}

	if activeSheetOriginal == "" {
		app.jsonError(w, http.StatusBadRequest, "Лист не выбран. Загрузите файл и выберите лист.")
		return
	}

	titleIdx := storage.FindColumnIndex(headersCopy, "Title")
	if titleIdx < 0 {
		titleIdx = storage.FindColumnIndex(headersCopy, "Название")
	}
	descIdx := storage.FindColumnIndex(headersCopy, "Description")
	if descIdx < 0 {
		descIdx = storage.FindColumnIndex(headersCopy, "Описание")
	}
	imageNamesIdx := storage.FindColumnIndex(headersCopy, "ImageNames")

	contactIdx := storage.FindColumnIndex(headersCopy, "Контактное лицо")
	phoneIdx := storage.FindColumnIndex(headersCopy, "Номер телефона")
	addressIdx := storage.FindColumnIndex(headersCopy, "Адрес")
	companyIdx := storage.FindColumnIndex(headersCopy, "Название компании")
	emailIdx := storage.FindColumnIndex(headersCopy, "Почта")

	settings, _ := storage.LoadSettings()
	settingsCount := req.VariantCount
	if settingsCount <= 0 {
		settingsCount = 10
	}
	newContacts := make([]string, settingsCount)
	newPhones := make([]string, settingsCount)
	newAddresses := make([]string, settingsCount)
	newCompanies := make([]string, settingsCount)
	newEmails := make([]string, settingsCount)
	newIDs := make([]string, settingsCount)
	newPlacements := make([]string, settingsCount)
	newContactMethods := make([]string, settingsCount)
	newCategories := make([]string, settingsCount)
	newProductTypes := make([]string, settingsCount)
	newSubProductTypes := make([]string, settingsCount)
	newPriceUnits := make([]string, settingsCount)
	newConditions := make([]string, settingsCount)
	newAvailabilities := make([]string, settingsCount)
	newAdTypes := make([]string, settingsCount)
	newSalesTypes := make([]string, settingsCount)
	newConnects := make([]string, settingsCount)
	newProcessing := make([]string, settingsCount)
	newPurpose := make([]string, settingsCount)

	categoryPath := ""
	if len(headersCopy) > 0 {
		categoryPath = headersCopy[0]
	}
	if !strings.Contains(categoryPath, " - ") && path != "" {
		if f, err := excelize.OpenFile(path); err == nil {
			if rows, err := f.GetRows(activeSheetOriginal); err == nil && len(rows) > 0 {
				firstCell := ""
				for _, v := range rows[0] {
					if v != "" {
						firstCell = v
						break
					}
				}
				if strings.Contains(firstCell, " - ") {
					categoryPath = firstCell
				}
			}
			_ = f.Close()
		}
	}
	categoryPart := ""
	productTypePart := ""
	subProductTypePart := ""
	if strings.Contains(categoryPath, " - ") {
		parts := strings.Split(categoryPath, " - ")
		if len(parts) >= 3 {
			categoryPart = strings.TrimSpace(parts[len(parts)-3])
			productTypePart = strings.TrimSpace(parts[len(parts)-2])
			subProductTypePart = strings.TrimSpace(parts[len(parts)-1])
		}
	}
	for i := range newCategories {
		newCategories[i] = categoryPart
	}
	for i := range newProductTypes {
		newProductTypes[i] = productTypePart
	}
	for i := range newSubProductTypes {
		newSubProductTypes[i] = subProductTypePart
	}

	lastID := 0
	if len(dataCopy) > 0 && len(dataCopy[len(dataCopy)-1]) > 0 {
		_, _ = fmt.Sscanf(dataCopy[len(dataCopy)-1][0], "%d", &lastID)
	}
	for i := range newIDs {
		newIDs[i] = strconv.Itoa(lastID + 1 + i)
	}

	for i := range newPlacements {
		newPlacements[i] = settings.Placement
	}

	contactMethodColIdx2 := -1
	if len(headersCopy) > 0 {
		contactMethodColIdx2 = storage.FindColumnIndex(headersCopy, "Способ связи")
	}
	contactDefault := settings.ContactMethod
	if len(dataCopy) > 0 && len(dataCopy[0]) > contactMethodColIdx2 && contactMethodColIdx2 >= 0 {
		contactDefault = dataCopy[0][contactMethodColIdx2]
	}
	for i := range newContactMethods {
		newContactMethods[i] = contactDefault
	}
	for i := range newPriceUnits {
		newPriceUnits[i] = settings.PriceUnit
	}
	for i := range newConditions {
		newConditions[i] = settings.Condition
	}
	for i := range newAvailabilities {
		newAvailabilities[i] = settings.Availability
	}
	for i := range newAdTypes {
		newAdTypes[i] = settings.AdType
	}
	for i := range newSalesTypes {
		newSalesTypes[i] = settings.SalesType
	}
	for i := range newConnects {
		newConnects[i] = settings.Connect
	}
	for i := range newProcessing {
		newProcessing[i] = "Строгание | Шлифование | Камерная сушка"
	}
	for i := range newPurpose {
		newPurpose[i] = "Баня | Дверь | Дом | Забор | Кровля | Лестница | Мебель | Окна | Опалубка | Поддоны | Пол | Полка | Потолок | Стена | Стропила | Терраса | Фасад"
	}

	if len(settings.Contacts) > 0 {
		for i := range newContacts {
			newContacts[i] = settings.Contacts[i%len(settings.Contacts)]
		}
	}
	if len(settings.Phones) > 0 {
		for i := range newPhones {
			newPhones[i] = settings.Phones[i%len(settings.Phones)]
		}
	}
	if len(settings.Addresses) > 0 && !settings.DisableAddressAutoFill {
		for i := range newAddresses {
			addr := settings.Addresses[i%len(settings.Addresses)]
			if strings.Contains(addr, "\n") {
				parts := strings.Split(addr, "\n")
				r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)))
				addr = parts[r.Intn(len(parts))]
			}
			newAddresses[i] = strings.TrimSpace(addr)
		}
	}
	if len(settings.Companies) > 0 {
		for i := range newCompanies {
			newCompanies[i] = settings.Companies[i%len(settings.Companies)]
		}
	}
	if len(settings.Emails) > 0 {
		for i := range newEmails {
			newEmails[i] = settings.Emails[i%len(settings.Emails)]
		}
	}

	fmt.Printf("[DEBUG] Column indices - Title: %d, Description: %d, ImageNames: %d, Contact: %d, Phone: %d, Address: %d, Company: %d, Email: %d\n", titleIdx, descIdx, imageNamesIdx, contactIdx, phoneIdx, addressIdx, companyIdx, emailIdx)

	gen := core.NewTextGenerator()
	var newTitles, newDescriptions []string
	var err error
	if req.BaseTitle != "" || req.BaseDescription != "" {
		baseTitle := strings.TrimSpace(req.BaseTitle)
		baseDescription := strings.TrimSpace(req.BaseDescription)

		newTitles = make([]string, req.VariantCount)
		newDescriptions = make([]string, req.VariantCount)
		for i := 0; i < req.VariantCount; i++ {
			title := baseTitle
			if i > 0 {
				title = fmt.Sprintf("%s #%d", baseTitle, i+1)
			}
			desc := baseDescription
			if i > 0 {
				desc = fmt.Sprintf("%s #%d", baseDescription, i+1)
			}
			if len(title) > 0 {
				title = strings.ToUpper(title[:1]) + title[1:]
			}
			newTitles[i] = title
			newDescriptions[i] = desc
		}
		_ = gen
		_ = err
	}

	fmt.Printf("[DEBUG] Text generation done: titles=%d descriptions=%d photoFolder=%q\n", len(newTitles), len(newDescriptions), req.PhotoFolder)

	var photoNames []string
	var imageNamesStrings []string
	if req.PhotoFolder != "" {
		avgPhotoSize := int64(500 * 1024)
		if err := storage.CheckSizeLimit(req.VariantCount, avgPhotoSize); err != nil {
			app.jsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		pg := &storage.PhotoGenerator{}
		photoNames, err = pg.GenerateUniquePhotos(req.PhotoFolder, req.VariantCount)
		if err != nil {
			app.jsonError(w, http.StatusInternalServerError, "Ошибка генерации фото: "+err.Error())
			return
		}
		imageNamesStrings = make([]string, len(photoNames))
		copy(imageNamesStrings, photoNames)
	}

	idColIdx := -1
	placementColIdx := -1
	contactMethodColIdx := -1
	categoryColIdx := -1
	productTypeColIdx := -1
	subProductTypeColIdx := -1
	priceUnitColIdx := -1
	conditionColIdx := -1
	availabilityColIdx := -1
	adTypeColIdx := -1
	salesTypeColIdx := -1
	connectColIdx := -1
	processingColIdx := -1
	purposeColIdx := -1

	if len(headersCopy) > 0 {
		idColIdx = storage.FindColumnIndex(headersCopy, "Уникальный идентификатор объявления")
		placementColIdx = storage.FindColumnIndex(headersCopy, "Способ размещения")
		contactMethodColIdx = storage.FindColumnIndex(headersCopy, "Способ связи")
		categoryColIdx = storage.FindColumnIndex(headersCopy, "Категория")
		productTypeColIdx = storage.FindColumnIndex(headersCopy, "Вид товара")
		subProductTypeColIdx = storage.FindColumnIndex(headersCopy, "Подвид товара")
		priceUnitColIdx = storage.FindColumnIndex(headersCopy, "Цена за")
		conditionColIdx = storage.FindColumnIndex(headersCopy, "Состояние")
		availabilityColIdx = storage.FindColumnIndex(headersCopy, "Доступность")
		adTypeColIdx = storage.FindColumnIndex(headersCopy, "Вид объявления")
		salesTypeColIdx = storage.FindColumnIndex(headersCopy, "Вид продажи")
		connectColIdx = storage.FindColumnIndex(headersCopy, "Соединять это объявление с другими объявлениями")
		processingColIdx = storage.FindColumnIndex(headersCopy, "Обработка")
		purposeColIdx = storage.FindColumnIndex(headersCopy, "Назначение")
	}

	fmt.Printf("[DEBUG] Additional column indices - id=%d placement=%d method=%d category=%d product=%d subProduct=%d priceUnit=%d condition=%d availability=%d adType=%d salesType=%d connect=%d processing=%d purpose=%d\n", idColIdx, placementColIdx, contactMethodColIdx, categoryColIdx, productTypeColIdx, subProductTypeColIdx, priceUnitColIdx, conditionColIdx, availabilityColIdx, adTypeColIdx, salesTypeColIdx, connectColIdx, processingColIdx, purposeColIdx)

	outputXLSX := "output_" + core.GenerateUniqueID() + ".xlsx"
	if err := storage.SaveExcelWithNewRows(path, outputXLSX, activeSheetOriginal, titleIdx, descIdx, imageNamesIdx, contactIdx, phoneIdx, addressIdx, companyIdx, emailIdx, newTitles, newDescriptions, imageNamesStrings, newContacts, newPhones, newAddresses, newCompanies, newEmails, idColIdx, placementColIdx, contactMethodColIdx, categoryColIdx, productTypeColIdx, subProductTypeColIdx, priceUnitColIdx, conditionColIdx, availabilityColIdx, adTypeColIdx, salesTypeColIdx, connectColIdx, processingColIdx, purposeColIdx, newIDs, newPlacements, newContactMethods, newCategories, newProductTypes, newSubProductTypes, newPriceUnits, newConditions, newAvailabilities, newAdTypes, newSalesTypes, newConnects, newProcessing, newPurpose); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка сохранения Excel: "+err.Error())
		return
	}

	zipPath := "photos_" + core.GenerateUniqueID() + ".zip"
	files, err := storage.CreatePhotoZip(req.PhotoFolder, zipPath)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания ZIP: "+err.Error())
		return
	}

	if req.PhotoFolder != "" {
		if err := storage.CheckTotalSize(zipPath, outputXLSX); err != nil {
			app.jsonError(w, http.StatusBadRequest, "Превышен лимит 100 МБ: "+err.Error())
			return
		}
	}

	app.jsonResponse(w, map[string]interface{}{
		"status":      "ok",
		"xlsx_file":   outputXLSX,
		"zip_file":    zipPath,
		"generated":   settingsCount,
		"photo_count": len(imageNamesStrings),
		"files":       files,
	})
}

func (app *App) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filename := r.URL.Query().Get("file")
	if filename == "" {
		app.jsonError(w, http.StatusBadRequest, "Файл не указан")
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		app.jsonError(w, http.StatusNotFound, "Файл не найден: "+err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	w.Header().Set("Content-Type", storage.GetMimeType(filename))
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	_, _ = io.Copy(w, file)
}

func (app *App) handleUploadFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка парсинга формы: "+err.Error())
		return
	}

	folderName := r.FormValue("folder_name")
	if folderName == "" {
		app.jsonError(w, http.StatusBadRequest, "Имя папки не указано")
		return
	}

	fullDir := filepath.Join(PhotosDir, filepath.Clean(folderName))
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания папки: "+err.Error())
		return
	}

	form := r.MultipartForm
	var uploaded int
	for _, fheaders := range form.File {
		for _, header := range fheaders {
			file, err := header.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(file)
			_ = file.Close()
			if err != nil {
				continue
			}
			savePath := filepath.Join(fullDir, filepath.Base(header.Filename))
			if err := os.WriteFile(savePath, data, 0644); err != nil {
				continue
			}
			uploaded++
		}
	}

	app.jsonResponse(w, map[string]interface{}{
		"status":    "ok",
		"folder":    folderName,
		"uploaded":  uploaded,
		"full_path": fullDir,
	})
}
