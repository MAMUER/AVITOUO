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

  function getSelectedLumberTypes() {
    return getCheckedValues('lumber-types');
  }

  function getSelectedWoodTypes() {
    return getCheckedValues('wood-types');
  }

  function syncDependentFields() {
    if (!AVITO_CATALOG) {
      console.warn('AVITO_CATALOG not loaded');
      return;
    }

    var lts = getSelectedLumberTypes();
    var woods = getSelectedWoodTypes();

    var validWoods = [];
    if (lts.length > 0) {
      var woodSet = {};
      lts.forEach(function(lt) {
        if (AVITO_CATALOG.LT[lt]) {
          AVITO_CATALOG.LT[lt].forEach(function(w) { woodSet[w] = true; });
        }
      });
      validWoods = Object.keys(woodSet).sort();
    } else {
      validWoods = AVITO_CATALOG.woods;
    }

    var currentWoods = getCheckedValues('wood-types');
    renderCheckboxes('wood-types', validWoods, currentWoods);

    var showEdge = lts.some(function(lt) { return lt === 'Доска'; });
    var validEdges = showEdge ? AVITO_CATALOG.rules.edge.options : [];
    var currentEdges = getCheckedValues('edges');
    renderCheckboxes('edges', validEdges, currentEdges);
    setContainerVisibility('edges', 'edges-hint', showEdge, 'Доступно только для типа пиломатериала «Доска»');

    var showGrade = lts.some(function(lt) { return inArray(lt, AVITO_CATALOG.rules.grade.lumberTypes); }) &&
                    woods.some(function(w) { return inArray(w, AVITO_CATALOG.rules.grade.woodTypes); });
    var validGrades = showGrade ? AVITO_CATALOG.rules.grade.options : [];
    var currentGrades = getCheckedValues('grades');
    renderCheckboxes('grades', validGrades, currentGrades);
    setContainerVisibility('grades', 'grades-hint', showGrade, 'Зависит от типа пиломатериала и вида древесины');

    var showMoisture = lts.some(function(lt) { return inArray(lt, AVITO_CATALOG.rules.moisture.lumberTypes); }) &&
                       woods.some(function(w) { return inArray(w, AVITO_CATALOG.rules.moisture.woodTypes); });
    var validMoistures = showMoisture ? AVITO_CATALOG.rules.moisture.options : [];
    var currentMoistures = getCheckedValues('moistures');
    renderCheckboxes('moistures', validMoistures, currentMoistures);
    setContainerVisibility('moistures', 'moistures-hint', showMoisture, 'Зависит от типа пиломатериала и вида древесины');

    var showProfile = lts.some(function(lt) { return lt === 'Брус'; });
    var validProfiles = showProfile ? AVITO_CATALOG.rules.profile.options : [];
    var currentProfiles = getCheckedValues('profiles');
    renderCheckboxes('profiles', validProfiles, currentProfiles);
    setContainerVisibility('profiles', 'profiles-hint', showProfile, 'Доступно только для типа пиломатериала «Брус»');

    var showStructure = lts.some(function(lt) { return inArray(lt, AVITO_CATALOG.rules.structure.lumberTypes); });
    var validStructures = showStructure ? AVITO_CATALOG.rules.structure.options : [];
    var currentStructures = getCheckedValues('structures');
    renderCheckboxes('structures', validStructures, currentStructures);
    setContainerVisibility('structures', 'structures-hint', showStructure, 'Зависит от типа пиломатериала');

    var showLumberProfile = lts.some(function(lt) { return AVITO_CATALOG.panelProfile[lt]; });
    var validLumberProfiles = [];
    if (showLumberProfile) {
      var profileSet = {};
      lts.forEach(function(lt) {
        if (AVITO_CATALOG.panelProfile[lt]) {
          AVITO_CATALOG.panelProfile[lt].forEach(function(p) { profileSet[p] = true; });
        }
      });
      validLumberProfiles = Object.keys(profileSet).sort();
    }
    var currentLumberProfiles = getCheckedValues('lumber-profiles');
    renderCheckboxes('lumber-profiles', validLumberProfiles, currentLumberProfiles);
    setContainerVisibility('lumber-profiles', 'lumber-profiles-hint', showLumberProfile, 'Для Вагонка/Планкен');
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

    renderStaticGroup('lumber-types', AVITO_CATALOG.lumberTypes);
    renderStaticGroup('wood-types', AVITO_CATALOG.woods);

    document.querySelectorAll('.checkbox-group').forEach(function(group) {
      var optionsAttr = group.getAttribute('data-options');
      var catalogAttr = group.getAttribute('data-options-from-catalog');
      if (optionsAttr) {
        renderStaticGroup(group.id, optionsAttr);
      } else if (catalogAttr && catalogAttr !== 'lumberTypes' && catalogAttr !== 'woods') {
        var options = [];
        if (catalogAttr === 'edges') options = AVITO_CATALOG.rules.edge.options;
        if (catalogAttr === 'grades') options = AVITO_CATALOG.rules.grade.options;
        if (catalogAttr === 'moistures') options = AVITO_CATALOG.rules.moisture.options;
        if (catalogAttr === 'profiles') options = AVITO_CATALOG.rules.profile.options;
        if (catalogAttr === 'structures') options = AVITO_CATALOG.rules.structure.options;
        if (catalogAttr === 'lumberProfiles') options = [];
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
