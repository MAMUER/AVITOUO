package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"AVITOUO/core"
	"AVITOUO/storage"
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
		})
	case http.MethodPost:
		var req struct {
			Contacts               []string `json:"contacts"`
			Phones                 []string `json:"phones"`
			Addresses              []string `json:"addresses"`
			Companies              string   `json:"companies"`
			Emails                 string   `json:"emails"`
			DisableAddressAutoFill bool     `json:"disable_address_auto_fill"`
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

	fmt.Printf("[DEBUG] Column indices - Title: %d, Description: %d, ImageNames: %d\n", titleIdx, descIdx, imageNamesIdx)

	gen := core.NewTextGenerator()
	var newTitles, newDescriptions []string
	var err error
	if req.BaseTitle != "" || req.BaseDescription != "" {
		var baseTitle, baseDescription string
		if titleIdx >= 0 && len(dataCopy) > 0 {
			baseTitle = dataCopy[0][titleIdx]
		} else {
			baseTitle = req.BaseTitle
		}
		if descIdx >= 0 && len(dataCopy) > 0 {
			baseDescription = dataCopy[0][descIdx]
		} else {
			baseDescription = req.BaseDescription
		}

		if baseTitle == "" {
			baseTitle = req.BaseTitle
		}
		if baseDescription == "" {
			baseDescription = req.BaseDescription
		}

		newTitles, newDescriptions, err = gen.GenerateUniqueTexts(baseTitle, baseDescription, req.VariantCount)
		if err != nil {
			app.jsonError(w, http.StatusInternalServerError, "Ошибка генерации текстов: "+err.Error())
			return
		}
	}

	fmt.Printf("[DEBUG] Text generation done: titles=%d descriptions=%d photoFolder=%q\n", len(newTitles), len(newDescriptions), req.PhotoFolder)

	var photoNames []string
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
	}

	var imageNamesStrings []string
	if len(photoNames) > 0 {
		imageNamesStrings = photoNames
	}

	fmt.Printf("[DEBUG] Calling SaveExcelWithNewRows: sheet=%q titleIdx=%d descIdx=%d imageIdx=%d newTitles=%d newDescs=%d newImages=%d\n", activeSheetOriginal, titleIdx, descIdx, imageNamesIdx, len(newTitles), len(newDescriptions), len(imageNamesStrings))

	outputXLSX := "output_" + core.GenerateUniqueID() + ".xlsx"
	if err := storage.SaveExcelWithNewRows(path, outputXLSX, activeSheetOriginal, titleIdx, descIdx, imageNamesIdx, newTitles, newDescriptions, imageNamesStrings); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка сохранения Excel: "+err.Error())
		return
	}

	zipPath := "photos_" + core.GenerateUniqueID() + ".zip"
	files, err := storage.CreatePhotoZip(req.PhotoFolder, zipPath)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания ZIP: "+err.Error())
		return
	}

	if err := storage.CheckTotalSize(zipPath, outputXLSX); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Превышен лимит 100 МБ: "+err.Error())
		return
	}

	app.jsonResponse(w, map[string]interface{}{
		"status":      "ok",
		"xlsx_file":   outputXLSX,
		"zip_file":    zipPath,
		"generated":   req.VariantCount,
		"photo_count": len(photoNames),
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
