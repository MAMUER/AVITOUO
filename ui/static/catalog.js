(function() {
  function inArray(v, arr) {
    return arr.includes(v);
  }

  function getCheckedValues(containerId) {
    var container = document.getElementById(containerId);
    if (!container) return [];
    var inputs = container.querySelectorAll('input[type="checkbox"]:checked');
    return Array.from(inputs).map(function(el) { return el.value; });
  }

  function getSelectedLumberType() {
    var el = document.getElementById('lumber-types');
    if (!el) return 'Доска';
    var v = el.value;
    if (!v) {
      var opts = el.options;
      if (opts && opts.length > 0) {
        v = opts[0].value;
      }
    }
    return v || 'Доска';
  }

  function getSelectedWoodTypes() {
    return getCheckedValues('wood-types');
  }

  function renderCheckboxes(containerId, options, selectedValues) {
    var container = document.getElementById(containerId);
    if (!container) return;
    container.innerHTML = '';
    options.forEach(function(opt) {
      var label = document.createElement('label');
      var input = document.createElement('input');
      input.type = 'checkbox';
      input.value = opt;
      if (inArray(opt, selectedValues)) {
        input.checked = true;
      }
      label.appendChild(input);
      label.appendChild(document.createTextNode(opt));
      container.appendChild(label);
    });
  }

  function setSelectOptions(selectId, options, selectedValue) {
    var select = document.getElementById(selectId);
    if (!select) return;
    var isMultiple = select.multiple === true;
    select.innerHTML = '';
    options.forEach(function(opt) {
      var option = document.createElement('option');
      option.value = opt;
      option.textContent = opt;
      select.appendChild(option);
    });
    if (Array.isArray(selectedValue)) {
      Array.from(select.options).forEach(function(opt) {
        opt.selected = selectedValue.includes(opt.value);
      });
    } else if (selectedValue && inArray(selectedValue, options)) {
      select.value = selectedValue;
    } else if (!isMultiple && options.length > 0) {
      select.value = options[0];
    }
  }

  function renderStaticGroup(containerId, options) {
    if (typeof options === 'string') {
      try {
        options = JSON.parse(options);
      } catch (e) {
        options = [];
      }
    }
    if (!Array.isArray(options)) {
      options = [];
    }
    var el = document.getElementById(containerId);
    if (!el) return;
    if (el.tagName === 'SELECT') {
      var current = [];
      if (el.multiple) {
        current = Array.from(el.selectedOptions).map(function(opt) { return opt.value; }).filter(Boolean);
      }
      setSelectOptions(containerId, options, current);
    } else {
      var currentChecked = getCheckedValues(containerId);
      if (options.length === 0 && AVITO_CATALOG?.dimensions?.[containerId]) {
        options = AVITO_CATALOG.dimensions[containerId].slice();
      }
      renderCheckboxes(containerId, options, currentChecked);
    }
  }

  function setContainerVisibility(containerId, hintId, visible, hintText) {
    var group = document.getElementById(containerId);
    if (!group) return;
    var parent = group.parentElement;
    if (!parent) return;
    if (visible) {
      parent.classList.remove('hidden');
    } else {
      parent.classList.add('hidden');
    }
    if (hintId) {
      var hint = document.getElementById(hintId);
      if (hint) hint.textContent = hintText || '';
    }
  }

  function getAvailability() {
    var el = document.getElementById('availability');
    return el ? el.value : '';
  }

  function syncDependentDimensions() {
    if (!AVITO_CATALOG) {
      return;
    }
    var lt = getSelectedLumberType();
    var availability = getAvailability();
    var dims = AVITO_CATALOG.dependent || {};
    ['thickness','width','length'].forEach(function(key) {
      var group = dims[key];
      var options = [];
      if (group?.[availability]?.[lt]) {
        options = group[availability][lt].slice();
      } else if (AVITO_CATALOG?.dimensions?.[key]) {
        options = AVITO_CATALOG.dimensions[key].slice();
      }
      var current = getCheckedValues(key);
      renderCheckboxes(key, options, current);
    });
  }

  function updateDimensionVisibility() {
    if (!AVITO_CATALOG) {
      return;
    }
    var lt = getSelectedLumberType();
    var visibility = AVITO_CATALOG.dimensionVisibility || {};
    Object.keys(visibility).forEach(function(fieldId) {
      var container = document.getElementById(fieldId);
      if (!container) return;
      var parent = container.parentElement;
      if (!parent) return;
      var allowed = visibility[fieldId] || [];
      var visible = allowed.includes(lt);
      if (visible) {
        parent.classList.remove('hidden');
      } else {
        parent.classList.add('hidden');
      }
    });
  }

  function renderCheckboxField(id, options, visible, hint) {
    var current = getCheckedValues(id);
    renderCheckboxes(id, options, current);
    setContainerVisibility(id, id + '-hint', visible, hint);
  }

  function syncDependentFields() {
    if (!AVITO_CATALOG) {
      console.warn('AVITO_CATALOG not loaded');
      return;
    }

    var lt = getSelectedLumberType();
    var woods = getSelectedWoodTypes();

    var validWoods = (lt && AVITO_CATALOG.LT[lt]) ? AVITO_CATALOG.LT[lt].slice() : AVITO_CATALOG.woods.slice();
    renderCheckboxField('wood-types', validWoods, true, '');

    var showEdge = lt === 'Доска';
    renderCheckboxField('edges', showEdge ? AVITO_CATALOG.rules.edge.options.slice() : [], showEdge, 'Доступно только для типа пиломатериала «Доска»');

    var showGrade = lt && inArray(lt, AVITO_CATALOG.rules.grade.lumberTypes) &&
                    woods.some(function(w) { return inArray(w, AVITO_CATALOG.rules.grade.woodTypes); });
    renderCheckboxField('grades', showGrade ? AVITO_CATALOG.rules.grade.options.slice() : [], showGrade, 'Зависит от типа пиломатериала и вида древесины');

    var showMoisture = lt && inArray(lt, AVITO_CATALOG.rules.moisture.lumberTypes) &&
                       woods.some(function(w) { return inArray(w, AVITO_CATALOG.rules.moisture.woodTypes); });
    renderCheckboxField('moistures', showMoisture ? AVITO_CATALOG.rules.moisture.options.slice() : [], showMoisture, 'Зависит от типа пиломатериала и вида древесины');

    var showProfile = lt === 'Брус';
    renderCheckboxField('profiles', showProfile ? AVITO_CATALOG.rules.profile.options.slice() : [], showProfile, 'Доступно только для типа пиломатериала «Брус»');

    var showStructure = lt && inArray(lt, AVITO_CATALOG.rules.structure.lumberTypes);
    renderCheckboxField('structures', showStructure ? AVITO_CATALOG.rules.structure.options.slice() : [], showStructure, 'Зависит от типа пиломатериала');

    var showLumberProfile = lt && AVITO_CATALOG.panelProfile[lt];
    renderCheckboxField('lumber-profiles', showLumberProfile ? AVITO_CATALOG.panelProfile[lt].slice() : [], showLumberProfile, 'Для Вагонка/Планкен');

    var showPriceUnits = lt && AVITO_CATALOG.rules.priceUnits.byLumberType[lt];
    var currentPriceUnit = document.getElementById('price-units')?.value || '';
    setSelectOptions('price-units', showPriceUnits ? AVITO_CATALOG.rules.priceUnits.byLumberType[lt].slice() : [], currentPriceUnit);
    setContainerVisibility('price-units', 'price-units-hint', showPriceUnits, 'Зависит от типа пиломатериала');

    syncDependentDimensions();
    updateDimensionVisibility();
  }

  function getCatalogOptions(catalogAttr) {
    var options = [];
    if (catalogAttr === 'edges') options = AVITO_CATALOG.rules.edge.options.slice();
    if (catalogAttr === 'grades') options = AVITO_CATALOG.rules.grade.options.slice();
    if (catalogAttr === 'moistures') options = AVITO_CATALOG.rules.moisture.options.slice();
    if (catalogAttr === 'profiles') options = AVITO_CATALOG.rules.profile.options.slice();
    if (catalogAttr === 'structures') options = AVITO_CATALOG.rules.structure.options.slice();
    if (catalogAttr === 'lumberProfiles') options = [];
    if (catalogAttr === 'priceUnits') options = AVITO_CATALOG.rules.priceUnits.options.slice();
    if (catalogAttr === 'thickness') options = (AVITO_CATALOG?.dimensions?.thickness) ? AVITO_CATALOG.dimensions.thickness.slice() : [];
    if (catalogAttr === 'width') options = (AVITO_CATALOG?.dimensions?.width) ? AVITO_CATALOG.dimensions.width.slice() : [];
    if (catalogAttr === 'length') options = (AVITO_CATALOG?.dimensions?.length) ? AVITO_CATALOG.dimensions.length.slice() : [];
    return options;
  }

  function initCheckboxGroups() {
    document.querySelectorAll('.checkbox-group').forEach(function(group) {
      var optionsAttr = group.dataset.options;
      var catalogAttr = group.dataset.optionsFromCatalog;
      if (optionsAttr) {
        renderStaticGroup(group.id, optionsAttr);
      } else if (catalogAttr && catalogAttr !== 'lumberTypes' && catalogAttr !== 'woods') {
        renderStaticGroup(group.id, getCatalogOptions(catalogAttr));
      }
    });
  }

  function initMultipleSelects() {
    document.querySelectorAll('select[multiple][data-options]').forEach(function(selectEl) {
      var optionsAttr = selectEl.dataset.options;
      if (optionsAttr) {
        renderStaticGroup(selectEl.id, optionsAttr);
      }
    });
  }

  function init() {
    if (!AVITO_CATALOG) {
      console.warn('AVITO_CATALOG not loaded');
      return;
    }

    var defaultLumberType = (AVITO_CATALOG.lumberTypes.includes('Доска')) ? 'Доска' : (AVITO_CATALOG.lumberTypes[0] || '');
    if (!defaultLumberType && AVITO_CATALOG.lumberTypes.length > 0) {
      defaultLumberType = AVITO_CATALOG.lumberTypes[0];
    }
    setSelectOptions('lumber-types', AVITO_CATALOG.lumberTypes.slice(), defaultLumberType);
    setSelectOptions('price-units', [], '');
    renderStaticGroup('wood-types', AVITO_CATALOG.woods.slice());

    initCheckboxGroups();
    initMultipleSelects();

    document.getElementById('lumber-types').addEventListener('change', syncDependentFields);
    document.getElementById('wood-types').addEventListener('change', syncDependentFields);
    document.getElementById('availability').addEventListener('change', syncDependentFields);

    syncDependentFields();
    updateDimensionVisibility();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
