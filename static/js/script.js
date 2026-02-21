/* --- GLOBAL VARIABLES --- */
let maxZIndex = 1000;
let activeTerminals = {};
let globalDelId = null;
let globalExpectedName = "";

function bringToFront(el) {
    maxZIndex++;
    el.style.zIndex = maxZIndex;

    const id = el.id.replace('term-win-', '');
    if (activeTerminals[id]) {
        activeTerminals[id].fitAddon.fit();
    }
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
        document.getElementById('authType').value = 'key';
        toggleAuthFields();
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
                const settings = res.data.settings || {};
                document.getElementById('defaultPath').value = settings.default_path || "/";
                document.getElementById('port').value = res.data.port;
                document.getElementById('username').value = res.data.username;
                const aType = res.data.auth_type || 'key';
                document.getElementById('authType').value = aType;

                if (aType === 'password') {
                    const passInput = document.getElementById('hostPassword');
                    passInput.value = ''; 
                    passInput.placeholder = '******** (leave empty to keep current)';
                } else {
                    document.getElementById('keyID').value = res.data.key_id || 0;
                }

                toggleAuthFields();
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
        if (document.getElementById('targetHostName')) document.getElementById('targetHostName').innerText = globalExpectedName;
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

window.toggleAuthFields = function() {
    const type = document.getElementById('authType').value;
    const keyField = document.getElementById('keyField');
    const passField = document.getElementById('passwordField');

    if (type === 'key') {
        keyField.style.display = 'block';
        passField.style.display = 'none';
        document.getElementById('hostPassword').value = '';
    } else {
        keyField.style.display = 'none';
        passField.style.display = 'block';
        document.getElementById('keyID').value = '0';
    }
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
        const csrfToken = document.getElementById('csrf_token').value;
        let payload = {};

        if (isHost) {
            const authType = document.getElementById('authType').value;
            const isEdit = e.target.action.includes('/edit/');
            
            payload = {
                name: document.getElementById('name').value,
                address: document.getElementById('address').value,
                port: document.getElementById('port').value.toString(),
                username: document.getElementById('username').value,
                auth_type: authType,
                settings: {
                    default_path: document.getElementById('defaultPath').value.trim() || "/"
                },
                csrf_token: csrfToken
            };

            if (authType === 'key') {
                const kVal = document.getElementById('keyID').value;
                if (kVal === "0" || !kVal) {
                    showErrorModal("Error: You must select a valid SSH key.");
                    return;
                }
                payload.key_id = kVal.toString();
                payload.password = ""; 
            } else {
                const pVal = document.getElementById('hostPassword').value.trim();
                
                // If this is an edit, an empty password is allowed..
                if (!isEdit && !pVal) {
                    showErrorModal("Error: Password cannot be empty.");
                    return;
                }
                
                payload.password = pVal; // Can be an empty string when Editing
                payload.key_id = "0"; 
            }
        } else {
            const kname = document.getElementById('keyName').value.trim();
            const kdata = document.getElementById('privateKey').value.trim();
            if (!kname || !kdata) {
                showErrorModal("Fields cannot be empty");
                return;
            }
            payload = {
                name: kname,
                key_data: kdata,
                csrf_token: csrfToken
            };
        }

        const targetUrl = e.target.action;
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

    try {
        const r = await fetch('/keys/list');
        const res = await r.json();
 
        const keys = (res && res.data && Array.isArray(res.data)) ? res.data : [];
        select.innerHTML = '<option value="0">-- No selected Keys --</option>';

        // If there are keys, we render them.
        keys.forEach(k => {
            const option = document.createElement('option');
            option.value = k.id;
            option.textContent = k.name;
            select.appendChild(option);
        });
    } catch (err) {
        select.innerHTML = '<option value="0">-- Error loading keys --</option>';
    }
}

/* --- TERMINAL AND SSH --- */
window.connectToHost = function(id, name, defaultPath = "/") {
    if (activeTerminals[id]) {
        window.maximizeTerminal(id);
        return;
    }

    // We save the path directly to the object.
    const initialPath = defaultPath || "/";

    const termWrapper = document.createElement('div');
    termWrapper.id = `term-win-${id}`;
    termWrapper.className = 'term-window';
    bringToFront(termWrapper);
    termWrapper.addEventListener('mousedown', () => bringToFront(termWrapper));

    termWrapper.addEventListener('click', (e) => {
        // If you clicked neither on a button nor on a link
        if (!e.target.closest('button') && !e.target.closest('a')) {
            term.focus(); 
        }
    });

    const offset = Object.keys(activeTerminals).length * 30;
    termWrapper.style.left = (50 + offset) + 'px';
    termWrapper.style.top = (50 + offset) + 'px';
    termWrapper.style.width = '800px';
    termWrapper.style.height = '500px';

    termWrapper.innerHTML = `
        <div class="term-header" id="header-${id}">
            <div class="term-header-left">
                <i class="fas fa-terminal" style="color: #888;"></i> 
                <span class="term-hostname">${name}</span>
            </div>
            
            <div class="term-tabs-center">
                <span class="tab-btn active" id="tab-term-${id}" onclick="window.switchTab(${id}, 'term')">Terminal</span>
                <span class="tab-btn" id="tab-files-${id}" onclick="window.switchTab(${id}, 'files')">Files</span>
            </div>

            <div class="term-controls">
                <button onclick="minimizeTerminal(${id})" title="Свернуть">_</button>
                <button onclick="maximizeTerminal(${id})" title="Развернуть/Сжать">&#9633;</button>
                <button onclick="window.closeTerminal(${id})" class="btn-close" title="Закрыть">&times;</button>
            </div>
        </div>
        
        <div class="window-content-viewer">
            <div id="term-${id}" class="term-body"></div>
            
            <div id="files-${id}" class="files-body" style="display:none;">
                <div class="sftp-toolbar">
                    <div class="sftp-toolbar-main" id="toolbar-main-${id}">
                        <button class="term-btn" onclick="window.goUp(${id})"><i class="fas fa-arrow-up"></i> Up</button>
                        <button class="term-btn" onclick="window.toggleSftpSelectMode(${id})"><i class="fas fa-check-square"></i> Select</button>
                        <button class="term-btn" onclick="document.getElementById('upload-input-${id}').click()"><i class="fas fa-upload"></i> Upload</button>
                        <input type="file" id="upload-input-${id}" multiple style="display:none" onchange="window.handleUpload(${id}, this)">
                    </div>
                    
                    <div class="sftp-toolbar-select" id="toolbar-select-${id}" style="display:none;">
                        <button class="term-btn term-btn-action" onclick="window.downloadSelected(${id})">Download Zip</button>
                        <button class="term-btn" onclick="window.toggleSftpSelectMode(${id})">Cancel</button>
                        <span id="select-count-${id}" class="select-count-text">Selected: 0</span>
                    </div>

                    <span id="path-${id}" class="sftp-path">/</span>
                </div>
                <div id="file-list-${id}" class="file-list-area"></div>
            </div>

            <div class="term-toolbar" id="toolbar-${id}">            
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
                    <button class="term-btn" onclick="sendSpecialKey(${id}, 'PgUp')">PU</button>
                    <button class="term-btn" onclick="sendSpecialKey(${id}, 'PgDn')">PD</button>
                </div>
                <div class="term-divider"></div>
                <div class="term-toolbar-group">
                    <button class="term-btn term-btn-danger" onclick="sendSpecialKey(${id}, 'Ctrl+C')">Ctrl+C</button>
                    <button class="term-btn term-btn-danger" onclick="sendSpecialKey(${id}, 'Delete')">Del</button>
                </div>
                <div class="term-divider"></div>
                <div class="term-toolbar-group">
                    <button id="sel-btn-${id}" class="term-btn term-btn-action" onclick="toggleSelectMode(${id})">Sel Off</button>
                    <button class="term-btn term-btn-action" onclick="pasteToTerminal(${id})">Paste</button>
                </div>
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
        term: term, 
        ws: ws, 
        fitAddon: fitAddon, 
        window: termWrapper,
        currentPath: initialPath,
        selectionMode: false,
        selectedFiles: new Set(),
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

    term.onData(data => { 
        const currentWs = activeTerminals[id]?.ws;
        if (currentWs && currentWs.readyState === WebSocket.OPEN) {
            currentWs.send(data);
        }
    });

    ws.onopen = () => {
        updateStatusBtn(id, true);
    };

    ws.onmessage = e => {
        if (e.data === "[STOPSESSION]") {
            updateStatusBtn(id, false); 
            setTimeout(() => window.closeTerminal(id), 800);
            return;
        }
        term.write(e.data);
    };


    ws.onclose = () => {
        updateStatusBtn(id, false);
        // If the client did not close the terminal themselves,
        // we try to restore the connection.
        if (activeTerminals[id] && !activeTerminals[id].explicitlyClosed) {
            setTimeout(() => silentReconnect(id), 2000);
        }
    };

    makeDraggable(termWrapper, document.getElementById(`header-${id}`));
    makeResizable(termWrapper, id);

    setTimeout(() => { fitAddon.fit(); term.focus(); }, 200);
};

/* --- AUXILIARY FUNCTIONS (DRAG, RESIZE, STATUS) --- */
function makeDraggable(el, header) {
    let pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;

    // Mouse handling on desktop when dragging.
    header.onmousedown = function(e) {
        if (e.target.closest('button')) return;
        e.preventDefault();
        
        pos3 = e.clientX;
        pos4 = e.clientY;
        
        document.onmouseup = closeDragElement;
        document.onmousemove = elementDrag;
    };

    function elementDrag(e) {
        e.preventDefault();
        pos1 = pos3 - e.clientX;
        pos2 = pos4 - e.clientY;
        pos3 = e.clientX;
        pos4 = e.clientY;
        el.style.top = (el.offsetTop - pos2) + "px";
        el.style.left = (el.offsetLeft - pos1) + "px";
        el.style.transform = "none"; 
    }

    function closeDragElement() {
        document.onmouseup = null;
        document.onmousemove = null;
    }

    // processing a wheelbarrow on a mobile phone.
    header.ontouchstart = function(e) {
        if (e.target.closest('button')) return;
        
        pos3 = e.touches[0].clientX;
        pos4 = e.touches[0].clientY;

        document.ontouchend = () => {
            document.ontouchend = null;
            document.ontouchmove = null;
        };

        document.ontouchmove = (ev) => {
            // Preventing page scrolling while dragging
            ev.preventDefault(); 
            
            pos1 = pos3 - ev.touches[0].clientX;
            pos2 = pos4 - ev.touches[0].clientY;
            pos3 = ev.touches[0].clientX;
            pos4 = ev.touches[0].clientY;
            
            el.style.top = (el.offsetTop - pos2) + "px";
            el.style.left = (el.offsetLeft - pos1) + "px";
            el.style.transform = "none";
        };
    };
}

window.addEventListener("orientationchange", () => {
    // Wait for the turn animation to finish in the mobile system.
    setTimeout(() => {
        Object.values(activeTerminals).forEach(data => {
            if (data && data.fitAddon) {
                data.fitAddon.fit();
                // If the window has become wider than the screen, reset it to 0.0.
                if (parseInt(data.window.style.left) > window.innerWidth) {
                    data.window.style.left = '10px';
                    data.window.style.top = '10px';
                }
            }
        });
    }, 300);
});

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

    // First, disable the observer so that it does not send resize to a closed socket.
    if (data.resizeObserver) {
        data.resizeObserver.disconnect();
    }

    // We set a flag so that the reconnect logic does not try to revive it.
    data.explicitlyClosed = true;

    // Close the socket before deleting the window.
    if (data.ws) {
        // We remove the onclose handler so that it does not call silentReconnect.
        data.ws.onclose = null; 
        data.ws.close();
    }

    updateStatusBtn(id, false);
    
    // Notify the server (optional, as Close on the socket should do this).
    const csrfToken = document.getElementById('csrf_token')?.value;
    try {
        await fetch(`/ssh/terminate?id=${id}`, { 
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ csrf_token: csrfToken })
        });
    } catch (e) {}

    if (data.window) {
        data.window.remove();
    }

    // Delete from memory.
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
    const keyMap = { 
        'Tab': '\t', 
        'Esc': '\x1b', 
        'Ctrl+C': '\x03', 
        'Up': '\x1b[A', 
        'Down': '\x1b[B', 
        'Left': '\x1b[D', 
        'Right': '\x1b[C',
        'Delete': '\x1b[3~',
        'PgUp': '\x1b[5~',
        'PgDn': '\x1b[6~'
    };
    if (keyMap[key]) { data.ws.send(keyMap[key]); data.term.focus(); }
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
        btn.innerText = 'Sel ON';
        layer.scrollTop = layer.scrollHeight;
    } else {
        layer.remove();
        btn.innerText = 'Sel Off';
        setTimeout(() => { data.term.focus(); data.term.scrollToBottom(); }, 50);
    }
};

window.visualViewport.addEventListener('resize', () => {
    const offset = window.innerHeight - window.visualViewport.height;
    const isKeyboardOpen = offset > 50;

    Object.entries(activeTerminals).forEach(([id, t]) => {
        const toolbar = t.window.querySelector('.term-toolbar');
        if (!toolbar) return;

        // Move the toolbar ONLY to the window that is currently active/in the foreground
        if (isKeyboardOpen) {
            toolbar.style.marginBottom = offset + 'px';
        } else {
            toolbar.style.marginBottom = '0';
        }

        // Resizing function
        const adjustTerminal = () => {
            t.fitAddon.fit();
            requestAnimationFrame(() => {
                t.term.scrollToBottom();

                if (t.ws && t.ws.readyState === WebSocket.OPEN) {
                    t.ws.send(JSON.stringify({ 
                        type: "resize", 
                        cols: t.term.cols, 
                        rows: t.term.rows 
                    }));
                }
            });
        };

        // We are making adjustments
        adjustTerminal();
        setTimeout(adjustTerminal, 300);
    });
});

// Silent socket recovery function.
function silentReconnect(id) {
    const tData = activeTerminals[id];
    if (!tData || tData.explicitlyClosed) return;
    if (tData.ws && tData.ws.readyState <= 1) return;

    console.log(`Reconnecting to host ${id}...`);
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${protocol}://${window.location.host}/ws/ssh?id=${id}`);

    ws.onopen = () => {
        tData.ws = ws; // Replace the socket in object.
        updateStatusBtn(id, true);
        tData.fitAddon.fit();
        ws.send(JSON.stringify({ type: "resize", cols: tData.term.cols, rows: tData.term.rows }));
    };

    ws.onmessage = e => {
        if (e.data === "[STOPSESSION]") {
            window.closeTerminal(id);
            return;
        }
        tData.term.write(e.data);
    };

    ws.onclose = () => {
        updateStatusBtn(id, false);
        if (activeTerminals[id] && !activeTerminals[id].explicitlyClosed) {
            setTimeout(() => silentReconnect(id), 5000);
        }
    };
}

// Checking when come back in the tab
document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
        Object.values(activeTerminals).forEach(t => {
            // If the tab is active and the socket is dead, we initiate a reconnect.
            const hostId = Object.keys(activeTerminals).find(key => activeTerminals[key] === t);
            if (hostId && (!t.ws || t.ws.readyState >= 2)) {
                silentReconnect(hostId);
            }
        });
    }
});

window.switchTab = function(id, tab) {
    const termBody = document.getElementById(`term-${id}`);
    const filesBody = document.getElementById(`files-${id}`);
    const toolbar = document.getElementById(`toolbar-${id}`);
    const tabTerm = document.getElementById(`tab-term-${id}`);
    const tabFiles = document.getElementById(`tab-files-${id}`);

    if (tab === 'term') {
        // Show the terminal, hide files.
        filesBody.style.display = 'none';
        
        if (window.innerWidth <= 768) toolbar.style.setProperty('display', 'flex', 'important');
        tabTerm.classList.add('active');
        tabFiles.classList.remove('active');
        
        if (activeTerminals[id]) {
            const t = activeTerminals[id];
            // A short pause for the DOM to recover
            setTimeout(() => {
                t.fitAddon.fit();
                t.term.focus();
                t.term.refresh(0, t.term.rows - 1);
            }, 30);
        }
    } else {
        // Show files on top of the terminal
        filesBody.style.display = 'flex';
        
        toolbar.style.setProperty('display', 'none', 'important');
        tabTerm.classList.remove('active');
        tabFiles.classList.add('active');
        
        window.loadFiles(id, activeTerminals[id].currentPath);
    }
};

window.loadFiles = async function(hostID, path) {
    const listArea = document.getElementById(`file-list-${hostID}`);
    const pathLabel = document.getElementById(`path-${hostID}`);
    
    listArea.innerHTML = '<div style="padding:15px; color: #888;">Loading...</div>';

    try {
        const url = `/sftp/list?host_id=${hostID}&path=${encodeURIComponent(path)}`;
        const r = await fetch(url);
        const res = await r.json();

        if (!res.success) {
            listArea.innerHTML = `<div style="color:#ff5555; padding:15px;">${res.message}</div>`;
            return;
        }

        const actualPath = res.data.current_path
        activeTerminals[hostID].currentPath = actualPath;
        renderBreadcrumbs(hostID, actualPath);
        
        let html = '<table class="sftp-table" style="width:100%; table-layout: fixed;"><tbody>';

        const files = res.data.files || [];
        
        if (files.length > 0) {
            files.forEach(f => {
                const icon = f.is_dir ? '📁' : '📄';
                const date = f.mod_time ? new Date(f.mod_time).toLocaleString('ru-RU', {day:'2-digit', month:'2-digit', year:'2-digit', hour:'2-digit', minute:'2-digit'}) : '';
                
                html += `
                    <tr id="file-${hostID}-${f.name}" onclick="window.handleSftpClick(${hostID}, '${f.name}', ${f.is_dir}, event)">
                        <td class="col-check" style="display: ${activeTerminals[hostID].selectionMode ? 'table-cell' : 'none'}">
                            <input type="checkbox" ${activeTerminals[hostID].selectedFiles?.has(f.name) ? 'checked' : ''} onchange="event.stopPropagation()">
                        </td>
                        <td style="width: 35px; padding-left: 10px; text-align: center;">${icon}</td>
                        <td class="file-name" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; padding-left: 5px;">${f.name}</td>
                        <td style="width: 120px; color: #666; font-size: 11px; white-space: nowrap; text-align: right;">${date}</td>
                        <td class="file-size" style="width: 80px; padding-right: 15px; text-align: right;">${f.is_dir ? '' : formatSize(f.size)}</td>
                    </tr>`;
            });
        } else {
            html += '<tr><td colspan="4" style="padding:15px; color:#666; text-align:center;">Directory is empty</td></tr>';
        }
        
        html += '</tbody></table>';
        listArea.innerHTML = html;
        // Scroll to top when changing folder.
        listArea.scrollTop = 0; 
    } catch (e) {
        listArea.innerHTML = `<div style="padding:15px; color:#ff5555;">Load failed: ${e.message}</div>`;
        console.error(e);
    }
};

window.goUp = function(hostID) {
    let path = activeTerminals[hostID].currentPath || "/";
    
    // Remove the slash at the end, if there is one..
    path = path.replace(/\/+$/, "");
    
    // Если путь пустой или уже корень, ничего не делаем
    if (!path || path === "") {
        window.loadFiles(hostID, "/");
        return;
    }

    // We break the path, remove the last piece.
    const parts = path.split('/').filter(Boolean);
    parts.pop();
    
    // Let's put it back together. If the array is empty, it means we've reached the root.
    const parentPath = "/" + parts.join('/');
    
    window.loadFiles(hostID, parentPath);
};

window.formatSize = function(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// Function for generating a clickable path.
function renderBreadcrumbs(id, path) {
    const container = document.getElementById(`path-${id}`);
    if (!container) return;

    const parts = path.split('/').filter(Boolean);
    let currentAccumulatedPath = '';
    
    let html = `<span style="cursor:pointer; color:#00ff00;" onclick="window.loadFiles(${id}, '/')"><i class="fas fa-hdd"></i> /</span> `;
    
    parts.forEach((part, index) => {
        currentAccumulatedPath += '/' + part;
        const isLast = index === parts.length - 1;

        // Escape the quotes in the onclick path to avoid breaking the html.
        const safePath = currentAccumulatedPath.replace(/'/g, "\\'");
        if (isLast) {
            html += `<span style="color:#eee;">${part}</span>`;
        } else {
            html += `<span style="cursor:pointer; color:#00ff00;" onclick="window.loadFiles(${id}, '${currentAccumulatedPath}')">${part}</span> / `;
        }
    });
    container.innerHTML = html;
}

window.toggleSftpSelectMode = function(hostID) {
    const t = activeTerminals[hostID];
    if (!t) return;
    
    t.selectionMode = !t.selectionMode;
    
    const mainToolbar = document.getElementById(`toolbar-main-${hostID}`);
    const selectToolbar = document.getElementById(`toolbar-select-${hostID}`);
    
    if (t.selectionMode) {
        mainToolbar.style.display = 'none';
        selectToolbar.style.display = 'flex';
        t.selectedFiles = new Set();
    } else {
        mainToolbar.style.display = 'flex';
        selectToolbar.style.display = 'none';
        t.selectedFiles.clear();
    }
    
    // Redrawing the list to show/hide the checkbox column.
    window.loadFiles(hostID, t.currentPath);
    updateSelectCount(hostID);
};

window.handleSftpClick = function(hostID, name, isDir, event) {
    const t = activeTerminals[hostID];
    if (!t) return;
    
    if (t.selectionMode) {
        const tr = document.getElementById(`file-${hostID}-${name}`);
        const cb = tr.querySelector('input[type="checkbox"]');
        
        if (t.selectedFiles.has(name)) {
            t.selectedFiles.delete(name);
            tr.classList.remove('selected');
            if(cb) cb.checked = false;
        } else {
            t.selectedFiles.add(name);
            tr.classList.add('selected');
            if(cb) cb.checked = true;
        }
        updateSelectCount(hostID);
    } else {
        if (isDir) {
            let current = t.currentPath;
            if (!current.endsWith('/')) current += '/';
            window.loadFiles(hostID, current + name);
        } else {
            window.downloadFile(hostID, name);
        }
    }
};

function updateSelectCount(hostID) {
    const t = activeTerminals[hostID];
    const count = t && t.selectedFiles ? t.selectedFiles.size : 0;
    const label = document.getElementById(`select-count-${hostID}`);
    if (label) {
        label.innerText = `Selected: ${count}`;
        label.className = "select-count-text";
    }
}

window.downloadFile = function(id, name) {
    const t = activeTerminals[id];
    if (!t) return;

    const path = t.currentPath;
    // We use path normalization to avoid double slashes.
    const fullPath = path.endsWith('/') ? path + name : path + '/' + name;
    const csrfToken = document.getElementById('csrf_token')?.value || "";

    // Generating a URL with a token.
    const params = new URLSearchParams({
        host_id: id,
        path: fullPath,
        csrf_token: csrfToken
    });

    // Click on the link and the browser will start downloading.
    window.location.href = `/sftp/download?${params.toString()}`;
};

window.downloadSelected = function(hostID) {
    const t = activeTerminals[hostID];
    if (!t || !t.selectedFiles || t.selectedFiles.size === 0) return;

    const filesArray = Array.from(t.selectedFiles);
    const csrfToken = document.getElementById('csrf_token')?.value || "";

    // We collect all the parameters into a string.
    const params = new URLSearchParams({
        host_id: hostID,
        parent_path: t.currentPath,
        files: JSON.stringify(filesArray),
        csrf_token: csrfToken
    });

    // This will force the mobile to open the native download manager.
    window.location.href = `/sftp/download-zip?${params.toString()}`;

    setTimeout(() => {
        window.toggleSftpSelectMode(hostID);
    }, 500);
};

window.handleUpload = async function(hostID, input) {
    const files = input.files;
    if (files.length === 0) return;

    const t = activeTerminals[hostID];
    const formData = new FormData();
    formData.append('host_id', hostID);
    formData.append('remote_path', t.currentPath);
    formData.append('csrf_token', document.getElementById('csrf_token')?.value || "");

    for (let i = 0; i < files.length; i++) {
        formData.append('files', files[i]);
    }

    try {
        const response = await fetch('/sftp/upload', {
            method: 'POST',
            body: formData
        });
        const res = await response.json();
        if (res.success) {
            window.loadFiles(hostID, t.currentPath);
        } else {
            alert("Upload failed: " + res.message);
        }
    } catch (e) {
        console.error(e);
        showErrorModal("Upload error");
    }
    input.value = '';
};
