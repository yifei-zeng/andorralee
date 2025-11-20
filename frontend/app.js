/**
 * AndorraLee Frontend Application
 * Single Page Application logic
 */

const API_BASE = '/api/v1';

// --- Utility Functions ---
const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => document.querySelectorAll(selector);

async function apiCall(endpoint, method = 'GET', data = null) {
    try {
        const options = {
            method: method,
            headers: { 'Content-Type': 'application/json' },
        };
        if (data && method !== 'GET') {
            options.body = JSON.stringify(data);
        }
        const response = await fetch(API_BASE + endpoint, options);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        showToast(`请求失败: ${error.message}`, 'error');
        console.error('API Error:', error);
        return null;
    }
}

function showToast(message, type = 'info') {
    const container = $('#toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
        <span style="margin-left: 10px">${message}</span>
    `;
    container.appendChild(toast);
    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => container.removeChild(toast), 300);
    }, 3000);
}

function updateTime() {
    const now = new Date();
    $('#current-time').textContent = now.toLocaleString('zh-CN');
}

// --- Router ---
const routes = {
    'dashboard': renderDashboard,
    'containers': renderContainers,
    'ports': renderPorts,
    'malware': renderMalware,
    'logs': renderLogs,
    'threats': renderThreats,
    'rules': renderRules,
    'settings': renderSettings
};

function navigateTo(route) {
    // Update Sidebar
    $$('.nav-item').forEach(el => el.classList.remove('active'));
    const activeLink = $(`.nav-item[data-route="${route}"]`);
    if (activeLink) activeLink.classList.add('active');

    // Update Title
    const titles = {
        'dashboard': '仪表盘',
        'containers': '容器管理',
        'ports': '端口管理',
        'malware': '病毒检测',
        'logs': '日志审计',
        'threats': '威胁情报',
        'rules': '安全规则',
        'settings': '系统设置'
    };
    $('#page-title').textContent = titles[route] || 'AndorraLee';

    // Render View
    const container = $('#view-container');
    container.innerHTML = '<div class="loading"><i class="fas fa-spinner fa-spin fa-2x"></i><p>加载中...</p></div>';
    
    if (routes[route]) {
        routes[route](container);
    } else {
        container.innerHTML = '<div class="empty-state">页面未找到</div>';
    }
}

// --- View Renderers ---

async function renderDashboard(container) {
    // Fetch stats in parallel
    const [malwareStats, portStats, containerStats] = await Promise.all([
        apiCall('/malware/statistics'),
        apiCall('/ports/statistics'),
        apiCall('/temp-containers/id-status')
    ]);

    container.innerHTML = `
        <div class="grid grid-3">
            <div class="card stat-card">
                <div class="stat-icon"><i class="fas fa-bug"></i></div>
                <div class="stat-info">
                    <h4>${malwareStats?.total_scans || 0}</h4>
                    <p>总病毒扫描</p>
                </div>
            </div>
            <div class="card stat-card">
                <div class="stat-icon"><i class="fas fa-network-wired"></i></div>
                <div class="stat-info">
                    <h4>${portStats?.data?.total_allocated || 0}</h4>
                    <p>已分配端口</p>
                </div>
            </div>
            <div class="card stat-card">
                <div class="stat-icon"><i class="fas fa-docker"></i></div>
                <div class="stat-info">
                    <h4>${containerStats?.data?.total_instances || 0}</h4>
                    <p>运行容器</p>
                </div>
            </div>
        </div>
        
        <div class="grid grid-2">
            <div class="card">
                <div class="card-header">
                    <h3><i class="fas fa-shield-alt"></i> 最近威胁</h3>
                </div>
                <div class="card-body">
                    <p class="text-muted">暂无最近威胁数据展示 (Mock)</p>
                </div>
            </div>
            <div class="card">
                <div class="card-header">
                    <h3><i class="fas fa-server"></i> 系统资源</h3>
                </div>
                <div class="card-body">
                    <p class="text-muted">CPU / 内存 使用率 (Mock)</p>
                </div>
            </div>
        </div>
    `;
}

async function renderContainers(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header">
                <h3>容器实例</h3>
                <div class="flex">
                    <button class="btn btn-sm btn-primary" onclick="showCreateContainerModal()"><i class="fas fa-plus"></i> 创建容器</button>
                    <button class="btn btn-sm btn-outline" onclick="refreshContainerList()"><i class="fas fa-sync"></i> 刷新</button>
                </div>
            </div>
            <div class="card-body">
                <div id="container-list" class="grid">加载中...</div>
            </div>
        </div>
        
        <div class="card">
            <div class="card-header"><h3>镜像管理</h3></div>
            <div class="card-body">
                <div class="form-group flex">
                    <input type="text" id="pull-image-name" class="form-control" placeholder="输入镜像名称 (e.g. nginx:latest)">
                    <button class="btn btn-primary" onclick="pullImage()">拉取镜像</button>
                </div>
                <div id="image-list" class="grid grid-4"></div>
            </div>
        </div>
    `;
    refreshContainerList();
    refreshImageList();
}

async function refreshContainerList() {
    const res = await apiCall('/temp-containers');
    const list = $('#container-list');
    if (!res || !res.data || res.data.length === 0) {
        list.innerHTML = '<div class="empty-state">暂无运行中的容器</div>';
        return;
    }
    
    list.innerHTML = res.data.map(c => `
        <div class="card" style="margin-bottom:0">
            <div class="card-body">
                <div class="flex flex-between mb-2">
                    <strong>${c.name}</strong>
                    <span class="badge badge-${c.status === 'running' ? 'success' : 'danger'}">${c.status}</span>
                </div>
                <p class="small text-muted">ID: ${c.id}</p>
                <p class="small text-muted">Image: ${c.image_name}</p>
                <p class="small text-muted">Proto: ${c.protocol}</p>
                <div class="mt-2 flex">
                    <button class="btn btn-sm btn-success" onclick="controlContainer('${c.id}', 'start')"><i class="fas fa-play"></i></button>
                    <button class="btn btn-sm btn-danger" onclick="controlContainer('${c.id}', 'stop')"><i class="fas fa-stop"></i></button>
                    <button class="btn btn-sm btn-warning" onclick="controlContainer('${c.id}', 'restart')"><i class="fas fa-redo"></i></button>
                    <button class="btn btn-sm btn-outline" onclick="deleteContainer('${c.id}')"><i class="fas fa-trash"></i></button>
                </div>
            </div>
        </div>
    `).join('');
}

async function refreshImageList() {
    const res = await apiCall('/docker/images');
    const list = $('#image-list');
    if (res && res.data) {
        list.innerHTML = res.data.map(img => `
            <div class="card" style="margin-bottom:0; font-size:0.8rem">
                <div class="card-body">
                    <strong>${img.repository}:${img.tag}</strong>
                    <p class="text-muted">${(img.size / 1024 / 1024).toFixed(1)} MB</p>
                </div>
            </div>
        `).join('');
    }
}

window.controlContainer = async (id, action) => {
    await apiCall(`/temp-containers/${id}/${action}`, 'POST');
    showToast(`容器 ${action} 指令已发送`, 'success');
    setTimeout(refreshContainerList, 1000);
};

window.deleteContainer = async (id) => {
    if(confirm('确定删除该容器?')) {
        await apiCall(`/temp-containers/${id}`, 'DELETE');
        showToast('容器已删除', 'success');
        setTimeout(refreshContainerList, 1000);
    }
};

window.pullImage = async () => {
    const name = $('#pull-image-name').value;
    if(!name) return showToast('请输入镜像名', 'warning');
    showToast('开始拉取镜像...', 'info');
    await apiCall('/docker/pull', 'POST', { image_name: name });
    showToast('镜像拉取请求已发送', 'success');
};

window.showCreateContainerModal = () => {
    // Render a lightweight modal form with image selection
    if (document.querySelector('#create-container-modal')) return;
    const modal = document.createElement('div');
    modal.id = 'create-container-modal';
    modal.style = 'position:fixed;left:0;top:0;right:0;bottom:0;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.4);z-index:2000';
    modal.innerHTML = `
        <div style="background:#fff;border-radius:8px;width:520px;max-width:95%;padding:18px;box-shadow:0 8px 24px rgba(0,0,0,0.2)">
            <h3 style="margin:0 0 10px">创建容器</h3>
            <div style="display:grid;gap:8px">
                <input id="cc-name" class="form-control" placeholder="容器名称" />
                <label style="font-size:13px;color:var(--text-muted)">选择镜像</label>
                <select id="cc-image-select" class="form-control">
                    <option value="andorralee/cowrie:v0.1">cowrie — andorralee/cowrie:v0.1</option>
                    <option value="andorralee/heralding:v0.1">heralding — andorralee/heralding:v0.1</option>
                    <option value="andorralee/mysql-preseed:8.0-arm64">mysql — andorralee/mysql-preseed:8.0-arm64</option>
                    <option value="custom">自定义镜像...</option>
                </select>
                <input id="cc-image-custom" class="form-control" placeholder="自定义镜像 (e.g. nginx:latest)" style="display:none" />
                <div style="display:flex;gap:8px">
                    <select id="cc-proto" class="form-control" style="width:160px">
                        <option value="http">HTTP</option>
                        <option value="ssh">SSH</option>
                        <option value="mysql">MySQL</option>
                    </select>
                    <input id="cc-ports" class="form-control" placeholder="端口映射（格式:80:auto,2222:22）" />
                </div>
                <div style="display:flex;justify-content:flex-end;gap:8px;margin-top:6px">
                    <button id="cc-cancel" class="btn btn-outline">取消</button>
                    <button id="cc-create" class="btn btn-primary">创建</button>
                </div>
            </div>
        </div>
    `;
    document.body.appendChild(modal);

    const sel = document.getElementById('cc-image-select');
    const custom = document.getElementById('cc-image-custom');
    sel.addEventListener('change', () => {
        if (sel.value === 'custom') custom.style.display = 'block';
        else custom.style.display = 'none';
    });

    document.getElementById('cc-cancel').addEventListener('click', () => {
        document.body.removeChild(modal);
    });

    document.getElementById('cc-create').addEventListener('click', async () => {
        const name = document.getElementById('cc-name').value.trim();
        let image = sel.value === 'custom' ? document.getElementById('cc-image-custom').value.trim() : sel.value;
        const protocol = document.getElementById('cc-proto').value;
        const portsRaw = document.getElementById('cc-ports').value.trim();
        if (!name) return showToast('请填写容器名称', 'warning');
        if (!image) return showToast('请填写镜像', 'warning');

        // parse simple port mapping format key:value pairs separated by comma
        const port_mappings = {};
        if (portsRaw) {
            portsRaw.split(',').forEach(p => {
                const parts = p.split(':').map(s => s.trim());
                if (parts.length === 2 && parts[0]) port_mappings[parts[0]] = parts[1] || 'auto';
            });
        } else {
            // default mapping for common protocols
            if (protocol === 'http') port_mappings['80'] = 'auto';
            if (protocol === 'ssh') port_mappings['22'] = 'auto';
            if (protocol === 'mysql') port_mappings['3306'] = 'auto';
        }

        const payload = {
            name,
            image_name: image,
            protocol,
            port_mappings,
            environment: {}
        };

        const res = await apiCall('/temp-containers', 'POST', payload);
        if (res) {
            showToast('创建请求已发送', 'success');
            setTimeout(refreshContainerList, 1000);
            document.body.removeChild(modal);
        }
    });
};

async function renderPorts(container) {
    container.innerHTML = `
        <div class="grid grid-2">
            <div class="card">
                <div class="card-header"><h3>端口检查</h3></div>
                <div class="card-body">
                    <div class="form-group flex">
                        <input type="number" id="check-port-input" class="form-control" placeholder="端口号">
                        <button class="btn btn-info" onclick="checkPort()">检查</button>
                    </div>
                    <div id="port-check-result"></div>
                </div>
            </div>
            <div class="card">
                <div class="card-header"><h3>端口分配</h3></div>
                <div class="card-body">
                    <div class="form-group">
                        <input type="text" id="alloc-container" class="form-control" placeholder="容器ID">
                    </div>
                    <div class="form-group">
                        <select id="alloc-service" class="form-control">
                            <option value="http">HTTP</option>
                            <option value="ssh">SSH</option>
                            <option value="mysql">MySQL</option>
                        </select>
                    </div>
                    <button class="btn btn-success" onclick="allocatePort()">自动分配</button>
                </div>
            </div>
        </div>
        <div class="card">
            <div class="card-header">
                <h3>已分配端口</h3>
                <button class="btn btn-sm btn-outline" onclick="refreshAllocatedPorts()">刷新</button>
            </div>
            <div class="card-body">
                <div id="allocated-ports-list" class="grid grid-4"></div>
            </div>
        </div>
    `;
    refreshAllocatedPorts();
}

window.checkPort = async () => {
    const port = $('#check-port-input').value;
    if(!port) return;
    const res = await apiCall(`/ports/${port}/check`);
    $('#port-check-result').innerHTML = res ? 
        `<span class="text-${res.available ? 'success':'danger'}">${res.message}</span>` : '';
};

window.allocatePort = async () => {
    const cid = $('#alloc-container').value;
    const svc = $('#alloc-service').value;
    if(!cid) return showToast('请输入容器ID', 'warning');
    await apiCall('/ports/allocate', 'POST', { container_id: cid, service_type: svc, description: 'Manual' });
    showToast('端口分配成功', 'success');
    refreshAllocatedPorts();
};

async function refreshAllocatedPorts() {
    const res = await apiCall('/ports/allocated');
    const list = $('#allocated-ports-list');
    if(res && res.data) {
        list.innerHTML = res.data.map(p => `
            <div class="card" style="margin-bottom:0; border-left: 3px solid var(--success-color)">
                <div class="card-body" style="padding:10px">
                    <h4 style="margin:0">${p.port}</h4>
                    <p class="small text-muted">${p.service_type}</p>
                    <p class="small text-muted">Container: ${p.container_id}</p>
                    <button class="btn btn-sm btn-danger mt-2" onclick="releasePort(${p.port})">释放</button>
                </div>
            </div>
        `).join('');
    }
}

window.releasePort = async (port) => {
    if(confirm(`释放端口 ${port}?`)) {
        await apiCall(`/ports/${port}/release`, 'DELETE');
        refreshAllocatedPorts();
    }
};

async function renderMalware(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header"><h3>文件扫描</h3></div>
            <div class="card-body">
                <div class="form-group">
                    <input type="file" id="malware-file" class="form-control">
                </div>
                <button class="btn btn-primary" onclick="scanFile()">上传并扫描</button>
                <div id="scan-result" class="mt-2 log-viewer" style="height:200px; display:none"></div>
            </div>
        </div>
        <div class="card">
            <div class="card-header"><h3>扫描历史</h3></div>
            <div class="card-body table-responsive">
                <table class="table">
                    <thead><tr><th>Hash</th><th>文件名</th><th>状态</th><th>时间</th></tr></thead>
                    <tbody id="scan-history-body"></tbody>
                </table>
            </div>
        </div>
    `;
    refreshScanHistory();
}

window.scanFile = async () => {
    const file = $('#malware-file').files[0];
    if(!file) return showToast('请选择文件', 'warning');
    
    const formData = new FormData();
    formData.append('file', file);
    
    $('#scan-result').style.display = 'block';
    $('#scan-result').textContent = '扫描中...';
    
    try {
        const res = await fetch(API_BASE + '/malware/scan/file', { method: 'POST', body: formData });
        const data = await res.json();
        $('#scan-result').textContent = JSON.stringify(data, null, 2);
        refreshScanHistory();
    } catch(e) {
        $('#scan-result').textContent = 'Error: ' + e.message;
    }
};

async function refreshScanHistory() {
    const res = await apiCall('/malware/scan/history');
    if(res && res.scan_history) {
        $('#scan-history-body').innerHTML = res.scan_history.map(h => `
            <tr>
                <td class="monospace">${h.sha256 ? h.sha256.substring(0,12)+'...' : '-'}</td>
                <td>${h.file_name}</td>
                <td><span class="badge badge-${h.is_infected ? 'danger':'success'}">${h.is_infected ? 'Infected':'Clean'}</span></td>
                <td>${h.scan_time}</td>
            </tr>
        `).join('');
    }
}

async function renderLogs(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header">
                <div class="tab-nav" style="margin-bottom:0; border:none">
                    <div class="tab-item active" onclick="switchLogTab('heralding')">Heralding</div>
                    <div class="tab-item" onclick="switchLogTab('cowrie')">Cowrie</div>
                    <div class="tab-item" onclick="switchLogTab('mysql')">MySQL</div>
                </div>
            </div>
            <div class="card-body">
                <div class="flex mb-2">
                    <input type="text" id="log-container-id" class="form-control" placeholder="容器ID (可选)">
                    <button class="btn btn-primary" onclick="pullLogs()">拉取日志</button>
                    <button class="btn btn-outline" onclick="fetchLogs()">刷新列表</button>
                </div>
                <div id="log-display" class="log-viewer"></div>
            </div>
        </div>
    `;
    window.currentLogType = 'heralding';
    fetchLogs();
}

window.switchLogTab = (type) => {
    $$('.tab-item').forEach(t => t.classList.remove('active'));
    event.target.classList.add('active');
    window.currentLogType = type;
    fetchLogs();
};

window.pullLogs = async () => {
    const cid = $('#log-container-id').value;
    const endpoint = {
        'heralding': '/heralding/pull-logs',
        'cowrie': '/cowrie/pull-logs',
        'mysql': '/mysql-honeypot/pull-logs'
    }[window.currentLogType];
    
    await apiCall(endpoint, 'POST', cid ? {container_id: cid} : {});
    showToast('日志拉取请求已发送', 'success');
    setTimeout(fetchLogs, 1000);
};

window.fetchLogs = async () => {
    const endpoint = {
        'heralding': '/heralding/logs',
        'cowrie': '/cowrie/logs',
        'mysql': '/mysql-honeypot/logs'
    }[window.currentLogType];
    
    const res = await apiCall(endpoint);
    const display = $('#log-display');
    if(res && (res.logs || res.data)) {
        const logs = res.logs || res.data;
        display.textContent = JSON.stringify(logs, null, 2);
    } else {
        display.textContent = '暂无日志';
    }
};

async function renderThreats(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header"><h3>威胁情报</h3></div>
            <div class="card-body">
                <p>开发中...</p>
            </div>
        </div>
    `;
}

async function renderRules(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header"><h3>安全规则</h3></div>
            <div class="card-body">
                <p>开发中...</p>
            </div>
        </div>
    `;
}

async function renderSettings(container) {
    const health = await apiCall('/health');
    container.innerHTML = `
        <div class="card">
            <div class="card-header"><h3>系统状态</h3></div>
            <div class="card-body">
                <pre>${JSON.stringify(health, null, 2)}</pre>
            </div>
        </div>
    `;
}

// --- Initialization ---
document.addEventListener('DOMContentLoaded', () => {
    // Sidebar Navigation
    $$('.nav-item').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const route = link.dataset.route;
            navigateTo(route);
        });
    });

    // Refresh Button
    $('#refresh-btn').addEventListener('click', () => {
        const active = $('.nav-item.active');
        if(active) navigateTo(active.dataset.route);
    });

    // Initial Route
    navigateTo('dashboard');    
    // Clock
    setInterval(updateTime, 1000);
    updateTime();    
    // Check System Status
    checkSystemStatus();
});

async function checkSystemStatus() {
    try {
        await fetch('/health');
        $('#system-status').innerHTML = '<span class="dot online"></span> 系统在线';
    } catch {
        $('#system-status').innerHTML = '<span class="dot offline"></span> 连接断开';
    }
}
