let currentFile = null;
let currentData = [];
let currentHeaders = [];
let currentSheets = [];
let currentActiveSheet = '';
let totalAds = 0;

const debugLog = (message, isError = false) => {
    const logEl = document.getElementById('upload-debug');
    if (!logEl) return;
    logEl.classList.remove('hidden');
    const entry = document.createElement('div');
    entry.className = 'entry' + (isError ? ' error' : ' success');
    entry.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
    logEl.appendChild(entry);
    logEl.scrollTop = logEl.scrollHeight;
};

const setStatus = (text, ok = true) => {
    const el = document.getElementById('connection-status');
    if (!el) return;
    el.textContent = text;
    el.className = 'status-badge ' + (ok ? 'ok' : 'err');
};

const showMessage = (id, html) => {
    const el = document.getElementById(id);
    if (!el) return;
    el.innerHTML = html;
};

const escapeHtml = (text) => {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
};

const api = async (url, options = {}) => {
    const res = await fetch(url, options);
    const data = await res.json().catch(() => ({ error: 'Некорректный JSON от сервера' }));
    if (!res.ok || data.error) {
        const err = data.error || `HTTP ${res.status}`;
        console.error(`[API] ${url} ->`, err);
        debugLog(`${options.method || 'GET'} ${url} — ${err}`, true);
        throw new Error(err);
    }
    debugLog(`${options.method || 'GET'} ${url} — OK`);
    return data;
};

const updateStats = () => {
    const totalEl = document.getElementById('stat-total');
    const adsEl = document.getElementById('stat-ads');
    const catsEl = document.getElementById('stat-categories');
    const sheetEl = document.getElementById('stat-sheet-name');
    if (!totalEl || !adsEl || !catsEl || !sheetEl) return;
    const total = (currentData || []).length;
    const ads = (currentData || []).filter(r => r && r.some(v => v)).length;
    const cats = (currentSheets || []).length;
    totalEl.textContent = total;
    adsEl.textContent = totalAds ?? ads;
    catsEl.textContent = cats;
    sheetEl.textContent = currentActiveSheet ? `Лист: ${currentActiveSheet}` : '';
};

const updateCharCount = () => {
    const title = document.getElementById('base-title').value;
    const counter = document.getElementById('title-count');
    if (counter) counter.textContent = Math.max(0, 100 - title.length);
};

const updateDescCount = () => {
    const desc = document.getElementById('base-description').value;
    const counter = document.getElementById('desc-count');
    if (counter) counter.textContent = Math.max(0, 7500 - desc.length);
};

const priceUnitRules = {
    "Брус": ["Штуку", "м³"],
    "Брусок": ["Штуку", "м³"],
    "Доска": ["Штуку", "м³"],
    "Планкен": ["Штуку", "м²"],
    "Вагонка": ["Штуку", "м²"],
    "Дрова": ["Штуку", "м³"]
};
const priceUnitApplicable = new Set(["Брус", "Брусок", "Доска", "Планкен", "Вагонка", "Дрова"]);

const setupPriceUnitDependency = () => {
    const updateFor = (productTypeId, priceUnitId) => {
        const productTypeSelect = document.getElementById(productTypeId);
        const priceUnitSelect = document.getElementById(priceUnitId);
        if (!productTypeSelect || !priceUnitSelect) return null;

        const update = () => {
            const type = productTypeSelect.value;
            console.log('[price-unit] update', productTypeId, 'type=', type);
            priceUnitSelect.innerHTML = '';
            if (!type || !priceUnitApplicable.has(type)) {
                priceUnitSelect.disabled = true;
                priceUnitSelect.innerHTML = '<option value="">— Сначала выберите тип пиломатериала —</option>';
                return;
            }
            priceUnitSelect.disabled = false;
            const options = priceUnitRules[type] || priceUnitRules["__default__"];
            console.log('[price-unit] options for', type, '=', options);
            options.forEach(opt => {
                const option = document.createElement('option');
                option.value = opt;
                option.textContent = opt;
                priceUnitSelect.appendChild(option);
            });
        };

        productTypeSelect.addEventListener('change', update);
        update();
        return update;
    };

    const updateSettingsPrice = updateFor('product-type-settings', 'price-unit-settings');
    const updateGenPrice = updateFor('product-type', 'price-unit');

    const sync = (sourceId) => {
        const value = document.getElementById(sourceId)?.value;
        if (!value) return;
        if (sourceId === 'product-type' && document.getElementById('product-type-settings')?.value !== value) {
            document.getElementById('product-type-settings').value = value;
            updateSettingsPrice?.();
        } else if (sourceId === 'product-type-settings' && document.getElementById('product-type')?.value !== value) {
            document.getElementById('product-type').value = value;
            updateGenPrice?.();
        }
    };

    document.getElementById('product-type')?.addEventListener('change', () => sync('product-type'));
    document.getElementById('product-type-settings')?.addEventListener('change', () => sync('product-type-settings'));
};

const renderTable = () => {
    const thead = document.getElementById('table-header');
    const tbody = document.getElementById('table-body');
    if (!thead || !tbody) return;
    if (!currentHeaders.length || !currentData.length) {
        thead.innerHTML = '';
        tbody.innerHTML = '';
        return;
    }
    thead.innerHTML = '<tr><th>#</th>' + currentHeaders.map(h => `<th>${escapeHtml(h)}</th>`).join('</tr>');
    tbody.innerHTML = currentData.map((row, i) => '<tr>' +
        `<td>${i + 1}</td>` +
        currentHeaders.map((_, j) => `<td>${escapeHtml(row[j] || '')}</td>`).join('') +
        '</tr>').join('');
};

const populateSheetSelect = () => {
    const select = document.getElementById('sheet-select');
    if (!select) return;
    select.innerHTML = '';
    currentSheets.forEach(name => {
        const opt = document.createElement('option');
        opt.value = name;
        opt.textContent = name;
        if (name === currentActiveSheet) opt.selected = true;
        select.appendChild(opt);
    });
};

const loadSettings = async () => {
    try {
        const data = await api('/api/settings');
        const set = (id, value) => { const el = document.getElementById(id); if (el) el.value = value; };
        const check = (id, value) => { const el = document.getElementById(id); if (el) el.checked = value; };
        set('contacts', (data.contacts || []).join('\n'));
        set('phones', (data.phones || []).join('\n'));
        set('addresses', (data.addresses || []).join('\n'));
        set('companies', data.companies || '');
        set('emails', data.emails || '');
        check('disableAddress', data.disable_address_auto_fill || false);
        if (data.ad_type) set('adType', data.ad_type);
        if (data.condition) set('condition', data.condition);
        if (data.availability) set('availability', data.availability);
        if (data.product_type) set('product-type-settings', data.product_type);
        if (data.price_unit) set('price-unit-settings', data.price_unit);
        if (data.connect) set('connect', data.connect);
    } catch (e) {
        debugLog('Не удалось загрузить настройки: ' + e.message, true);
    }
};

const saveSettings = async () => {
    try {
        const contacts = document.getElementById('contacts').value.split('\n').filter(Boolean);
        const phones = document.getElementById('phones').value.split('\n').filter(Boolean);
        const addresses = document.getElementById('addresses').value.split('\n').filter(Boolean);
        const companies = document.getElementById('companies').value;
        const emails = document.getElementById('emails').value;
        const disableAddress = document.getElementById('disableAddress').checked;
        const adType = document.getElementById('adType').value;
        const condition = document.getElementById('condition').value;
        const availability = document.getElementById('availability').value;
        const productType = document.getElementById('product-type-settings').value;
        const priceUnit = document.getElementById('price-unit-settings').value;
        const connect = document.getElementById('connect').value;

        await api('/api/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ contacts, phones, addresses, companies, emails, disable_address_auto_fill: disableAddress, product_type: productType, ad_type: adType, condition, availability, price_unit: priceUnit, connect })
        });
        showMessage('settings-msg', '<div class="success">✅ Настройки сохранены</div>');
    } catch (e) {
        showMessage('settings-msg', `<div class="error">❌ ${escapeHtml(e.message)}</div>`);
    }
};

const processUploadResponse = (data, file) => {
    if (data.error) {
        showMessage('upload-msg', `<div class="error">❌ ${escapeHtml(data.error)}</div>`);
        return;
    }
    currentFile = data;
    currentSheets = data.sheets || [];
    currentActiveSheet = data.active_sheet || currentSheets[0] || '';
    currentData = data.rows || [];
    currentHeaders = data.headers || [];
    totalAds = data.total_ads ?? currentData.length;

    debugLog(`Загружен файл: ${file.name}, строк: ${currentData.length}, листов: ${currentSheets.length}`);

    if (currentSheets.length > 0) {
        document.getElementById('sheet-selector')?.classList.remove('hidden');
        populateSheetSelect();
    }
    document.getElementById('stats-block')?.classList.remove('hidden');
    document.getElementById('table-block')?.classList.remove('hidden');
    document.getElementById('generation-block')?.classList.remove('hidden');
    renderTable();
    updateStats();
    showMessage('upload-msg', `<div class="success">✅ Файл загружен: ${escapeHtml(file.name)} (${data.data_rows || currentData.length} строк)</div>`);
};

const uploadFile = async (source) => {
    const file = source?.target?.files?.[0] || source;
    if (!file) {
        showMessage('upload-msg', '<div class="error">❌ Файл не выбран</div>');
        console.warn('[upload] source=', source, 'file=', file);
        return;
    }

    const formData = new FormData();
    formData.append('file', file);

    const preview = {
        name: file.name,
        size: file.size,
        type: file.type || '<empty>',
        method: 'POST',
        url: '/api/upload',
        boundary: '<see network tab>',
        hasFile: !!formData.get('file'),
    };

    debugLog(`Попытка загрузки: ${file.name} (${file.size} байт, type=${file.type || 'empty'})`);
    try {
        const resp = await fetch('/api/upload', { method: 'POST', body: formData });
        const contentType = resp.headers.get('content-type') || '';
        const text = await resp.text();
        let data;
        try { data = JSON.parse(text); } catch { data = { raw: text }; }

        console.groupCollapsed('[upload] request/response');
        console.log('preview', preview);
        console.log('status', resp.status, 'ok=', resp.ok);
        console.log('content-type', contentType);
        console.log('body', data);
        console.groupEnd();

        debugLog(`${preview.method} ${preview.url} — ${resp.status} ${resp.ok ? 'OK' : 'FAIL'}`);
        if (!resp.ok || data.error) {
            const msg = data.error || `HTTP ${resp.status}`;
            showMessage('upload-msg', `<div class="error">❌ Загрузка не удалась: ${escapeHtml(msg)}</div>`);
            return;
        }
        processUploadResponse(data, file);
    } catch (e) {
        console.groupCollapsed('[upload] error');
        console.log('preview', preview);
        console.error(e);
        console.groupEnd();
        debugLog(`${preview.method} ${preview.url} — ${e.message}`, true);
        showMessage('upload-msg', `<div class="error">❌ Ошибка загрузки: ${escapeHtml(e.message)}</div>`);
    }
};

const loadSheet = async (sheetName) => {
    if (!currentFile) return;
    currentActiveSheet = sheetName;
    try {
        const data = await api(`/api/sheet?name=${encodeURIComponent(sheetName)}`);
        currentHeaders = data.headers || [];
        currentData = data.rows || [];
        debugLog(`Лист "${sheetName}" загружен, строк: ${currentData.length}`);
        renderTable();
        updateStats();
    } catch (e) {
        alert(e.message);
    }
};

const handleFolderSelect = (event) => {
    const files = event.target.files;
    if (!files || files.length === 0) return;

    const firstPath = files[0].name;
    const folderName = firstPath.split('/')[0].split('\\')[0];
    document.getElementById('photo-folder').value = folderName;
    document.getElementById('selected-folder-name').textContent = `Выбрано: ${folderName} (${files.length} файлов)`;

    const formData = new FormData();
    for (let i = 0; i < files.length; i++) {
        formData.append('files', files[i]);
    }
    formData.append('folder_name', folderName);

    fetch('/api/upload-folder', { method: 'POST', body: formData })
        .then(r => r.json())
        .then(data => {
            if (data.error) {
                alert(data.error);
            } else {
                debugLog(`Загружено ${data.uploaded} фото в ${data.full_path}`);
            }
        })
        .catch(err => {
            debugLog('Ошибка загрузки фото: ' + err.message, true);
            console.error('Upload error:', err);
        });
};

const generateAndExport = async () => {
    const baseTitle = document.getElementById('base-title').value;
    const baseDescription = document.getElementById('base-description').value;
    const photoFolder = document.getElementById('photo-folder').value;
    const variantCount = parseInt(document.getElementById('variant-count').value) || 10;
    const productType = document.getElementById('product-type')?.value || document.getElementById('product-type-settings')?.value || '';
    const priceUnit = document.getElementById('price-unit')?.value || document.getElementById('price-unit-settings')?.value || '';

    const msgEl = document.getElementById('generation-msg');
    msgEl.innerHTML = '<div class="success">⏳ Генерация... Пожалуйста, подождите</div>';

    try {
        const data = await api('/api/generate-and-export', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                base_title: baseTitle,
                base_description: baseDescription,
                photo_folder: photoFolder,
                variant_count: variantCount,
                product_type: productType,
                price_unit: priceUnit
            })
        });

        msgEl.innerHTML = `<div class="success">✅ Сгенерировано: ${data.generated} вариантов, фото: ${data.photo_count}</div>`;

        const downloadBlock = document.getElementById('download-block');
        const linksDiv = document.getElementById('download-links');

        linksDiv.innerHTML =
            '<p><a href="/api/download?file=' + encodeURIComponent(data.xlsx_file) + '" class="download-link" download>📄 Скачать Excel файл</a></p>' +
            '<p><a href="/api/download?file=' + encodeURIComponent(data.zip_file) + '" class="download-link" download>📦 Скачать ZIP архив с фото</a></p>';

        downloadBlock.classList.remove('hidden');
    } catch (e) {
        msgEl.innerHTML = `<div class="error">❌ ${escapeHtml(e.message)}</div>`;
    }
};

const init = async () => {
    try {
        await loadSettings();
        setStatus('Сервер OK', true);
    } catch (e) {
        setStatus('Нет связи', false);
        debugLog('Ошибка подключения к серверу: ' + e.message, true);
    }

    const uploadArea = document.getElementById('upload-area');
    if (uploadArea) {
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('drag-over');
        });

        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('drag-over');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('drag-over');
            const files = e.dataTransfer.files;
            if (!files || files.length === 0) return;
            const file = files[0];
            if (!file.name.match(/\.(xlsx|xls)$/i)) {
                alert('Пожалуйста, загрузите XLSX файл');
                return;
            }
            uploadFile(file);
        });
    }

    document.getElementById('base-title')?.addEventListener('input', updateCharCount);
    document.getElementById('base-description')?.addEventListener('input', updateDescCount);
    updateCharCount();
    updateDescCount();
    setupPriceUnitDependency();
};

init();
