const { invoke } = window.__TAURI__.core;
const { listen } = window.__TAURI__.event;
const { open } = window.__TAURI__.dialog;

let config = {};
let pollInterval = null;

// 初始化
document.addEventListener('DOMContentLoaded', async () => {
    await loadConfig();
    await loadState();
    
    // 监听上传事件
    listen('upload-started', (event) => {
        loadState();
        playSound();
        showNotification('开始上传', `目录: ${getFolderName(event.payload)}`);
    });
    
    listen('upload-completed', (event) => {
        loadState();
        playSound();
        showNotification('上传完成', `目录: ${getFolderName(event.payload)}`);
    });
    
    // 绑定事件
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', () => switchTab(tab.dataset.tab));
    });
    
    document.getElementById('folderSelect').addEventListener('click', selectFolder);
    document.getElementById('btnStart').addEventListener('click', startMonitor);
    document.getElementById('btnStop').addEventListener('click', stopMonitor);
    document.getElementById('btnSignAll').addEventListener('click', signAllBatches);
    document.getElementById('btnClear').addEventListener('click', clearBatches);
    document.getElementById('btnSave').addEventListener('click', saveSettings);
    
    // 开关绑定
    document.getElementById('toggleSubdirs').addEventListener('click', () => toggleSetting('subdirs'));
    document.getElementById('toggleSound').addEventListener('click', () => toggleSetting('sound'));
    document.getElementById('toggleHistory').addEventListener('click', () => toggleSetting('history'));
});

function switchTab(tabName) {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    document.querySelector(`.tab[data-tab="${tabName}"]`).classList.add('active');
    document.getElementById('tab-' + tabName).classList.add('active');
}

async function loadConfig() {
    config = await invoke('get_config');
    updateConfigUI();
}

async function loadState() {
    const state = await invoke('get_state');
    updateStateUI(state);
}

function updateConfigUI() {
    document.getElementById('folderPath').textContent = config.watch_folder || '点击选择文件夹';
    document.getElementById('fileTypes').value = (config.file_types || []).join(', ');
    document.getElementById('ignoreFolders').value = (config.ignore_folders || []).join(', ');
    document.getElementById('toggleSubdirs').classList.toggle('active', config.watch_subdirs);
    document.getElementById('toggleSound').classList.toggle('active', config.sound_enabled);
    document.getElementById('toggleHistory').classList.toggle('active', config.save_history);
}

function updateStateUI(state) {
    const card = document.getElementById('statusCard');
    const icon = document.getElementById('statusIcon');
    const title = document.getElementById('statusTitle');
    const desc = document.getElementById('statusDesc');
    const btnStart = document.getElementById('btnStart');
    const btnStop = document.getElementById('btnStop');
    const unsignedBadge = document.getElementById('unsignedBadge');
    
    if (state.is_running) {
        card.classList.remove('stopped');
        icon.textContent = '✓';
        title.textContent = '监控运行中';
        const uploadingText = state.uploading_count > 0 ? ` | ${state.uploading_count} 个上传中` : '';
        desc.textContent = `监控目录${uploadingText}`;
        btnStart.disabled = true;
        btnStop.disabled = false;
        startPolling();
    } else {
        card.classList.add('stopped');
        icon.textContent = '⏸';
        title.textContent = '监控已停止';
        desc.textContent = '点击开始监控视频上传';
        btnStart.disabled = false;
        btnStop.disabled = true;
        stopPolling();
    }
    
    // 待签收数量
    if (state.unsigned_count > 0) {
        unsignedBadge.textContent = state.unsigned_count;
        unsignedBadge.style.display = 'inline';
    } else {
        unsignedBadge.style.display = 'none';
    }
    
    // 更新批次列表
    updateBatchList(state.batches || []);
}

function updateBatchList(batches) {
    const list = document.getElementById('batchList');
    
    if (!batches || batches.length === 0) {
        list.innerHTML = '<div class="no-batches">暂无上传批次</div>';
        return;
    }
    
    const statusText = {
        'Uploading': '⏳ 上传中',
        'Completed': '✅ 待签收',
        'Signed': '✓ 已签收'
    };
    
    const statusClass = {
        'Uploading': 'uploading',
        'Completed': 'completed',
        'Signed': 'signed'
    };
    
    list.innerHTML = batches.map(batch => `
        <div class="batch-item ${statusClass[batch.status]}">
            <div class="batch-header">
                <span class="batch-status ${statusClass[batch.status]}">
                    ${statusText[batch.status]}
                </span>
                <span class="batch-time">${batch.started_at}</span>
            </div>
            <div class="batch-folder">📁 ${getFolderName(batch.folder)}</div>
            <div class="file-count">🎬 ${batch.file_count} 个视频文件</div>
            <div class="batch-files">
                ${batch.files.slice(0, 5).map(f => `<div class="file-item">${getFileName(f)}</div>`).join('')}
                ${batch.files.length > 5 ? `<div class="file-item">+${batch.files.length - 5} 更多</div>` : ''}
            </div>
            ${batch.status === 'Completed' ? `<button class="btn-sign" onclick="signBatch('${batch.id}')">📝 确认签收</button>` : ''}
            ${batch.completed_at ? `<div class="batch-time" style="margin-top:5px;">完成: ${batch.completed_at}</div>` : ''}
        </div>
    `).join('');
}

function getFileName(path) {
    return path.split(/[/\\]/).pop();
}

function getFolderName(path) {
    const parts = path.split(/[/\\]/);
    return parts[parts.length - 1] || parts[parts.length - 2] || path;
}

async function selectFolder() {
    const selected = await open({
        directory: true,
        multiple: false,
        title: '选择要监控的文件夹'
    });
    
    if (selected) {
        config.watch_folder = selected;
        await invoke('set_config', { config });
        document.getElementById('folderPath').textContent = selected;
    }
}

async function startMonitor() {
    const success = await invoke('start_monitor');
    if (success) {
        loadState();
    } else {
        alert('启动监控失败，请检查文件夹是否存在');
    }
}

async function stopMonitor() {
    await invoke('stop_monitor');
    loadState();
}

window.signBatch = async function(batchId) {
    await invoke('sign_batch', { batchId });
    loadState();
}

async function signAllBatches() {
    await invoke('sign_all_batches');
    loadState();
}

async function clearBatches() {
    await invoke('clear_batches');
    loadState();
}

async function saveSettings() {
    config.file_types = document.getElementById('fileTypes').value
        .split(',').map(s => s.trim()).filter(s => s);
    config.ignore_folders = document.getElementById('ignoreFolders').value
        .split(',').map(s => s.trim()).filter(s => s);
    
    await invoke('set_config', { config });
    alert('设置已保存！');
}

function toggleSetting(type) {
    if (type === 'subdirs') {
        config.watch_subdirs = !config.watch_subdirs;
        document.getElementById('toggleSubdirs').classList.toggle('active', config.watch_subdirs);
    } else if (type === 'sound') {
        config.sound_enabled = !config.sound_enabled;
        document.getElementById('toggleSound').classList.toggle('active', config.sound_enabled);
    } else if (type === 'history') {
        config.save_history = !config.save_history;
        document.getElementById('toggleHistory').classList.toggle('active', config.save_history);
    }
}

function playSound() {
    if (config.sound_enabled) {
        try {
            const audio = new AudioContext();
            const oscillator = audio.createOscillator();
            const gain = audio.createGain();
            oscillator.connect(gain);
            gain.connect(audio.destination);
            oscillator.frequency.value = 800;
            oscillator.type = 'sine';
            gain.gain.value = 0.1;
            oscillator.start();
            oscillator.stop(audio.currentTime + 0.15);
        } catch(e) {}
    }
}

function showNotification(title, body) {
    if (Notification.permission === 'granted') {
        new Notification(title, { body });
    } else if (Notification.permission !== 'denied') {
        Notification.requestPermission().then(permission => {
            if (permission === 'granted') {
                new Notification(title, { body });
            }
        });
    }
}

function startPolling() {
    if (pollInterval) return;
    pollInterval = setInterval(loadState, 2000);
}

function stopPolling() {
    if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
    }
}
