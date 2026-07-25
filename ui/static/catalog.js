(function() {
  function inArray(v, arr) {
    return arr.indexOf(v) >= 0;
  }

  function setOptions(selectEl, values, placeholder) {
    var current = selectEl.value;
    selectEl.innerHTML = '';
    if (placeholder) {
      var ph = document.createElement('option');
      ph.value = '';
      ph.textContent = placeholder || '— Не выбрано —';
      selectEl.appendChild(ph);
    }
    values.forEach(function(v) {
      var opt = document.createElement('option');
      opt.value = v;
      opt.textContent = v;
      selectEl.appendChild(opt);
    });
    if (inArray(current, values)) {
      selectEl.value = current;
    } else {
      selectEl.value = '';
    }
  }

  function disableSelect(id) {
    var el = document.getElementById(id);
    if (!el) return;
    el.disabled = true;
    el.value = '';
  }

  function enableSelect(id) {
    var el = document.getElementById(id);
    if (el) el.disabled = false;
  }

  function syncDependentFields() {
    if (!AVITO_CATALOG) {
      console.warn('AVITO_CATALOG not loaded');
      return;
    }

    var lt = document.getElementById('product-type-settings').value;
    var wood = document.getElementById('wood-type').value;

    if (lt && AVITO_CATALOG.LT[lt]) {
      setOptions(document.getElementById('wood-type'), AVITO_CATALOG.LT[lt]);
      enableSelect('wood-type');
    } else {
      setOptions(document.getElementById('wood-type'), AVITO_CATALOG.woods);
    }

    if (lt === AVITO_CATALOG.rules.edge.onlyLumberType) {
      enableSelect('edge');
      setOptions(document.getElementById('edge'), AVITO_CATALOG.rules.edge.options);
    } else {
      disableSelect('edge');
    }

    if (lt && inArray(lt, AVITO_CATALOG.rules.grade.lumberTypes)) {
      if (wood && inArray(wood, AVITO_CATALOG.rules.grade.woodTypes)) {
        enableSelect('grade');
        setOptions(document.getElementById('grade'), AVITO_CATALOG.rules.grade.options);
      } else {
        disableSelect('grade');
      }
    } else {
      disableSelect('grade');
    }

    if (lt && inArray(lt, AVITO_CATALOG.rules.moisture.lumberTypes)) {
      if (wood && inArray(wood, AVITO_CATALOG.rules.moisture.woodTypes)) {
        enableSelect('moisture');
        setOptions(document.getElementById('moisture'), AVITO_CATALOG.rules.moisture.options);
      } else {
        disableSelect('moisture');
      }
    } else {
      disableSelect('moisture');
    }

    if (lt === AVITO_CATALOG.rules.profile.onlyLumberType) {
      enableSelect('profile');
      setOptions(document.getElementById('profile'), AVITO_CATALOG.rules.profile.options);
    } else {
      disableSelect('profile');
    }

    if (lt && inArray(lt, AVITO_CATALOG.rules.structure.lumberTypes)) {
      enableSelect('structure');
      setOptions(document.getElementById('structure'), AVITO_CATALOG.rules.structure.options);
    } else {
      disableSelect('structure');
    }

    if (lt && AVITO_CATALOG.panelProfile[lt]) {
      enableSelect('lumber-profile');
      setOptions(document.getElementById('lumber-profile'), AVITO_CATALOG.panelProfile[lt]);
    } else {
      disableSelect('lumber-profile');
    }
  }

  function init() {
    if (!AVITO_CATALOG) {
      console.warn('AVITO_CATALOG not loaded');
      return;
    }

    setOptions(document.getElementById('wood-type'), AVITO_CATALOG.woods);

    Object.keys(AVITO_CATALOG.dimensions || {}).forEach(function(k) {
      setOptions(document.getElementById(k), AVITO_CATALOG.dimensions[k]);
    });

    document.getElementById('product-type-settings').addEventListener('change', syncDependentFields);
    document.getElementById('wood-type').addEventListener('change', syncDependentFields);

    Object.keys(AVITO_CATALOG.dimensions || {}).forEach(function(k) {
      enableSelect(k);
    });

    syncDependentFields();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
