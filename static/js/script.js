/* --- GLOBAL VARIABLES --- */
let maxZIndex = 1000;
let activeTerminals = {};
let globalDelId = null;
let globalExpectedName = "";

function bringToFront(el) {
    maxZIndex++;
    el.style.zIndex = maxZIndex;
}

/* --- GENERAL FUNCTIONALITY (MODALS AND ERRORS) --- */
function showErrorModal(message) {
    const el = document.getElementById('errorMessage');
    if (el) el.innerText = message;
    const modal = document.getElementById('errorModal');
    if (modal) modal.style.display = 'block';
}

window.closeErrorModal = function() {
    const modal = document.getElementById('errorModal');
    if (modal) modal.style.display = 'none';
};

window.openLogoutModal = function() { document.getElementById('logoutModal').style.display = 'block'; };
window.closeLogoutModal = function() { document.getElementById('logoutModal').style.display = 'none'; };
window.confirmLogout = function() { window.location.href = '/logout'; };

window.closeModal = function(id) {
    const m = document.getElementById(id) || document.querySelector('.modal[style*="block"]');
    if (m) m.style.display = 'none';
};

window.openPasswordModal = function() {
    const m = document.getElementById('passwordModal');
    if (m) m.style.display = 'block';
};

window.closePasswordModal = function() {
    const m = document.getElementById('passwordModal');
    if (m) m.style.display = 'none';
};

window.closeDeleteModal = function() {
    const m = document.getElementById('deleteModal');
    if (m) m.style.display = 'none';
};

/* ---  LOGIN AND PROFILE --- */
const loginForm = document.getElementById('loginForm');
if (loginForm) {
    loginForm.addEventListener('submit', function(e) {
        e.preventDefault();
        fetch('/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username: document.getElementById('username').value,
                password: document.getElementById('password').value,
                csrf_token: document.getElementById('csrf_token').value
            })
        }).then(r => r.json()).then(data => data.success ? window.location.href = '/' : showErrorModal(data.message));
    });
}

const usernameForm = document.getElementById('usernameForm');
if (usernameForm) {
    usernameForm.addEventListener('submit', function(e) {
        e.preventDefault();
        fetch('/profile/update-username', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username: document.getElementById('username').value,
                csrf_token: document.getElementById('csrf_token_username').value
            })
        }).then(r => r.json()).then(data => data.success ? location.reload() : showErrorModal(data.message));
    });
}

const passwordForm = document.getElementById('passwordForm');
if (passwordForm) {
    passwordForm.addEventListener('submit', function(e) {
        e.preventDefault();
        fetch('/profile/update-password', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                new_password: document.getElementById('newPassword').value,
                confirm_password: document.getElementById('confirmPassword').value,
                csrf_token: document.getElementById('csrf_token_password').value
            })
        }).then(r => r.json()).then(data => data.success ? location.reload() : showErrorModal(data.message));
    });
}

/* --- HOST AND KEY MANAGEMENT --- */
window.openAddModal = async function() {
    const hForm = document.getElementById('hostForm');
    const kForm = document.getElementById('keyForm');

    if (hForm) {
        await loadKeysForSelect();
        document.getElementById('modalTitle').innerText = 'Add Host';
        hForm.action = '/hosts/add';
        hForm.reset();
        document.getElementById('port').value = 22;
        document.getElementById('keyField').style.display = 'block';
        document.getElementById('hostModal').style.display = 'block';
    } else if (kForm) {
        document.getElementById('modalTitle').innerText = 'Add Key';
        if (document.getElementById('partialKeyHint')) document.getElementById('partialKeyHint').style.display = 'none';
        kForm.action = '/keys/add';
        kForm.reset();
        document.getElementById('keyModal').style.display = 'block';
    }
};

window.openEditModal = async function(id) {
    const hForm = document.getElementById('hostForm');
    const kForm = document.getElementById('keyForm');

    if (hForm) {
        await loadKeysForSelect();
        fetch(`/hosts/data/${id}`).then(r => r.json()).then(res => {
            if (res.success) {
                document.getElementById('modalTitle').innerText = 'Edit Host';
                hForm.action = `/hosts/edit/${id}`;
                document.getElementById('name').value = res.data.name;
                document.getElementById('address').value = res.data.address;
                document.getElementById('port').value = res.data.port;
                document.getElementById('username').value = res.data.username;
                document.getElementById('keyID').value = res.data.key_id || 0;
                document.getElementById('keyField').style.display = 'block';
                document.getElementById('hostModal').style.display = 'block';
            }
        });
    } else if (kForm) {
        fetch(`/keys/edit/${id}`).then(r => r.json()).then(res => {
            const key = res.data || res;
            document.getElementById('modalTitle').innerText = 'Edit Key';
            kForm.action = `/keys/edit/${id}`;
            document.getElementById('keyName').value = key.name || "";
            document.getElementById('privateKey').value = key.key_data || "";
            if (document.getElementById('partialKeyHint')) document.getElementById('partialKeyHint').style.display = 'block';
            document.getElementById('keyModal').style.display = 'block';
        });
    }
};

window.openDeleteModal = function(arg1, arg2) {
    const isHost = !!document.getElementById('deleteHostName');
    globalDelId = isHost ? arg1 : arg2;
    globalExpectedName = isHost ? arg2 : arg1;

    if (isHost) {
        document.getElementById('deleteHostName').value = '';
        document.getElementById('confirmDeleteBtn').disabled = true;
    } else {
        if (document.getElementById('targetKeyName')) document.getElementById('targetKeyName').innerText = globalExpectedName;
        document.getElementById('confirmName').value = '';
        document.getElementById('confirmDeleteButton').disabled = true;
    }
    document.getElementById('deleteModal').style.display = 'block';
};

window.confirmDelete = function() {
    const isHost = !!document.getElementById('deleteHostName');
    const url = isHost ? `/hosts/delete/${globalDelId}` : `/keys/delete/${globalDelId}`;
    fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ csrf_token: document.getElementById('csrf_token').value })
    }).then(r => r.json()).then(data => data.success ? location.reload() : showErrorModal(data.message));
};

/* --- SOME EVENTS AND VALIDATIONS --- */
document.addEventListener('input', (e) => {
    if (e.target.id === 'deleteHostName' || e.target.id === 'confirmName') {
        const btn = document.getElementById('confirmDeleteBtn') || document.getElementById('confirmDeleteButton');
        if (btn) btn.disabled = e.target.value !== globalExpectedName;
    }
});

document.addEventListener('submit', (e) => {
    if (e.target.id === 'hostForm' || e.target.id === 'keyForm') {
        e.preventDefault();
        const isHost = e.target.id === 'hostForm';

        if (isHost) {
            const kVal = document.getElementById('keyID').value;
            if (kVal === "0" || !kVal) {
                showErrorModal("Error: You must select a valid SSH key.");
                return;
            }
        } else {
            const kname = document.getElementById('keyName').value.trim();
            const kdata = document.getElementById('privateKey').value.trim();
            if (!kname || !kdata) {
                showErrorModal("Fields cannot be empty");
                return;
            }
        }

        const csrfToken = document.getElementById('csrf_token').value;
        let targetUrl = e.target.action;
        
        const payload = isHost ? {
            name: document.getElementById('name').value,
            address: document.getElementById('address').value,
            port: document.getElementById('port').value.toString(),
            username: document.getElementById('username').value,
            key_id: document.getElementById('keyID').value.toString(),
            csrf_token: csrfToken
        } : {
            name: document.getElementById('keyName').value.trim(),
            key_data: document.getElementById('privateKey').value.trim(),
            csrf_token: csrfToken
        };

        fetch(targetUrl, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        })
        .then(r => r.json())
        .then(data => data.success ? location.reload() : showErrorModal(data.message))
        .catch(err => showErrorModal(err.message));
    }
});

async function loadKeysForSelect() {
    const select = document.getElementById('keyID');
    if (!select) return;
    const r = await fetch('/keys/list');
    const res = await r.json();
    if (res.success) {
        select.innerHTML = '<option value="0">-- No selected Keys --</option>';
        res.data.forEach(k => {
            select.innerHTML += `<option value="${k.id}">${k.name}</option>`;
        });
    }
}

/* --- TERMINAL AND SSH --- */
window.connectToHost = function(id, name) {
    if (activeTerminals[id]) {
        window.maximizeTerminal(id);
        return;
    }

    const termWrapper = document.createElement('div');
    termWrapper.id = `term-win-${id}`;
    termWrapper.className = 'term-window';
    bringToFront(termWrapper);
    termWrapper.addEventListener('mousedown', () => bringToFront(termWrapper));

    const offset = Object.keys(activeTerminals).length * 30;
    termWrapper.style.left = (50 + offset) + 'px';
    termWrapper.style.top = (50 + offset) + 'px';
    termWrapper.style.width = '800px';
    termWrapper.style.height = '500px';

    termWrapper.innerHTML = `
        <div class="term-header" id="header-${id}">
            <span><i class="fas fa-terminal"></i> ${name}</span>
            <div class="term-controls">
                <button onclick="minimizeTerminal(${id})" title="Свернуть">_</button>
                <button onclick="maximizeTerminal(${id})" title="Развернуть/Сжать">&#9633;</button>
                <button onclick="window.closeTerminal(${id})" class="btn-close" title="Закрыть">&times;</button>
            </div>
        </div>
        <div id="term-${id}" class="term-body"></div>
        <div class="term-toolbar">            
            <div class="term-toolbar-group">
                <button class="term-btn" onclick="sendSpecialKey(${id}, 'Esc')">Esc</button>
                <button class="term-btn" onclick="sendSpecialKey(${id}, 'Tab')">Tab</button>
            </div>
            <div class="term-divider"></div>
            <div class="term-toolbar-group">
                <button class="term-btn term-btn-arrow" onclick="sendSpecialKey(${id}, 'Left')">&larr;</button>
                <button class="term-btn term-btn-arrow" onclick="sendSpecialKey(${id}, 'Up')">&uarr;</button>
                <button class="term-btn term-btn-arrow" onclick="sendSpecialKey(${id}, 'Down')">&uarr;</button>
                <button class="term-btn term-btn-arrow" onclick="sendSpecialKey(${id}, 'Right')">&rarr;</button>
            </div>
            <div class="term-divider"></div>
            <div class="term-toolbar-group">
                <button class="term-btn term-btn-danger" onclick="sendSpecialKey(${id}, 'Ctrl+C')">Ctrl+C</button>
            </div>
            <div class="term-divider"></div>
            <div class="term-toolbar-group">
                <button id="sel-btn-${id}" class="term-btn term-btn-action" onclick="toggleSelectMode(${id})">Select Off</button>
                <button class="term-btn term-btn-action" onclick="copyTerminalSelection(${id})">Copy</button>
                <button class="term-btn term-btn-action" onclick="pasteToTerminal(${id})">Paste</button>
            </div>
        </div>`;

    document.getElementById('terminal-container').appendChild(termWrapper);

    const term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Consolas, "Courier New", monospace',
        theme: { background: '#1e1e1e', foreground: '#ffffff' }
    });

    let fitAddon = new (window.FitAddon ? window.FitAddon.FitAddon : window.TerminalAddonFit.FitAddon)();
    term.loadAddon(fitAddon);
    term.open(document.getElementById(`term-${id}`));

    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${protocol}://${window.location.host}/ws/ssh?id=${id}`);

    activeTerminals[id] = { 
        term, ws, fitAddon, window: termWrapper,
        resizeObserver: new ResizeObserver(() => {
            window.requestAnimationFrame(() => {
                fitAddon.fit();
                if (ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
                }
            });
        })
    };
    activeTerminals[id].resizeObserver.observe(termWrapper);

    ws.onopen = () => {
        updateStatusBtn(id, true);
        term.onData(data => { if (ws.readyState === WebSocket.OPEN) ws.send(data); });
    };

    ws.onmessage = e => {
        if (e.data === "[STOPSESSION]") {
            updateStatusBtn(id, false); 
            setTimeout(() => window.closeTerminal(id), 800);
            return;
        }
        term.write(e.data);
    };

    ws.onclose = () => updateStatusBtn(id, false);

    makeDraggable(termWrapper, document.getElementById(`header-${id}`));
    makeResizable(termWrapper, id);

    setTimeout(() => { fitAddon.fit(); term.focus(); }, 200);
};

/* --- AUXILIARY FUNCTIONS (DRAG, RESIZE, STATUS) --- */
function makeDraggable(el, header) {
    let pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;
    header.onmousedown = function(e) {
        if (e.target.closest('button')) return;
        e.preventDefault();
        pos3 = e.clientX;
        pos4 = e.clientY;
        document.onmouseup = () => { document.onmouseup = null; document.onmousemove = null; };
        document.onmousemove = (ev) => {
            pos1 = pos3 - ev.clientX;
            pos2 = pos4 - ev.clientY;
            pos3 = ev.clientX;
            pos4 = ev.clientY;
            el.style.top = (el.offsetTop - pos2) + "px";
            el.style.left = (el.offsetLeft - pos1) + "px";
        };
    };
}

window.minimizeTerminal = function(id) {
    const win = document.getElementById(`term-win-${id}`);
    if (win) {
        win.style.opacity = '0';
        win.style.pointerEvents = 'none';
        win.style.transform = 'scale(0.9) translateY(1000px)'; 
    }
};

window.maximizeTerminal = function(id) {
    const data = activeTerminals[id];
    if (!data || !data.window) return;
    bringToFront(data.window);
    if (data.window.style.opacity === '0') {
        data.window.style.opacity = '1';
        data.window.style.pointerEvents = 'auto';
        data.window.style.transform = 'scale(1) translateY(0)';
    } else {
        data.window.classList.toggle('is-maximized');
    }
    setTimeout(() => { data.fitAddon.fit(); data.term.focus(); }, 150);
};

function updateStatusBtn(id, active) {
    const hostId = id.toString();
    const resumeBtn = document.getElementById(`resume-${hostId}`);
    const stdBtns = document.querySelectorAll(`.std-btn-${hostId}`);
    if (active) {
        stdBtns.forEach(btn => btn.style.display = 'none');
        if (resumeBtn) resumeBtn.style.display = 'inline-block';
    } else {
        stdBtns.forEach(btn => btn.style.display = 'inline-block');
        if (resumeBtn) resumeBtn.style.display = 'none';
    }
}

window.closeTerminal = async function(id) {
    const data = activeTerminals[id];
    if (!data) return;
    updateStatusBtn(id, false);
    const csrfToken = document.getElementById('csrf_token')?.value;
    try {
        await fetch(`/ssh/terminate?id=${id}`, { 
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ csrf_token: csrfToken })
        });
    } catch (e) {}
    if (data.ws) data.ws.close();
    if (data.window) data.window.remove();
    delete activeTerminals[id];
};

function makeResizable(el, id) {
    const resizer = document.createElement('div');
    resizer.className = 'resizer';
    el.appendChild(resizer);
    resizer.onmousedown = function(e) {
        e.preventDefault();
        const overlay = document.createElement('div');
        overlay.style = "position:fixed;top:0;left:0;width:100vw;height:100vh;z-index:99998;";
        document.body.appendChild(overlay);
        window.onmousemove = (ev) => {
            const w = ev.clientX - el.offsetLeft;
            const h = ev.clientY - el.offsetTop;
            if (w > 350) el.style.width = w + 'px';
            if (h > 200) el.style.height = h + 'px';
        };
        window.onmouseup = () => {
            window.onmousemove = null; window.onmouseup = null;
            document.body.removeChild(overlay);
            if (activeTerminals[id]) activeTerminals[id].fitAddon.fit();
        };
    };
}

/* --- INTERACTION WITH THE TERMINAL --- */
window.sendSpecialKey = function(id, key) {
    const data = activeTerminals[id];
    if (!data || !data.ws || data.ws.readyState !== WebSocket.OPEN) return;
    const keyMap = { 'Tab': '\t', 'Esc': '\x1b', 'Ctrl+C': '\x03', 'Up': '\x1b[A', 'Down': '\x1b[B', 'Left': '\x1b[D', 'Right': '\x1b[C' };
    if (keyMap[key]) { data.ws.send(keyMap[key]); data.term.focus(); }
};

window.copyTerminalSelection = function(id) {
    const data = activeTerminals[id];
    if (data && data.term.getSelection()) {
        navigator.clipboard.writeText(data.term.getSelection());
    }
};

window.pasteToTerminal = async function(id) {
    const data = activeTerminals[id];
    if (data && data.ws) {
        try {
            const text = await navigator.clipboard.readText();
            if (text && data.ws.readyState === WebSocket.OPEN) {
                data.ws.send(text);
                data.term.focus();
            }
        } catch (err) { showErrorModal("Clipboard access denied"); }
    }
};

window.toggleSelectMode = function(id) {
    const data = activeTerminals[id];
    if (!data) return;
    const termBody = data.window.querySelector('.term-body');
    let layer = termBody.querySelector('.selection-layer');
    const btn = document.getElementById(`sel-btn-${id}`);
    if (!layer) {
        layer = document.createElement('div');
        layer.className = 'selection-layer';
        let text = "";
        const buffer = data.term.buffer.active;
        for (let i = 0; i < buffer.length; i++) {
            const line = buffer.getLine(i);
            if (line) text += line.translateToString() + "\n";
        }
        layer.innerText = text;
        termBody.appendChild(layer);
        btn.innerText = 'Select ON';
        layer.scrollTop = layer.scrollHeight;
    } else {
        layer.remove();
        btn.innerText = 'Select Off';
        setTimeout(() => { data.term.focus(); data.term.scrollToBottom(); }, 50);
    }
};

if (window.visualViewport) {
    const adjustMobileTerminal = () => {
        // 1. Проверяем, что мы на мобилке
        if (window.innerWidth <= 768) {
            const activeWindows = document.querySelectorAll('.term-window');
            
            // 2. Срабатываем ТОЛЬКО если есть открытые терминалы
            if (activeWindows.length > 0) {
                const viewport = window.visualViewport;
                
                activeWindows.forEach(win => {
                    win.style.height = `${viewport.height}px`;
                    win.style.top = `${viewport.offsetTop}px`;
                });

                // 3. Скроллим в 0 только если фокус НЕ в обычном инпуте (логин/пароль)
                const activeEl = document.activeElement;
                const isInput = activeEl.tagName === 'INPUT' || activeEl.tagName === 'TEXTAREA';
                const isTermInput = activeEl.classList.contains('xterm-helper-textarea');

                if (!isInput || isTermInput) {
                    window.scrollTo(0, 0);
                }

                // Обновляем xterm
                Object.values(activeTerminals).forEach(data => {
                    if (data.fitAddon) data.fitAddon.fit();
                });
            }
        }
    };

    window.visualViewport.addEventListener('resize', adjustMobileTerminal);
    window.visualViewport.addEventListener('scroll', adjustMobileTerminal);
}