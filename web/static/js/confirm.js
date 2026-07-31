/* Reusable JOBHOO-styled confirmation dialog.
   Usage: const { ok, value } = await jhConfirm({ title, message, confirmText, confirmClass, inputLabel, inputPlaceholder }) */
(function () {
  function esc(s) {
    return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
  }

  window.jhConfirm = function (opts) {
    return new Promise(function (resolve) {
      var o = opts || {};
      var hasInput = !!o.inputLabel;

      var overlay = document.createElement('div');
      overlay.className = 'jh-confirm-overlay';
      overlay.innerHTML =
        '<div class="jh-confirm-box">' +
          (o.title ? '<p class="jh-confirm-title">' + esc(o.title) + '</p>' : '') +
          '<p class="jh-confirm-msg">' + esc(o.message || '') + '</p>' +
          (hasInput
            ? '<div class="field mb-4"><label class="font-sm font-semibold text-white">' +
              esc(o.inputLabel) + '</label>' +
              '<input type="text" id="jh-ci" class="mt-2" placeholder="' +
              esc(o.inputPlaceholder || '') + '"></div>'
            : '') +
          '<div class="jh-confirm-actions">' +
            '<button type="button" class="btn btn--ghost btn--sm" data-r="0">' + esc(o.cancelText || 'Cancel') + '</button>' +
            '<button type="button" class="btn ' + esc(o.confirmClass || 'btn--primary') + ' btn--sm" data-r="1">' + esc(o.confirmText || 'Confirm') + '</button>' +
          '</div>' +
        '</div>';

      document.body.appendChild(overlay);
      if (hasInput) { var ci = overlay.querySelector('#jh-ci'); if (ci) ci.focus(); }

      function done(ok) {
        var ci = overlay.querySelector('#jh-ci');
        var val = (hasInput && ci) ? ci.value : '';
        overlay.remove();
        resolve({ ok: ok, value: val });
      }

      overlay.querySelector('[data-r="1"]').addEventListener('click', function () { done(true); });
      overlay.querySelector('[data-r="0"]').addEventListener('click', function () { done(false); });
      overlay.addEventListener('click', function (e) { if (e.target === overlay) done(false); });
      if (hasInput) {
        overlay.querySelector('#jh-ci').addEventListener('keydown', function (e) {
          if (e.key === 'Enter') { e.preventDefault(); done(true); }
          if (e.key === 'Escape') done(false);
        });
      }
    });
  };
})();
