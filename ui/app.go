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

	mux := http.NewServeMux()
	app := &App{
		server:       &http.Server{Addr: ":" + port, Handler: mux},
		port:         port,
		usedIDs:      make(map[string]bool),
		usedTitles:   make(map[string]bool),
		sheetNameMap: make(map[string]string),
	}

	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/api/settings", app.handleSettings)
	mux.HandleFunc("/api/upload", app.handleUpload)
	mux.HandleFunc("/api/sheet", app.handleSheet)
	mux.HandleFunc("/api/generate-and-export", app.handleGenerateAndExport)
	mux.HandleFunc("/api/download", app.handleDownloadFile)

	return app
}

func (app *App) Run() {
	fmt.Printf("Сервер запущен: http://localhost:%s\n", app.port)
	if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(htmlTemplate))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Редактор шаблонов Авито</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }
		body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 0; background: #f5f5f5; }
		.header { background: #1976d2; color: white; padding: 15px 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header h1 { font-size: 20px; font-weight: 500; }
		.container { max-width: 1400px; margin: 0 auto; padding: 20px; }
		.card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
		.card h2 { color: #1976d2; margin-bottom: 15px; font-size: 18px; cursor: pointer; }
		.form-group { margin-bottom: 15px; }
		label { display: block; margin-bottom: 5px; font-weight: 500; color: #333; }
		input, select, textarea { width: 100%; padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
		button { padding: 10px 20px; background: #1976d2; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; }
		button:hover { background: #1565c0; }
		.error { color: #d32f2f; margin-top: 10px; padding: 10px; background: #ffebee; border-radius: 4px; }
		.success { color: #388e3c; margin-top: 10px; padding: 10px; background: #e8f5e9; border-radius: 4px; }
		.upload-area { border: 2px dashed #1976d2; border-radius: 8px; padding: 40px; text-align: center; cursor: pointer; }
		.upload-area:hover { background: #e3f2fd; }
		table { width: 100%; border-collapse: collapse; margin-top: 15px; }
		th, td { padding: 8px; text-align: left; border: 1px solid #ddd; font-size: 13px; }
		th { background: #f5f5f5; font-weight: 600; position: sticky; top: 0; }
		.section { margin-bottom: 15px; padding: 15px; background: #fafafa; border-radius: 4px; border: 1px solid #eee; }
		.section h3 { color: #1976d2; font-size: 14px; margin-bottom: 10px; }
		.stats { display: flex; gap: 20px; margin-bottom: 15px; }
		.stat { background: #f5f5f5; padding: 10px 15px; border-radius: 4px; }
		.stat-label { font-size: 12px; color: #666; }
		.stat-value { font-size: 20px; font-weight: bold; color: #1976d2; }
		.hidden { display: none; }
		.accordion-content { display: none; }
		.accordion-content.active { display: block; }
	</style>
</head>
<body>
	<div class="header">
		<div class="container">
			<h1>🚀 Генератор объявлений Авито</h1>
		</div>
	</div>

	<div class="container">
		<!-- Блок 1: Загрузка Excel шаблона -->
		<div class="card">
			<h2>📁 1. Загрузка шаблона XLSX</h2>
			<div class="upload-area" onclick="document.getElementById('file-input').click()">
				<p>Нажмите для выбора XLSX файла или перетащите сюда</p>
				<input type="file" id="file-input" accept=".xlsx,.xls" style="display:none" onchange="uploadFile(event)">
			</div>
			<div id="upload-msg"></div>
			<div id="sheet-selector" class="hidden" style="margin-top: 15px;">
				<label>Выберите лист:</label>
				<select id="sheet-select" onchange="loadSheet(this.value)" style="margin-top: 5px;"></select>
			</div>
		</div>

		<!-- Блок 2: Статистика -->
		<div id="stats-block" class="card hidden">
			<h2>📊 2. Статистика</h2>
			<div class="stats">
				<div class="stat">
					<div class="stat-label">Всего строк</div>
					<div class="stat-value" id="stat-total">0</div>
				</div>
				<div class="stat">
					<div class="stat-label">Объявлений</div>
					<div class="stat-value" id="stat-ads">0</div>
				</div>
				<div class="stat">
					<div class="stat-label">Категорий</div>
					<div class="stat-value" id="stat-categories">0</div>
				</div>
			</div>
		</div>

		<!-- Блок 3: Таблица данных -->
		<div id="table-block" class="card hidden">
			<h2>📋 3. Данные листа</h2>
			<div style="max-height: 300px; overflow-y: auto;">
				<table>
					<thead id="table-header"></thead>
					<tbody id="table-body"></tbody>
				</table>
			</div>
		</div>

		<!-- Блок 4: Настройки по умолчанию -->
		<div class="card">
			<h2>⚙️ 4. Настройки по умолчанию</h2>
			<div class="section">
				<div class="form-group">
					<label>Контактные лица (каждый с новой строки):</label>
					<textarea id="contacts" rows="3" placeholder="Мариелена"></textarea>
				</div>
				<div class="form-group">
					<label>Телефоны (каждый с новой строки):</label>
					<textarea id="phones" rows="2" placeholder="79268509135"></textarea>
				</div>
				<div class="form-group">
					<label>Адреса (каждый с новой строки):</label>
					<textarea id="addresses" rows="2" placeholder="Мытищи, Волковское ш., 21А"></textarea>
					<input type="checkbox" id="disableAddress" style="margin-top: 5px;">
					<label for="disableAddress" style="display: inline; font-weight: normal; color: #666;">Отключить автозаполнение адреса</label>
				</div>
				<div class="form-group">
					<label>Название компании:</label>
					<input type="text" id="companies" placeholder="СтройДерево">
				</div>
				<div class="form-group">
					<label>Почта:</label>
					<input type="text" id="emails" placeholder="info@example.com">
				</div>
				<button onclick="saveSettings()">💾 Сохранить настройки</button>
			</div>
			<div id="settings-msg"></div>
		</div>

		<!-- Блок 5: Параметры генерации -->
		<div id="generation-block" class="card hidden">
			<h2>🔧 5. Параметры генерации</h2>
			
			<div class="section">
				<h3>Тексты для генерации</h3>
				<div class="form-group">
					<label>Базовое Название (до 100 символов):</label>
					<input type="text" id="base-title" maxlength="100">
					<div style="font-size:12px;color:#666;margin-top:5px">Осталось символов: <span id="title-count">100</span></div>
				</div>
				<div class="form-group">
					<label>Базовое Описание (до 7500 символов):</label>
					<textarea id="base-description" rows="4" maxlength="7500"></textarea>
					<div style="font-size:12px;color:#666;margin-top:5px">Осталось символов: <span id="desc-count">7500</span></div>
				</div>
			</div>

			<div class="section">
				<h3>Фотографии</h3>
				<div class="form-group">
					<label>Подпапка с фото-шаблонами:</label>
					<input type="text" id="photo-folder" placeholder="Например: Пиломатериалы" value="">
					<div style="font-size:12px;color:#666;margin-top:5px">Фото будут взяты из папки photos/подпапка/</div>
				</div>
				<div class="form-group">
					<label>Количество вариантов (N):</label>
					<input type="number" id="variant-count" value="10" min="1" max="50000" style="width:120px">
					<div style="font-size:12px;color:#666;margin-top:5px">Максимум 50000 за раз</div>
				</div>
			</div>

			<button onclick="generateAndExport()" style="background: #388e3c; font-size: 16px; padding: 12px 24px;">🚀 Сгенерировать и Экспортировать</button>
			<div id="generation-msg" style="margin-top: 15px;"></div>
		</div>

		<!-- Блок 6: Результат -->
		<div id="download-block" class="card hidden">
			<h2>📥 6. Результат</h2>
			<div id="download-links"></div>
		</div>
	</div>

	<script>
		let currentFile = null;
		let currentData = [];
		let currentHeaders = [];
		let currentSheets = [];
		let currentActiveSheet = '';

		function updateStats() {
			const totalEl = document.getElementById('stat-total');
			const adsEl = document.getElementById('stat-ads');
			const catsEl = document.getElementById('stat-categories');
			if (!totalEl || !adsEl || !catsEl) return;
			const total = (currentData || []).length;
			const ads = (currentData || []).filter(r => r && r.some(v => v)).length;
			const cats = (currentSheets || []).length;
			totalEl.textContent = total;
			adsEl.textContent = ads;
			catsEl.textContent = cats;
		}

		function updateCharCount() {
			const title = document.getElementById('base-title').value;
			document.getElementById('title-count').textContent = 100 - title.length;
		}

		function updateDescCount() {
			const desc = document.getElementById('base-description').value;
			document.getElementById('desc-count').textContent = 7500 - desc.length;
		}

		async function loadSettings() {
			const res = await fetch('/api/settings');
			const data = await res.json();
			if (data.error) return;
			document.getElementById('contacts').value = (data.contacts || []).join('\\n');
			document.getElementById('phones').value = (data.phones || []).join('\\n');
			document.getElementById('addresses').value = (data.addresses || []).join('\\n');
			document.getElementById('companies').value = data.companies || '';
			document.getElementById('emails').value = data.emails || '';
			document.getElementById('disableAddress').checked = data.disable_address_auto_fill || false;
		}

		async function saveSettings() {
			const contacts = document.getElementById('contacts').value.split('\\n').filter(Boolean);
			const phones = document.getElementById('phones').value.split('\\n').filter(Boolean);
			const addresses = document.getElementById('addresses').value.split('\\n').filter(Boolean);
			const companies = document.getElementById('companies').value;
			const emails = document.getElementById('emails').value;
			const disableAddress = document.getElementById('disableAddress').checked;

			const res = await fetch('/api/settings', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ contacts, phones, addresses, companies, emails, disable_address_auto_fill: disableAddress })
			});
			const data = await res.json();
			document.getElementById('settings-msg').innerHTML = data.error ?
				'<div class="error">' + data.error + '</div>' :
				'<div class="success">✅ Настройки сохранены</div>';
		}

		async function uploadFile(event) {
			const file = event.target.files[0];
			if (!file) return;

			const formData = new FormData();
			formData.append('file', file);

			const res = await fetch('/api/upload', { method: 'POST', body: formData });
			const data = await res.json();

			const msgEl = document.getElementById('upload-msg');
			if (data.error) {
				msgEl.innerHTML = '<div class="error">❌ ' + data.error + '</div>';
			} else {
				currentFile = data;
				currentSheets = data.sheets || [];
				currentActiveSheet = data.active_sheet || currentSheets[0] || '';
				currentData = data.rows || [];
				currentHeaders = data.headers || [];
				
				console.log('[DEBUG] Upload response:', data);
				console.log('[DEBUG] Current data rows:', currentData.length);
				
				if (currentSheets.length > 0) {
					document.getElementById('sheet-selector').classList.remove('hidden');
					populateSheetSelect();
				}
				document.getElementById('stats-block').classList.remove('hidden');
				document.getElementById('table-block').classList.remove('hidden');
				document.getElementById('generation-block').classList.remove('hidden');
				renderTable();
				updateStats();
				msgEl.innerHTML = '<div class="success">✅ Файл загружен: ' + file.name + '</div>';
			}
		}

		function populateSheetSelect() {
			const select = document.getElementById('sheet-select');
			select.innerHTML = '';
			currentSheets.forEach(name => {
				const opt = document.createElement('option');
				opt.value = name;
				opt.textContent = name;
				if (name === currentActiveSheet) opt.selected = true;
				select.appendChild(opt);
			});
		}

		async function loadSheet(sheetName) {
			if (!currentFile) return;
			currentActiveSheet = sheetName;
			const res = await fetch('/api/sheet?name=' + encodeURIComponent(sheetName));
			const data = await res.json();
			if (data.error) {
				alert(data.error);
				return;
			}
			currentHeaders = data.headers || [];
			currentData = data.rows || [];
			console.log('[DEBUG] Sheet loaded:', data);
			console.log('[DEBUG] Data rows:', currentData.length);
			renderTable();
			updateStats();
		}

		function renderTable() {
			const thead = document.getElementById('table-header');
			const tbody = document.getElementById('table-body');
			if (!thead || !tbody) return;
			thead.innerHTML = '<tr><th>#</th>' + currentHeaders.map(h => '<th>' + escapeHtml(h) + '</th>').join('</tr>');
			tbody.innerHTML = currentData.map((row, i) => '<tr>' +
				'<td>' + (i + 1) + '</td>' +
				currentHeaders.map((_, j) => '<td>' + escapeHtml(row[j] || '') + '</td>').join('') +
				'</tr>').join('');
		}

		async function generateAndExport() {
			const baseTitle = document.getElementById('base-title').value;
			const baseDescription = document.getElementById('base-description').value;
			const photoFolder = document.getElementById('photo-folder').value;
			const variantCount = parseInt(document.getElementById('variant-count').value) || 10;

			const msgEl = document.getElementById('generation-msg');
			msgEl.innerHTML = '<div class="success">⏳ Генерация... Пожалуйста, подождите</div>';

			const res = await fetch('/api/generate-and-export', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ 
					base_title: baseTitle, 
					base_description: baseDescription,
					photo_folder: photoFolder,
					variant_count: variantCount
				})
			});
			const data = await res.json();

			if (data.error) {
				msgEl.innerHTML = '<div class="error">❌ ' + data.error + '</div>';
			} else {
				msgEl.innerHTML = '<div class="success">✅ Сгенерировано: ' + data.generated + ' вариантов, фото: ' + data.photo_count + '</div>';
				
				const downloadBlock = document.getElementById('download-block');
				const linksDiv = document.getElementById('download-links');
				
				linksDiv.innerHTML = 
					'<p><a href="/api/download?file=' + encodeURIComponent(data.xlsx_file) + '" class="download-link" download>📄 Скачать Excel файл</a></p>' +
					'<p><a href="/api/download?file=' + encodeURIComponent(data.zip_file) + '" class="download-link" download>📦 Скачать ZIP архив с фото</a></p>';
				
				downloadBlock.classList.remove('hidden');
			}
		}

		function escapeHtml(text) {
			const div = document.createElement('div');
			div.textContent = text;
			return div.innerHTML;
		}

		loadSettings();
	</script>
</body>
</html>`

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
		app.jsonError(w, http.StatusBadRequest, "Ошибка парсинга формы")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Файл не найден")
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия файла: %v\n", closeErr)
		}
	}()

	tmpPath := filepath.Join(os.TempDir(), "upload_"+header.Filename)
	out, err := os.Create(tmpPath)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания временного файла")
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия временного файла: %v\n", closeErr)
		}
		app.jsonError(w, http.StatusInternalServerError, "Ошибка сохранения файла")
		return
	}
	if closeErr := out.Close(); closeErr != nil {
		fmt.Printf("Ошибка закрытия временного файла: %v\n", closeErr)
	}

	app.mu.Lock()
	if old := app.uploadPath; old != "" {
		if removeErr := os.Remove(old); removeErr != nil {
			fmt.Printf("Ошибка удаления старого временного файла: %v\n", removeErr)
		}
	}
	app.uploadPath = tmpPath
	app.sheetNameMap = make(map[string]string)
	app.mu.Unlock()

	f, err := storage.LoadTemplate(tmpPath)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, err.Error())
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
		// Пропускаем служебные листы
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

		// Проверяем, что лист имеет данные (более 1 строки = заголовок + данные)
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

	// Поиск конкретного листа "Стройматериалы-Пиломатериалы"
	var activeSheetIdx int
	for i, s := range originalSheets {
		if strings.Contains(strings.ToLower(s), "пиломатериалы") || strings.Contains(strings.ToLower(s), "стройматериалы") {
			activeSheetIdx = i
			break
		}
	}

	// Get total rows count from the active sheet
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

	// Сохраняем данные для дальнейшего использования в generate-and-export
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

	// Сохраняем данные для дальнейшего использования
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
		for i := 0; i < req.VariantCount; i++ {
			imageNamesStrings = append(imageNamesStrings, strings.Join(photoNames, " | "))
		}
	}

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
		app.jsonError(w, http.StatusNotFound, "Файл не найден")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	w.Header().Set("Content-Type", storage.GetMimeType(filename))
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	_, _ = io.Copy(w, file)
}
