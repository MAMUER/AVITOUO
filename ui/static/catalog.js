(function() {
  function inArray(v, arr) {
    return arr.indexOf(v) >= 0;
  }

  function getCheckedValues(containerId) {
    var container = document.getElementById(containerId);
    if (!container) return [];
    var inputs = container.querySelectorAll('input[type="checkbox"]:checked');
    return Array.from(inputs).map(function(el) { return el.value; });
  }

  function getSelectedLumberType() {
    var el = document.getElementById('lumber-types');
    return el ? el.value : '';
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
    select.innerHTML = '';
    var ph = document.createElement('option');
    ph.value = '';
    ph.textContent = '— Не выбрано —';
    select.appendChild(ph);
    options.forEach(function(opt) {
      var option = document.createElement('option');
      option.value = opt;
      option.textContent = opt;
      select.appendChild(option);
    });
    if (selectedValue && inArray(selectedValue, options)) {
      select.value = selectedValue;
    } else {
      select.value = '';
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

  function getSelectedWoodTypes() {
    return getCheckedValues('wood-types');
  }

  function syncDependentFields() {
    if (!AVITO_CATALOG) {
      console.warn('AVITO_CATALOG not loaded');
      return;
    }

    var lt = getSelectedLumberType();
    var woods = getSelectedWoodTypes();

    var validWoods = [];
    if (lt && AVITO_CATALOG.LT[lt]) {
      validWoods = AVITO_CATALOG.LT[lt].slice();
    } else {
      validWoods = AVITO_CATALOG.woods.slice();
    }
    var currentWoods = getCheckedValues('wood-types');
    renderCheckboxes('wood-types', validWoods, currentWoods);

    var showEdge = lt === 'Доска';
    var validEdges = showEdge ? AVITO_CATALOG.rules.edge.options.slice() : [];
    var currentEdges = getCheckedValues('edges');
    renderCheckboxes('edges', validEdges, currentEdges);
    setContainerVisibility('edges', 'edges-hint', showEdge, 'Доступно только для типа пиломатериала «Доска»');

    var showGrade = lt && inArray(lt, AVITO_CATALOG.rules.grade.lumberTypes) &&
                    woods.some(function(w) { return inArray(w, AVITO_CATALOG.rules.grade.woodTypes); });
    var validGrades = showGrade ? AVITO_CATALOG.rules.grade.options.slice() : [];
    var currentGrades = getCheckedValues('grades');
    renderCheckboxes('grades', validGrades, currentGrades);
    setContainerVisibility('grades', 'grades-hint', showGrade, 'Зависит от типа пиломатериала и вида древесины');

    var showMoisture = lt && inArray(lt, AVITO_CATALOG.rules.moisture.lumberTypes) &&
                       woods.some(function(w) { return inArray(w, AVITO_CATALOG.rules.moisture.woodTypes); });
    var validMoistures = showMoisture ? AVITO_CATALOG.rules.moisture.options.slice() : [];
    var currentMoistures = getCheckedValues('moistures');
    renderCheckboxes('moistures', validMoistures, currentMoistures);
    setContainerVisibility('moistures', 'moistures-hint', showMoisture, 'Зависит от типа пиломатериала и вида древесины');

    var showProfile = lt === 'Брус';
    var validProfiles = showProfile ? AVITO_CATALOG.rules.profile.options.slice() : [];
    var currentProfiles = getCheckedValues('profiles');
    renderCheckboxes('profiles', validProfiles, currentProfiles);
    setContainerVisibility('profiles', 'profiles-hint', showProfile, 'Доступно только для типа пиломатериала «Брус»');

    var showStructure = lt && inArray(lt, AVITO_CATALOG.rules.structure.lumberTypes);
    var validStructures = showStructure ? AVITO_CATALOG.rules.structure.options.slice() : [];
    var currentStructures = getCheckedValues('structures');
    renderCheckboxes('structures', validStructures, currentStructures);
    setContainerVisibility('structures', 'structures-hint', showStructure, 'Зависит от типа пиломатериала');

    var showLumberProfile = lt && AVITO_CATALOG.panelProfile[lt];
    var validLumberProfiles = showLumberProfile ? AVITO_CATALOG.panelProfile[lt].slice() : [];
    var currentLumberProfiles = getCheckedValues('lumber-profiles');
    renderCheckboxes('lumber-profiles', validLumberProfiles, currentLumberProfiles);
    setContainerVisibility('lumber-profiles', 'lumber-profiles-hint', showLumberProfile, 'Для Вагонка/Планкен');

    var showPriceUnits = lt && AVITO_CATALOG.rules.priceUnits.byLumberType[lt];
    var validPriceUnits = showPriceUnits ? AVITO_CATALOG.rules.priceUnits.byLumberType[lt].slice() : [];
    var currentPriceUnit = document.getElementById('price-units')?.value || '';
    setSelectOptions('price-units', validPriceUnits, currentPriceUnit);
    setContainerVisibility('price-units', 'price-units-hint', showPriceUnits, 'Зависит от типа пиломатериала');
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
    renderCheckboxes(containerId, options, []);
  }

  function init() {
    if (!AVITO_CATALOG) {
      console.warn('AVITO_CATALOG not loaded');
      return;
    }

    setSelectOptions('lumber-types', AVITO_CATALOG.lumberTypes.slice(), '');
    renderStaticGroup('wood-types', AVITO_CATALOG.woods.slice());

    document.querySelectorAll('.checkbox-group').forEach(function(group) {
      var optionsAttr = group.getAttribute('data-options');
      var catalogAttr = group.getAttribute('data-options-from-catalog');
      if (optionsAttr) {
        renderStaticGroup(group.id, optionsAttr);
      } else if (catalogAttr && catalogAttr !== 'lumberTypes' && catalogAttr !== 'woods') {
        var options = [];
        if (catalogAttr === 'edges') options = AVITO_CATALOG.rules.edge.options.slice();
        if (catalogAttr === 'grades') options = AVITO_CATALOG.rules.grade.options.slice();
        if (catalogAttr === 'moistures') options = AVITO_CATALOG.rules.moisture.options.slice();
        if (catalogAttr === 'profiles') options = AVITO_CATALOG.rules.profile.options.slice();
        if (catalogAttr === 'structures') options = AVITO_CATALOG.rules.structure.options.slice();
        if (catalogAttr === 'lumberProfiles') options = [];
        if (catalogAttr === 'priceUnits') options = AVITO_CATALOG.rules.priceUnits.options.slice();
        renderStaticGroup(group.id, options);
      }
    });

    document.getElementById('lumber-types').addEventListener('change', syncDependentFields);
    document.getElementById('wood-types').addEventListener('change', syncDependentFields);

    syncDependentFields();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
