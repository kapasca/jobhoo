/* Chip input — converts a comma-separated hidden input into a visual tag UI.
   Usage: initChips('chip-box-id', 'hidden-input-id') */
(function () {
  function initChips(boxId, hiddenId) {
    var box    = document.getElementById(boxId);
    var hidden = document.getElementById(hiddenId);
    var txt    = box && box.querySelector('.chip-input__text');
    if (!box || !hidden || !txt) return;

    /* bootstrap chips from existing value, then sync back */
    var init = (hidden.value || '').split(',').map(function (s) { return s.trim(); }).filter(Boolean);
    init.forEach(add);
    sync();

    function add(val) {
      val = String(val || '').replace(/,/g, '').trim();
      if (!val) return;
      /* deduplicate case-insensitively */
      var lower = val.toLowerCase();
      if (Array.from(box.querySelectorAll('.chip-input__chip')).some(function (c) {
        return c.dataset.val.toLowerCase() === lower;
      })) return;
      var chip = document.createElement('span');
      chip.className = 'chip-input__chip';
      chip.dataset.val = val;
      chip.innerHTML =
        '<span>' + esc(val) + '</span>' +
        '<button type="button" aria-label="Remove ' + esc(val) + '">\u00d7</button>';
      chip.querySelector('button').addEventListener('click', function () {
        chip.remove(); sync();
      });
      box.insertBefore(chip, txt);
      sync();
    }

    function sync() {
      hidden.value = Array.from(box.querySelectorAll('.chip-input__chip'))
        .map(function (c) { return c.dataset.val; }).join(', ');
    }

    function esc(s) {
      return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
    }

    txt.addEventListener('keydown', function (e) {
      if (e.key === ',' || e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault();
        add(txt.value); txt.value = '';
      } else if (e.key === 'Backspace' && txt.value === '') {
        var chips = box.querySelectorAll('.chip-input__chip');
        if (chips.length) { chips[chips.length - 1].remove(); sync(); }
      }
    });

    txt.addEventListener('blur', function () {
      if (txt.value.trim()) { add(txt.value); txt.value = ''; }
    });

    /* clicking anywhere in the box focuses the text input */
    box.addEventListener('click', function () { txt.focus(); });
  }

  window.initChips = initChips;
})();
