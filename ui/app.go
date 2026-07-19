package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	mux.HandleFunc("/api/validate", app.handleValidate)
	mux.HandleFunc("/api/generate-id", app.handleGenerateID)
	mux.HandleFunc("/api/generate-mass", app.handleGenerateMass)
	mux.HandleFunc("/api/shuffle", app.handleShuffle)
	mux.HandleFunc("/api/upload", app.handleUpload)
	mux.HandleFunc("/api/sheet", app.handleSheet)
	mux.HandleFunc("/api/save", app.handleSave)
	mux.HandleFunc("/api/export", app.handleExport)
	mux.HandleFunc("/api/columns", app.handleColumns)
	mux.HandleFunc("/api/uniquify-image", app.handleUniquifyImage)
	mux.HandleFunc("/api/photos", app.handlePhotosList)
	mux.HandleFunc("/api/photos/upload", app.handlePhotosUpload)
	mux.HandleFunc("/api/photos/delete", app.handlePhotosDelete)
	mux.HandleFunc("/api/photos/download", app.handlePhotosDownload)

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
		.tabs { display: flex; gap: 5px; margin-bottom: 20px; border-bottom: 2px solid #e0e0e0; }
		.tab-btn { padding: 12px 24px; border: none; background: #e0e0e0; cursor: pointer; border-radius: 4px 4px 0 0; font-size: 14px; }
		.tab-btn.active { background: #1976d2; color: white; }
		.tab { display: none; }
		.tab.active { display: block; }
		.card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
		.card h2 { color: #1976d2; margin-bottom: 15px; font-size: 18px; }
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
		.hidden { display: none; }
		.stats { display: flex; gap: 20px; margin-bottom: 15px; }
		.stat { background: #f5f5f5; padding: 10px 15px; border-radius: 4px; }
		.stat-label { font-size: 12px; color: #666; }
		.stat-value { font-size: 20px; font-weight: bold; color: #1976d2; }
		.photo-gallery { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 12px; }
		.photo-item { border: 1px solid #ddd; border-radius: 6px; padding: 8px; text-align: center; cursor: pointer; background: #fafafa; transition: box-shadow .2s; }
		.photo-item:hover { box-shadow: 0 2px 8px rgba(0,0,0,.15); }
		.photo-item img { width: 100%; height: 120px; object-fit: cover; border-radius: 4px; }
		.photo-item .photo-name { font-size: 12px; color: #333; margin-top: 6px; word-break: break-all; }
		.photo-item .photo-actions { margin-top: 6px; display: flex; gap: 6px; justify-content: center; }
		.photo-item button { padding: 4px 8px; font-size: 12px; }
	</style>
</head>
<body>
	<div class="header">
		<div class="container">
			<h1>Редактор шаблонов Авито</h1>
		</div>
	</div>

	<div class="container">
		<div class="tabs">
			<button class="tab-btn active" onclick="showTab('settings')">⚙️ Настройки</button>
			<button class="tab-btn" onclick="showTab('editor')">✏️ Редактор</button>
			<button class="tab-btn" onclick="showTab('export')">📦 Фото и Экспорт</button>
		</div>

		<div id="settings" class="tab active">
			<div class="card">
				<h2>Настройки по умолчанию</h2>
				<div class="form-group">
					<label>Контактные лица (каждый с новой строки):</label>
					<textarea id="contacts" rows="5"></textarea>
				</div>
				<div class="form-group">
					<label>Телефоны (каждый с новой строки):</label>
					<textarea id="phones" rows="3"></textarea>
				</div>
				<div class="form-group">
					<label>Адреса (каждый с новой строки):</label>
					<textarea id="addresses" rows="3"></textarea>
					<div class="checkbox-group">
						<input type="checkbox" id="disableAddress">
						<label for="disableAddress" style="margin:0">Отключить автозаполнение адреса</label>
					</div>
				</div>
				<div class="form-group">
					<label>Название компании:</label>
					<input type="text" id="companies" placeholder="Например: СтройДерево">
				</div>
				<div class="form-group">
					<label>Почта:</label>
					<input type="text" id="emails" placeholder="Например: info@example.com">
				</div>
				<button onclick="saveSettings()">💾 Сохранить настройки</button>
				<div id="settings-msg"></div>
			</div>
		</div>

		<div id="editor" class="tab active">
			<div class="card">
				<h2>Загрузка шаблона</h2>
				<div class="upload-area" onclick="document.getElementById('file-input').click()">
					<p>📁 Нажмите для выбора XLSX файла или перетащите сюда</p>
					<input type="file" id="file-input" accept=".xlsx,.xls" style="display:none" onchange="uploadFile(event)">
				</div>
				<div id="upload-msg"></div>
			</div>

			<div id="editor-content" class="hidden">
				<div class="card">
					<h2>Лист</h2>
					<div class="form-group">
						<select id="sheet-select" onchange="loadSheet(this.value)"></select>
					</div>
				</div>

				<div class="card">
					<h2>Статистика</h2>
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

				<div class="card">
					<h2>Создание объявления</h2>
					<div class="form-group">
						<label>Уникальный ID:</label>
						<div style="display:flex;gap:10px">
							<input type="text" id="ad-id" readonly>
							<button onclick="generateID()">🔄 Сгенерировать</button>
						</div>
					</div>
					<div class="form-group">
						<label>Название объявления (до 100 символов):</label>
						<input type="text" id="ad-title" maxlength="100" oninput="updateCharCount()">
						<div style="font-size:12px;color:#666;margin-top:5px">Осталось символов: <span id="title-count">100</span></div>
					</div>
					<div class="form-group">
						<label>Описание (до 7500 символов):</label>
						<textarea id="ad-description" rows="6" maxlength="7500" oninput="updateDescCount()"></textarea>
						<div style="font-size:12px;color:#666;margin-top:5px">Осталось символов: <span id="desc-count">7500</span></div>
					</div>
					<div class="form-group">
						<label>Шаблон для массовой генерации:</label>
						<textarea id="mass-template" rows="4" placeholder="Например: {Купить|Продать} {брус|доска} {в Мытищах|с доставкой}"></textarea>
						<div style="font-size:12px;color:#666;margin-top:5px">Используйте фигурные скобки с вариантами: {вариант1|вариант2}</div>
					</div>
					<div class="form-group">
						<label>Количество вариантов:</label>
						<input type="number" id="mass-count" value="10" min="1" max="1000" style="width:120px">
					</div>
					<button onclick="generateMassAds()">🚀 Сгенерировать варианты</button>
					<div id="mass-result"></div>
					<div id="editor-msg"></div>
				</div>
			</div>
		</div>

		<div id="export" class="tab">
			<div class="card">
				<h2>Фотографии</h2>
				<p>Папка для фото: <code>photos/[категория_or_ID]/</code></p>
				<div class="form-group">
					<label>Подпапка:</label>
					<input type="text" id="photo-dir" placeholder="Например: Для дома" value="">
				</div>
				<div class="upload-area" onclick="document.getElementById('photo-file-input').click()">
					<p>📷 Нажмите для загрузки фото или перетащите сюда</p>
					<input type="file" id="photo-file-input" accept="image/*" multiple style="display:none" onchange="uploadPhoto(event)">
				</div>
				<div id="photo-upload-msg"></div>
			</div>

			<div class="card">
				<h2>Галерея</h2>
				<div id="photo-gallery" class="photo-gallery">
					<p style="color:#666">Фото не загружены</p>
				</div>
			</div>

			<div id="photo-editor" class="card hidden">
				<h2>Редактирование фото</h2>
				<div style="display:flex;gap:20px;flex-wrap:wrap">
					<div style="flex:1;min-width:280px">
						<img id="editor-preview" style="max-width:100%;border:1px solid #ddd;border-radius:4px">
					</div>
					<div style="width:260px">
						<div class="form-group">
							<label>Яркость: <span id="brightness-val">100</span>%</label>
							<input type="range" id="brightness" min="0" max="200" value="100" oninput="updateFilterPreview()">
						</div>
						<div class="form-group">
							<label>Контраст: <span id="contrast-val">100</span>%</label>
							<input type="range" id="contrast" min="0" max="200" value="100" oninput="updateFilterPreview()">
						</div>
						<div class="form-group">
							<label>Насыщенность: <span id="saturate-val">100</span>%</label>
							<input type="range" id="saturate" min="0" max="200" value="100" oninput="updateFilterPreview()">
						</div>
						<div class="form-group">
							<label>Размытие: <span id="blur-val">0</span>px</label>
							<input type="range" id="blur" min="0" max="10" value="0" step="0.5" oninput="updateFilterPreview()">
						</div>
						<div class="form-group">
							<label>Оттенки серого:</label>
							<input type="range" id="grayscale" min="0" max="100" value="0" oninput="updateFilterPreview()">
						</div>
						<div class="form-group">
							<label>Сепия: <span id="sepia-val">0</span>%</label>
							<input type="range" id="sepia" min="0" max="100" value="0" oninput="updateFilterPreview()">
						</div>
						<div style="display:flex;gap:10px;margin-top:10px;flex-wrap:wrap">
							<button onclick="downloadEditedPhoto()">💾 Скачать</button>
							<button onclick="resetFilters()">🔄 Сбросить</button>
							<button onclick="deleteCurrentPhoto()" style="background:#d32f2f">🗑 Удалить</button>
						</div>
					</div>
				</div>
			</div>

			<div class="card">
				<h2>Экспорт</h2>
				<button onclick="createZip()">📦 Создать ZIP-архив из photos/</button>
				<div id="export-msg"></div>
			</div>
		</div>
	</div>

	<script>
		let currentFile = null;
		let currentData = [];
		let currentHeaders = [];
		let currentSheets = [];
		let currentActiveSheet = '';

		function showTab(id) {
			document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
			document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
			document.getElementById(id).classList.add('active');
			event.target.classList.add('active');
		}

		function updateCharCount() {
			const title = document.getElementById('ad-title').value;
			document.getElementById('title-count').textContent = 100 - title.length;
		}

		function updateDescCount() {
			const desc = document.getElementById('ad-description').value;
			document.getElementById('desc-count').textContent = 7500 - desc.length;
		}

		function updateStats() {
			document.getElementById('stat-total').textContent = currentData.length;
			document.getElementById('stat-ads').textContent = currentData.filter(r => r.some(v => v)).length;
			document.getElementById('stat-categories').textContent = currentSheets.length;
		}

		async function loadSettings() {
			const res = await fetch('/api/settings');
			const data = await res.json();
			if (data.error) { alert(data.error); return; }
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
				populateSheetSelect();
				document.getElementById('editor-content').classList.remove('hidden');
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
			renderTable();
			updateStats();
		}

		function renderTable() {
			const thead = document.getElementById('table-header');
			const tbody = document.getElementById('table-body');
			thead.innerHTML = '<th>#</th>' + currentHeaders.map(h => '<th>' + escapeHtml(h) + '</th>').join('');
			tbody.innerHTML = currentData.map((row, i) => '<tr>' +
				'<td>' + (i + 1) + '</td>' +
				currentHeaders.map((_, j) => '<td>' + escapeHtml(row[j] || '') + '</td>').join('') +
			'</tr>').join('');
		}

		async function generateID() {
			const res = await fetch('/api/generate-id');
			const data = await res.json();
			document.getElementById('ad-id').value = data.id;
		}

		async function generateMassAds() {
			const template = document.getElementById('mass-template').value;
			const count = parseInt(document.getElementById('mass-count').value) || 10;
			const res = await fetch('/api/generate-mass', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ template, count })
			});
			const data = await res.json();
			const msgEl = document.getElementById('mass-result');
			if (data.error) {
				msgEl.innerHTML = '<div class="error">❌ ' + data.error + '</div>';
			} else {
				msgEl.innerHTML = '<div class="success">✅ Сгенерировано: ' + data.generated + ' вариантов</div>';
			}
		}

		async function createZip() {
			const dir = document.getElementById('photo-dir').value;
			if (!dir) {
				document.getElementById('export-msg').innerHTML = '<div class="error">❌ Укажите путь к папке</div>';
				return;
			}
			const res = await fetch('/api/export?dir=' + encodeURIComponent(dir));
			const data = await res.json();
			document.getElementById('export-msg').innerHTML = data.error ?
				'<div class="error">❌ ' + data.error + '</div>' :
				'<div class="success">✅ ZIP создан. Файлы: ' + (data.files || '') + '</div>';
		}

		function escapeHtml(text) {
			const div = document.createElement('div');
			div.textContent = text;
			return div.innerHTML;
		}

		let currentPhotoDir = '';
		let currentPhotoName = '';

		async function uploadPhoto(event) {
			const files = event.target.files;
			if (!files || files.length === 0) return;

			const dir = document.getElementById('photo-dir').value.trim();
			if (!dir) {
				document.getElementById('photo-upload-msg').innerHTML = '<div class="error">❌ Укажите подпапку</div>';
				return;
			}

			currentPhotoDir = dir;
			const formData = new FormData();
			for (let i = 0; i < files.length; i++) {
				formData.append('file', files[i]);
			}
			formData.append('dir', dir);

			const res = await fetch('/api/photos/upload', { method: 'POST', body: formData });
			const data = await res.json();
			const msgEl = document.getElementById('photo-upload-msg');
			if (data.error) {
				msgEl.innerHTML = '<div class="error">❌ ' + data.error + '</div>';
			} else {
				msgEl.innerHTML = '<div class="success">✅ Загружено: ' + files.length + ' файл(ов)</div>';
				loadPhotoGallery();
			}
			event.target.value = '';
		}

		async function loadPhotoGallery() {
			const dir = document.getElementById('photo-dir').value.trim();
			if (!dir) return;

			const res = await fetch('/api/photos?dir=' + encodeURIComponent(dir));
			const data = await res.json();
			const gallery = document.getElementById('photo-gallery');

			if (data.error || !data.files || data.files.length === 0) {
				gallery.innerHTML = '<p style="color:#666">Фото не загружены</p>';
				return;
			}

			gallery.innerHTML = data.files.map(function(name) {
				return '<div class="photo-item" onclick="openPhotoEditor(\'' + escapeHtml(name).replace(/'/g, '&#39;') + '\')">' +
					'<img src="/api/photos/download?path=' + encodeURIComponent(dir + '/' + name) + '" loading="lazy">' +
					'<div class="photo-name">' + escapeHtml(name) + '</div>' +
					'<div class="photo-actions">' +
					'<button onclick="event.stopPropagation();downloadPhoto(\'' + escapeHtml(dir).replace(/'/g, '&#39;') + '\',\'' + escapeHtml(name).replace(/'/g, '&#39;') + '\')">&#9660;</button>' +
					'<button onclick="event.stopPropagation();deletePhoto(\'' + escapeHtml(dir).replace(/'/g, '&#39;') + '\',\'' + escapeHtml(name).replace(/'/g, '&#39;') + '\')" style="background:#d32f2f">&#128465;</button>' +
					'</div></div>';
			}).join('');
		}

		async function openPhotoEditor(name) {
			const dir = document.getElementById('photo-dir').value.trim();
			currentPhotoDir = dir;
			currentPhotoName = name;

			const preview = document.getElementById('editor-preview');
			preview.src = '/api/photos/download?path=' + encodeURIComponent(dir + '/' + name);
			document.getElementById('photo-editor').classList.remove('hidden');
			resetFilters();
		}

		function updateFilterPreview() {
			const brightness = document.getElementById('brightness').value;
			const contrast = document.getElementById('contrast').value;
			const saturate = document.getElementById('saturate').value;
			const blur = document.getElementById('blur').value;
			const grayscale = document.getElementById('grayscale').value;
			const sepia = document.getElementById('sepia').value;

			document.getElementById('brightness-val').textContent = brightness;
			document.getElementById('contrast-val').textContent = contrast;
			document.getElementById('saturate-val').textContent = saturate;
			document.getElementById('blur-val').textContent = blur;
			document.getElementById('sepia-val').textContent = sepia;

			const preview = document.getElementById('editor-preview');
			preview.style.filter = 'brightness(' + brightness + '%) contrast(' + contrast + '%) saturate(' + saturate + '%) blur(' + blur + 'px) grayscale(' + grayscale + '%) sepia(' + sepia + '%)';
		}

		function resetFilters() {
			document.getElementById('brightness').value = 100;
			document.getElementById('contrast').value = 100;
			document.getElementById('saturate').value = 100;
			document.getElementById('blur').value = 0;
			document.getElementById('grayscale').value = 0;
			document.getElementById('sepia').value = 0;
			updateFilterPreview();
		}

		async function downloadEditedPhoto() {
			if (!currentPhotoDir || !currentPhotoName) return;

			const brightness = document.getElementById('brightness').value;
			const contrast = document.getElementById('contrast').value;
			const saturate = document.getElementById('saturate').value;
			const blur = document.getElementById('blur').value;
			const grayscale = document.getElementById('grayscale').value;
			const sepia = document.getElementById('sepia').value;

			const res = await fetch('/api/photos/download?path=' + encodeURIComponent(currentPhotoDir + '/' + currentPhotoName) + '&filter=' + encodeURIComponent(JSON.stringify({brightness, contrast, saturate, blur, grayscale, sepia})));
			if (res.ok) {
				const blob = await res.blob();
				const url = URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = 'edited_' + currentPhotoName;
				a.click();
				URL.revokeObjectURL(url);
			}
		}

		async function deleteCurrentPhoto() {
			if (!currentPhotoDir || !currentPhotoName) return;
			if (!confirm('Удалить фото ' + currentPhotoName + '?')) return;

			const res = await fetch('/api/photos/delete?path=' + encodeURIComponent(currentPhotoDir + '/' + currentPhotoName), { method: 'DELETE' });
			const data = await res.json();
			if (data.error) {
				alert(data.error);
			} else {
				document.getElementById('photo-editor').classList.add('hidden');
				loadPhotoGallery();
			}
		}

		async function deletePhoto(dir, name) {
			if (!confirm('Удалить фото ' + name + '?')) return;

			const res = await fetch('/api/photos/delete?path=' + encodeURIComponent(dir + '/' + name), { method: 'DELETE' });
			const data = await res.json();
			if (data.error) {
				alert(data.error);
			} else {
				loadPhotoGallery();
			}
		}

		async function downloadPhoto(dir, name) {
			window.open('/api/photos/download?path=' + encodeURIComponent(dir + '/' + name), '_blank');
		}

		async function createZip() {
			const dir = document.getElementById('photo-dir').value.trim();
			if (!dir) {
				document.getElementById('export-msg').innerHTML = '<div class="error">❌ Укажите подпапку</div>';
				return;
			}

			const res = await fetch('/api/export?dir=' + encodeURIComponent(PhotosDir + '/' + dir));
			const data = await res.json();
			document.getElementById('export-msg').innerHTML = data.error ?
				'<div class="error">❌ ' + data.error + '</div>' :
				'<div class="success">✅ ZIP создан. Файлы: ' + (data.files || '') + '</div>';
		}

		document.getElementById('photo-dir').addEventListener('input', loadPhotoGallery);

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

func (app *App) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Phone       string `json:"phone"`
		Price       string `json:"price"`
	}
	if err := app.decodeJSON(r, &req); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Неверный JSON")
		return
	}

	if req.Title != "" {
		if err := core.ValidateTitle(req.Title); err != nil {
			app.jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if req.Phone != "" {
		if !core.ValidatePhone(req.Phone) {
			app.jsonError(w, http.StatusBadRequest, "Неверный формат телефона")
			return
		}
	}
	if req.Price != "" {
		if !core.ValidatePrice(req.Price) {
			app.jsonError(w, http.StatusBadRequest, "Цена должна быть целым числом")
			return
		}
	}
	if req.Description != "" {
		if err := core.ValidateDescription(req.Description); err != nil {
			app.jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	app.jsonResponse(w, map[string]string{"status": "ok"})
}

func (app *App) handleGenerateID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := core.GenerateUniqueID()
	app.jsonResponse(w, map[string]string{"id": id})
}

func (app *App) handleGenerateMass(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Template string `json:"template"`
		Count    int    `json:"count"`
	}
	if err := app.decodeJSON(r, &req); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Неверный JSON")
		return
	}
	if req.Count <= 0 {
		req.Count = 10
	}
	if req.Count > 1000 {
		req.Count = 1000
	}

	gen := core.NewTextGenerator()
	results, err := gen.GenerateVariations(req.Template, req.Count)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	app.jsonResponse(w, map[string]interface{}{
		"generated": len(results),
		"results":   results,
	})
}

func (app *App) handleShuffle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := app.decodeJSON(r, &req); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Неверный JSON")
		return
	}

	app.mu.Lock()
	app.usedTitles[req.Title] = true
	app.mu.Unlock()

	newTitle, err := core.ShuffleWords(req.Title, app.usedTitles)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	app.jsonResponse(w, map[string]string{"title": newTitle})
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
	fmt.Printf("Все листы файла: %v\n", sheets)
	if len(sheets) == 0 {
		app.jsonError(w, http.StatusBadRequest, "Файл не содержит листов")
		return
	}

	categorySheets := make([]string, 0, len(sheets))
	originalSheets := make([]string, 0, len(sheets))
	seen := make(map[string]bool)
	for _, s := range sheets {
		if strings.EqualFold(s, "Инструкция") {
			continue
		}
		if strings.HasPrefix(s, "Спр-") || strings.HasPrefix(s, "Спр") {
			continue
		}
		rows, _ := f.GetRows(s)
		if len(rows) <= 1 {
			continue
		}
		normalized := normalizeSheetName(s)
		if normalized == "" {
			continue
		}
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		categorySheets = append(categorySheets, normalized)
		originalSheets = append(originalSheets, s)
	}
	fmt.Printf("Категорийные листы: %v (уникальных: %d)\n", categorySheets, len(categorySheets))
	if len(categorySheets) == 0 {
		app.jsonError(w, http.StatusBadRequest, "В файле нет категорийных листов")
		return
	}

	activeSheet := originalSheets[0]
	activeSheetNormalized := categorySheets[0]
	rows, err := f.GetRows(activeSheet)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка чтения листа")
		return
	}

	if len(rows) == 0 {
		app.jsonError(w, http.StatusBadRequest, "Лист пустой")
		return
	}

	headers := rows[0]
	data := make([][]string, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := make([]string, len(headers))
		for j := 0; j < len(headers) && j < len(rows[i]); j++ {
			row[j] = rows[i][j]
		}
		data = append(data, row)
	}

	app.mu.Lock()
	for i := range categorySheets {
		app.sheetNameMap[categorySheets[i]] = originalSheets[i]
	}
	app.mu.Unlock()

	app.jsonResponse(w, map[string]interface{}{
		"headers":      headers,
		"rows":         data,
		"sheets":       categorySheets,
		"active_sheet": activeSheetNormalized,
	})
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

	rows, err := f.GetRows(originalName)
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка чтения листа")
		return
	}

	if len(rows) == 0 {
		app.jsonResponse(w, map[string]interface{}{"headers": []string{}, "rows": [][]string{}})
		return
	}

	headers := rows[0]
	data := make([][]string, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := make([]string, len(headers))
		for j := 0; j < len(headers) && j < len(rows[i]); j++ {
			row[j] = rows[i][j]
		}
		data = append(data, row)
	}

	app.jsonResponse(w, map[string]interface{}{
		"headers": headers,
		"rows":    data,
	})
}

func (app *App) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Headers []string   `json:"headers"`
		Rows    [][]string `json:"rows"`
		Path    string     `json:"path"`
	}
	if err := app.decodeJSON(r, &req); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Неверный JSON")
		return
	}

	if len(req.Rows) > 50000 {
		app.jsonError(w, http.StatusBadRequest, "Превышен лимит в 50 000 объявлений")
		return
	}

	f := excelize.NewFile()
	sheet := "Лист1"
	if err := f.SetSheetName(f.GetSheetName(f.GetActiveSheetIndex()), sheet); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания листа")
		return
	}

	for i, header := range req.Headers {
		if err := f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), header); err != nil {
			app.jsonError(w, http.StatusInternalServerError, "Ошибка записи заголовка")
			return
		}
	}

	for i, row := range req.Rows {
		for j, val := range row {
			if err := f.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+j, i+2), val); err != nil {
				app.jsonError(w, http.StatusInternalServerError, "Ошибка записи данных")
				return
			}
		}
	}

	path := req.Path
	if path == "" {
		path = "output.xlsx"
	}

	if err := f.SaveAs(path); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка сохранения: "+err.Error())
		return
	}

	app.jsonResponse(w, map[string]string{"status": "ok", "path": path})
}

func (app *App) handleExport(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		app.jsonError(w, http.StatusBadRequest, "Папка не указана")
		return
	}

	fullDir := filepath.Join(PhotosDir, filepath.Clean(dir))
	zipPath := "photos_" + core.GenerateUniqueID() + ".zip"
	fileNames, err := storage.CreatePhotoZip(fullDir, zipPath)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания ZIP: "+err.Error())
		return
	}

	if err := storage.CheckTotalSize(zipPath, "dummy.xlsx"); err != nil {
		app.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	app.jsonResponse(w, map[string]string{"status": "ok", "files": fileNames, "zip": zipPath})
}

func (app *App) handleColumns(w http.ResponseWriter, r *http.Request) {
	columns := map[string]interface{}{
		"placement_method": []string{"Package"},
		"contact_method":   []string{"По телефону и в сообщениям", "По телефону", "В сообщениях"},
		"ad_type":          []string{"Товар от производителя", "Товар приобретен на продажу"},
		"condition":        []string{"Новое", "Б/у"},
		"availability":     []string{"В наличии", "Под заказ"},
		"sales_type":       []string{"Товар куплен на продажу", "Товар произведён мной"},
		"ceiling_type":     []string{"Светильник", "Люстра"},
		"mounting_type":    []string{"Подвесное", "Потолочное"},
		"led":              []string{"Нет", "Да"},
		"picture_type":     []string{"Картины", "Рамки", "Панно", "Постеры и таблички", "Иконы"},
		"lighting_parts":   []string{"Плафоны и абажуры", "Лампочки", "Питание и управление светом"},
		"components_type":  []string{"Столбы и балясины", "Тетива", "Площадки", "Поручни"},
		"price_per":        []string{"Штуку", "м²", "м³", "Биг-бэг", "Мешок"},
		"lumber_type":      []string{"Брус", "Брусок", "Вагонка", "Горбыль", "Доска", "Дрова", "Другой", "Имитация бревна, блок-хаус", "Имитация бруса, рау-хаус", "Лес-кругляк", "Мебельный щит", "Наличник", "Настил", "Нащельник", "Оцилиндрованное бревно", "Планкен", "Плинтус", "Поддон", "Полок", "Потолочный плинтус, галтель", "Раскладка", "Рейка", "Слэб", "Столб для забора", "Шкант", "Штапик", "Уголок"},
		"wood_type":        []string{"Липа", "Лиственница", "Магнолия", "Меранти", "Мербау", "Ольха", "Орех", "Осина", "Падук", "Палисандр", "Пихта", "Розовое дерево", "Самшит", "Сосна", "Тик", "Тополь", "Цирикоте", "Чёрное дерево", "Ясень"},
		"wood_grade":       []string{"Отборный, экстра", "1 (A)", "1–2 (AB)", "1–3 (ABC)", "2 (B)", "2–3 (BC)", "3 (C)", "3–4 (CD)", "4 (D)"},
		"moisture":         []string{"Сухая", "Естественная"},
		"profiled":         []string{"", "Да", "Нет"},
		"gost":             []string{"", "Да", "Нет"},
		"connect_ads":      []string{"", "Да", "Нет"},
		"structure":        []string{"", "Цельная", "Клеёная"},
		"profile":          []string{"", "Евровагонка", "Прямой", "Скошенный", "Софтлайн", "Штиль"},
		"thickness":        []int{16, 18, 19, 20, 22, 23, 24, 25, 26, 27, 28, 30, 32, 34, 35, 36, 38, 40, 42, 44, 45, 50, 60, 75, 250},
		"width":            []int{10, 15, 20, 25, 30, 35, 40, 45, 50, 65, 70, 75, 80, 85, 90, 95, 100, 120, 125, 127, 130, 135, 140, 141, 142, 143, 145, 146, 150, 160, 170, 180, 190, 195, 200, 250, 300},
		"length":           []int{300, 600, 800, 900, 1000, 1100, 1200, 1400, 1500, 2000, 2500, 2700, 2900, 3000, 3500, 4000, 4500, 5000, 5500, 6000, 9000, 12000, 15000},
		"door_type":        []string{"Межкомнатные", "Входные", "Фурнитура", "Перегородки", "Другое"},
	}
	app.jsonResponse(w, columns)
}

func (app *App) handleUniquifyImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка парсинга формы")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Файл не найден")
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия файла: %v\n", closeErr)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка чтения файла")
		return
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Не удалось decode изображение: "+err.Error())
		return
	}

	uniqueImg := uniquifyImage(img)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, uniqueImg, &jpeg.Options{Quality: 90}); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка кодирования")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", `attachment; filename="unique.jpg"`)
	if _, err := w.Write(buf.Bytes()); err != nil {
		fmt.Printf("Ошибка записи изображения: %v\n", err)
	}
}

func (app *App) handlePhotosList(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		app.jsonError(w, http.StatusBadRequest, "Папка не указана")
		return
	}

	fullPath := filepath.Join(PhotosDir, filepath.Clean(dir))
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		app.jsonError(w, http.StatusNotFound, "Папка не найдена")
		return
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && isImage(entry.Name()) {
			files = append(files, entry.Name())
		}
	}

	app.jsonResponse(w, map[string]interface{}{"files": files, "dir": dir})
}

func (app *App) handlePhotosUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.jsonError(w, http.StatusBadRequest, "Ошибка парсинга формы")
		return
	}

	dir := r.FormValue("dir")
	if dir == "" {
		app.jsonError(w, http.StatusBadRequest, "Папка не указана")
		return
	}

	fullDir := filepath.Join(PhotosDir, filepath.Clean(dir))
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка создания папки")
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

	data, err := io.ReadAll(file)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка чтения файла")
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}

	savePath := filepath.Join(fullDir, header.Filename)
	counter := 1
	baseName := strings.TrimSuffix(header.Filename, ext)
	for {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			break
		}
		savePath = filepath.Join(fullDir, fmt.Sprintf("%s_%d%s", baseName, counter, ext))
		counter++
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка сохранения файла")
		return
	}

	app.jsonResponse(w, map[string]string{"status": "ok", "path": savePath, "name": filepath.Base(savePath)})
}

func (app *App) handlePhotosDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		app.jsonError(w, http.StatusBadRequest, "Путь не указан")
		return
	}

	fullPath := filepath.Join(PhotosDir, filepath.Clean(path))
	if err := os.Remove(fullPath); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка удаления файла")
		return
	}

	app.jsonResponse(w, map[string]string{"status": "ok"})
}

func (app *App) handlePhotosDownload(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		app.jsonError(w, http.StatusBadRequest, "Путь не указан")
		return
	}

	fullPath := filepath.Join(PhotosDir, filepath.Clean(path))
	file, err := os.Open(fullPath)
	if err != nil {
		app.jsonError(w, http.StatusNotFound, "Файл не найден")
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия файла: %v\n", closeErr)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка чтения файла")
		return
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		app.jsonError(w, http.StatusBadRequest, "Не удалось decode изображение: "+err.Error())
		return
	}

	filterStr := r.URL.Query().Get("filter")
	var brightness, contrast, saturate, blur, grayscale, sepia int
	brightness = 100
	contrast = 100
	saturate = 100
	grayscale = 0
	sepia = 0
	if filterStr != "" {
		var filters map[string]int
		if err := json.Unmarshal([]byte(filterStr), &filters); err == nil {
			if v, ok := filters["brightness"]; ok {
				brightness = v
			}
			if v, ok := filters["contrast"]; ok {
				contrast = v
			}
			if v, ok := filters["saturate"]; ok {
				saturate = v
			}
			if v, ok := filters["blur"]; ok {
				blur = v
			}
			if v, ok := filters["grayscale"]; ok {
				grayscale = v
			}
			if v, ok := filters["sepia"]; ok {
				sepia = v
			}
		}
	}

	if brightness != 100 || contrast != 100 || saturate != 100 || blur != 0 || grayscale != 0 || sepia != 0 {
		img = applyFilters(img, brightness, contrast, saturate, blur, grayscale, sepia)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		app.jsonError(w, http.StatusInternalServerError, "Ошибка кодирования")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(fullPath)+`"`)
	if _, err := w.Write(buf.Bytes()); err != nil {
		fmt.Printf("Ошибка записи изображения: %v\n", err)
	}
}

func isImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

func normalizeSheetName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "Спр-")
	name = strings.TrimPrefix(name, "Спр")
	name = strings.TrimPrefix(name, "_xlnm.")
	name = strings.TrimPrefix(name, "Print_Titles")
	return strings.TrimSpace(name)
}

func uniquifyImage(img image.Image) image.Image {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, image.Point{}, draw.Src)

	for i := 0; i < bounds.Dx(); i++ {
		for j := 0; j < bounds.Dy(); j++ {
			if i == 0 && j == 0 {
				r, g, b, a := dst.At(i, j).RGBA()
				r = (r + 1) & 0xFF
				g = (g + 1) & 0xFF
				b = (b + 1) & 0xFF
				dst.SetRGBA(i, j, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a >> 8)})
			}
		}
	}

	return dst
}

func applyFilters(img image.Image, brightness, contrast, saturate, blur, grayscale, sepia int) image.Image {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, image.Point{}, draw.Src)

	brightnessFactor := float64(brightness) / 100.0
	contrastFactor := float64(contrast) / 100.0
	grayscaleFactor := float64(grayscale) / 100.0
	sepiaFactor := float64(sepia) / 100.0

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, a := dst.At(x, y).RGBA()
			rf := float64(r >> 8)
			gf := float64(g >> 8)
			bf := float64(b >> 8)

			rf = clamp(rf * brightnessFactor)
			gf = clamp(gf * brightnessFactor)
			bf = clamp(bf * brightnessFactor)

			rf = clamp(((rf - 128) * contrastFactor) + 128)
			gf = clamp(((gf - 128) * contrastFactor) + 128)
			bf = clamp(((bf - 128) * contrastFactor) + 128)

			gray := 0.299*rf + 0.587*gf + 0.114*bf
			rf = rf*(1-grayscaleFactor) + gray*grayscaleFactor
			gf = gf*(1-grayscaleFactor) + gray*grayscaleFactor
			bf = bf*(1-grayscaleFactor) + gray*grayscaleFactor

			sR := 0.393*rf + 0.769*gf + 0.189*bf
			sG := 0.349*rf + 0.686*gf + 0.168*bf
			sB := 0.272*rf + 0.534*gf + 0.131*bf
			rf = rf*(1-sepiaFactor) + sR*sepiaFactor
			gf = gf*(1-sepiaFactor) + sG*sepiaFactor
			bf = bf*(1-sepiaFactor) + sB*sepiaFactor

			dst.SetRGBA(x, y, color.RGBA{R: uint8(clamp(rf)), G: uint8(clamp(gf)), B: uint8(clamp(bf)), A: uint8(a >> 8)})
		}
	}

	return dst
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
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
