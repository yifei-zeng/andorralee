/**
 * AndorraLee Frontend Application
 * Comprehensive Management Interface
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
        const result = await response.json();
        
        if (!response.ok) {
            throw new Error(result.message || result.error || `HTTP ${response.status}`);
        }
        return result;
    } catch (error) {
        showToast(`操作失败: ${error.message}`, 'error');
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
    'logs': renderLogs,
    'malware': renderMalware,
    'network': renderNetwork,
    'threats': renderThreats,
    'system': renderSystem
};

function navigateTo(route) {
    $$('.nav-item').forEach(el => el.classList.remove('active'));
    const activeLink = $(`.nav-item[data-route="${route}"]`);
    if (activeLink) activeLink.classList.add('active');

    const titles = {
        'dashboard': '仪表盘',
        'containers': '容器管理',
        'logs': '日志审计',
        'malware': '病毒检测',
        'network': '网络端口',
        'threats': '威胁情报',
        'system': '系统设置'
    };
    $('#page-title').textContent = titles[route] || 'AndorraLee';

    const container = $('#content-area');
    container.innerHTML = '<div class="loading-overlay"><i class="fas fa-spinner fa-spin fa-3x"></i></div>';
    
    if (routes[route]) {
        routes[route](container);
    } else {
        container.innerHTML = '<div class="section">页面未找到</div>';
    }
}

// --- View Renderers ---

// 1. Dashboard
async function renderDashboard(container) {
    const [malwareStats, portStats, containerStats, health] = await Promise.all([
        apiCall('/malware/statistics'),
        apiCall('/ports/statistics'),
        apiCall('/temp-containers/id-status'),
        apiCall('/health')
    ]);

    container.innerHTML = `
        <div class="grid-3">
            <div class="section" style="border-left: 4px solid var(--danger)">
                <div style="display:flex; justify-content:space-between; align-items:center">
                    <div>
                        <h3>病毒扫描</h3>
                        <h1 style="margin:0; font-size:36px">${malwareStats?.total_scans || 0}</h1>
                        <p style="color:#666; margin:5px 0">发现威胁: ${malwareStats?.infected_count || 0}</p>
                    </div>
                    <i class="fas fa-bug fa-3x" style="color:var(--danger); opacity:0.2"></i>
                </div>
            </div>
            <div class="section" style="border-left: 4px solid var(--info)">
                <div style="display:flex; justify-content:space-between; align-items:center">
                    <div>
                        <h3>活跃端口</h3>
                        <h1 style="margin:0; font-size:36px">${portStats?.data?.total_allocated || 0}</h1>
                        <p style="color:#666; margin:5px 0">端口池使用率: ${portStats?.data?.usage_percent || '0'}%</p>
                    </div>
                    <i class="fas fa-network-wired fa-3x" style="color:var(--info); opacity:0.2"></i>
                </div>
            </div>
            <div class="section" style="border-left: 4px solid var(--success)">
                <div style="display:flex; justify-content:space-between; align-items:center">
                    <div>
                        <h3>运行容器</h3>
                        <h1 style="margin:0; font-size:36px">${containerStats?.data?.total_instances || 0}</h1>
                        <p style="color:#666; margin:5px 0">系统状态: ${health?.status || 'Unknown'}</p>
                    </div>
                    <i class="fas fa-docker fa-3x" style="color:var(--success); opacity:0.2"></i>
                </div>
            </div>
        </div>

        <div class="grid-2">
            <div class="section">
                <h3><i class="fas fa-chart-line"></i> 快速操作</h3>
                <div class="button-group">
                    <button class="btn btn-primary" onclick="navigateTo('containers')">管理容器</button>
                    <button class="btn btn-info" onclick="navigateTo('logs')">查看日志</button>
                    <button class="btn btn-danger" onclick="navigateTo('malware')">上传样本</button>
                </div>
            </div>
            <div class="section">
                <h3><i class="fas fa-server"></i> 系统信息</h3>
                <div class="log-box" style="height:150px">${JSON.stringify(health, null, 2)}</div>
            </div>
        </div>
    `;
}

// 2. Containers
async function renderContainers(container) {
    container.innerHTML = `
        <div class="section">
            <h3><i class="fas fa-plus-circle"></i> 创建容器实例</h3>
            <div class="form-row">
                <div class="form-group">
                    <label>容器名称</label>
                    <input id="cc-name" type="text" placeholder="例如: honeypot-1">
                </div>
                <div class="form-group">
                    <label>选择镜像</label>
                    <select id="cc-image-select" onchange="toggleCustomImageInput()">
                        <option value="">-- 选择预设镜像 --</option>
                        <option value="andorralee/cowrie:v0.1">Cowrie (andorralee/cowrie:v0.1)</option>
                        <option value="andorralee/heralding:v0.1">Heralding (andorralee/heralding:v0.1)</option>
                        <option value="andorralee/mysql-preseed:8.0-arm64">MySQL (andorralee/mysql-preseed:8.0-arm64)</option>
                        <option value="nginx:latest">Nginx (Web)</option>
                        <option value="ubuntu:20.04">Ubuntu (Base)</option>
                        <option value="custom">自定义镜像...</option>
                    </select>
                </div>
                <div class="form-group" id="cc-custom-image-group" style="display:none">
                    <label>自定义镜像名</label>
                    <input id="cc-image-custom" type="text" placeholder="image:tag">
                </div>
                <div class="form-group">
                    <label>协议类型</label>
                    <select id="cc-proto">
                        <option value="http">HTTP</option>
                        <option value="ssh">SSH</option>
                        <option value="mysql">MySQL</option>
                        <option value="ftp">FTP</option>
                        <option value="telnet">Telnet</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>端口映射 (宿主机:容器)</label>
                    <input id="cc-ports" type="text" placeholder="80:80, 2222:22">
                </div>
                <div class="form-group" style="align-self: flex-end;">
                    <button class="btn btn-success" onclick="createContainer()"><i class="fas fa-check"></i> 创建</button>
                </div>
            </div>
        </div>

        <div class="section">
            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:15px">
                <h3><i class="fas fa-list"></i> 容器列表</h3>
                <div class="button-group">
                    <button class="btn btn-sm btn-outline" onclick="refreshContainerList()"><i class="fas fa-sync"></i> 刷新</button>
                    <button class="btn btn-sm btn-warning" onclick="syncContainerStatus()">同步状态</button>
                </div>
            </div>
            <div id="container-list" class="grid-3">加载中...</div>
        </div>

        <div class="section">
            <h3><i class="fas fa-images"></i> Docker 镜像管理</h3>
            <div class="form-row" style="margin-bottom:15px; background:#f9f9f9; padding:10px; border-radius:4px">
                <input id="pull-image-name" type="text" placeholder="输入镜像名称 (e.g. redis:alpine)">
                <button class="btn btn-primary" onclick="pullImage()">拉取镜像</button>
                <span style="color:#666; font-size:12px">或点击预设:</span>
                <button class="btn btn-sm btn-outline" onclick="quickPull('andorralee/cowrie:v0.1')">Cowrie</button>
                <button class="btn btn-sm btn-outline" onclick="quickPull('andorralee/heralding:v0.1')">Heralding</button>
            </div>
            <div id="image-list" class="grid-4"></div>
        </div>
    `;
    refreshContainerList();
    refreshImageList();
}

window.toggleCustomImageInput = () => {
    const val = $('#cc-image-select').value;
    $('#cc-custom-image-group').style.display = (val === 'custom') ? 'block' : 'none';
};

window.createContainer = async () => {
    const name = $('#cc-name').value.trim();
    const selectVal = $('#cc-image-select').value;
    const customVal = $('#cc-image-custom').value.trim();
    const image = (selectVal === 'custom') ? customVal : selectVal;
    const proto = $('#cc-proto').value;
    const portsRaw = $('#cc-ports').value.trim();

    if (!name || !image) return showToast('请填写名称和镜像', 'warning');

    const port_mappings = {};
    if (portsRaw) {
        portsRaw.split(/[,;]/).forEach(p => {
            const [host, container] = p.split(':').map(s => s.trim());
            if (host) port_mappings[host] = container || 'auto';
        });
    } else {
        // Defaults
        if (proto === 'http') port_mappings['80'] = '80';
        if (proto === 'ssh') port_mappings['2222'] = '22';
        if (proto === 'mysql') port_mappings['3306'] = '3306';
    }

    const payload = {
        name,
        honeypot_name: name,
        image_name: image,
        protocol: proto,
        interface_type: 'network',
        port_mappings,
        environment: { "TEST": "true" },
        description: 'Created from web UI'
    };

    const res = await apiCall('/temp-containers', 'POST', payload);
    if (res) {
        showToast('容器创建请求已发送', 'success');
        setTimeout(refreshContainerList, 1500);
    }
};

window.syncContainerStatus = async () => {
    await apiCall('/temp-containers/sync', 'POST'); // Try temp sync
    await apiCall('/container-instances/sync', 'POST'); // Try persistent sync
    showToast('状态同步请求已发送', 'info');
    setTimeout(refreshContainerList, 1000);
};

async function refreshContainerList() {
    const res = await apiCall('/temp-containers');
    const list = $('#container-list');
    if (!res || !res.data || res.data.length === 0) {
        list.innerHTML = '<div style="grid-column:1/-1; text-align:center; color:#999">暂无容器</div>';
        return;
    }
    list.innerHTML = res.data.map(c => `
        <div class="section" style="margin-bottom:0; padding:15px; border-left:4px solid ${c.status === 'running' ? 'var(--success)' : 'var(--danger)'}">
            <div style="display:flex; justify-content:space-between">
                <strong>${c.name}</strong>
                <span class="badge badge-${c.status === 'running' ? 'success' : 'danger'}">${c.status}</span>
            </div>
            <div style="font-size:12px; color:#666; margin:10px 0">
                <div>ID: ${c.id}</div>
                <div>DockerID: ${c.container_id ? c.container_id.substring(0,12) : '-'}</div>
                <div>Image: ${c.image_name}</div>
                <div>Proto: ${c.protocol}</div>
            </div>
            <div class="button-group">
                <button class="btn btn-sm btn-success" onclick="controlContainer('${c.id}', 'start')">启动</button>
                <button class="btn btn-sm btn-danger" onclick="controlContainer('${c.id}', 'stop')">停止</button>
                <button class="btn btn-sm btn-outline" onclick="deleteContainer('${c.id}')">删除</button>
                <button class="btn btn-sm btn-info" onclick="showDebugInfo('${c.id}')">调试</button>
                ${c.container_id ? `<button class=\"btn btn-sm btn-primary\" onclick=\"pullCowrieFor('${c.container_id}')\">拉取日志</button>` : ''}
            </div>
        </div>
    `).join('');
}

window.controlContainer = async (id, action) => {
    await apiCall(`/temp-containers/${id}/${action}`, 'POST');
    showToast(`发送 ${action} 指令`, 'success');
    setTimeout(refreshContainerList, 1000);
};

window.deleteContainer = async (id) => {
    if (confirm('确定删除?')) {
        await apiCall(`/temp-containers/${id}`, 'DELETE');
        showToast('已删除', 'success');
        refreshContainerList();
    }
};

window.showDebugInfo = async (id) => {
    const res = await apiCall(`/container-instances/${id}/debug`); // Try persistent endpoint first
    alert(JSON.stringify(res || {msg: "No debug info"}, null, 2));
};

window.pullCowrieFor = async (dockerId) => {
    await apiCall('/cowrie/pull-logs', 'POST', { container_id: dockerId });
    showToast('已触发Cowrie日志拉取', 'success');
    // 跳到日志页并预填容器ID
    navigateTo('logs');
    // 等视图渲染后设置值再查询
    setTimeout(() => {
        const sel = $('#log-type-select');
        if (sel) sel.value = 'cowrie';
        if ($('#log-cid')) $('#log-cid').value = dockerId;
        if ($('#log-cid-select')) $('#log-cid-select').value = dockerId;
        fetchLogs();
    }, 300);
};

async function refreshImageList() {
    const res = await apiCall('/docker/images');
    const list = $('#image-list');
    if (res && res.data) {
        list.innerHTML = res.data.map(img => {
            const tag = img.RepoTags ? img.RepoTags[0] : '<none>';
            const size = (img.Size / 1024 / 1024).toFixed(1);
            return `
            <div style="background:white; padding:10px; border:1px solid #eee; border-radius:4px">
                <div style="font-weight:bold; font-size:13px; word-break:break-all">${tag}</div>
                <div style="font-size:12px; color:#888; margin:5px 0">${size} MB | ${img.Id.substring(7,19)}</div>
                <button class="btn btn-sm btn-danger" onclick="deleteImage('${img.Id}')">删除</button>
            </div>
            `;
        }).join('');
    }
}

window.pullImage = async () => {
    const name = $('#pull-image-name').value.trim();
    if (!name) return;
    showToast('开始拉取镜像...', 'info');
    await apiCall('/docker/pull', 'POST', { image_name: name });
    showToast('拉取请求已发送', 'success');
};

window.quickPull = (name) => {
    $('#pull-image-name').value = name;
    window.pullImage();
};

window.deleteImage = async (id) => {
    if (confirm('删除镜像?')) {
        await apiCall(`/docker/images/${id}`, 'DELETE');
        refreshImageList();
    }
};


// 3. Logs
async function renderLogs(container) {
    container.innerHTML = `
        <div class="section">
            <h3><i class="fas fa-history"></i> 蜜罐日志审计</h3>
            <div class="form-row" style="margin-bottom:15px">
                <select id="log-type-select" onchange="fetchLogs()">
                    <option value="heralding">Heralding (凭证捕获)</option>
                    <option value="cowrie">Cowrie (SSH/Telnet)</option>
                    <option value="mysql">MySQL Honeypot</option>
                </select>
                <select id="log-cid-select" style="min-width:260px">
                    <option value="">选择运行容器 (自动填充Docker ID)</option>
                </select>
                <input id="log-cid" type="text" placeholder="或手动输入 Docker 容器ID (可选)">
                <button class="btn btn-primary" onclick="fetchLogs()">查询</button>
                <button class="btn btn-outline" onclick="pullLogs()">主动拉取最新日志</button>
            </div>
            <div id="log-display" class="log-viewer">选择日志类型并查询...</div>
        </div>
    `;
    populateContainerIdOptions();
    fetchLogs();
}

async function populateContainerIdOptions() {
    const sel = $('#log-cid-select');
    sel.innerHTML = '<option value="">选择运行容器 (自动填充Docker ID)</option>';
    // 先加载内存实例（/temp-containers）
    const tmp = await apiCall('/temp-containers');
    if (tmp && Array.isArray(tmp.data)) {
        tmp.data.forEach(c => {
            if (c.container_id) {
                const opt = document.createElement('option');
                const shortId = String(c.container_id).substring(0,12);
                opt.value = c.container_id;
                opt.textContent = `${c.name || 'unnamed'} | ${shortId} | ${c.image_name || ''}`;
                sel.appendChild(opt);
            }
        });
    }
    // 再尝试加载持久实例（/container-instances）
    const inst = await apiCall('/container-instances');
    if (inst && Array.isArray(inst.data)) {
        inst.data.forEach(c => {
            if (c.container_id) {
                const opt = document.createElement('option');
                const shortId = String(c.container_id).substring(0,12);
                opt.value = c.container_id;
                opt.textContent = `${c.name || 'instance'} | ${shortId} | ${c.image_name || ''}`;
                sel.appendChild(opt);
            }
        });
    }
    // 选择变化时，把值同步到文本输入，便于复制/手填
    sel.addEventListener('change', () => {
        const v = sel.value;
        if (v) $('#log-cid').value = v;
    });
}

window.fetchLogs = async () => {
    const type = $('#log-type-select').value;
    const cid = ($('#log-cid').value.trim()) || $('#log-cid-select').value;
    let endpoint = '';
    
    if (type === 'heralding') endpoint = cid ? `/heralding/logs/container/${cid}` : '/heralding/logs';
    else if (type === 'cowrie') endpoint = cid ? `/cowrie/logs/container/${cid}` : '/cowrie/logs';
    else if (type === 'mysql') endpoint = cid ? `/mysql-honeypot/logs/container/${cid}` : '/mysql-honeypot/logs';

    const res = await apiCall(endpoint);
    const display = $('#log-display');
    
    if (res && (res.logs || res.data)) {
        const logs = res.logs || res.data;
        if (logs.length === 0) {
            display.innerHTML = '<div style="padding:20px; text-align:center; color:#888">暂无日志数据</div>';
            return;
        }

        // 渲染表格
        let tableHtml = `
            <div style="overflow-x:auto">
                <table class="table table-striped">
                    <thead>
                        <tr>
                            <th>时间</th>
                            <th>来源IP</th>
                            <th>协议</th>
                            <th>事件/命令</th>
                            <th>详情</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        logs.forEach(log => {
            const time = new Date(log.event_time || log.timestamp || log.created_at).toLocaleString();
            const ip = log.source_ip || '-';
            const proto = log.protocol || type;
            
            let eventContent = '';
            let details = '';

            if (type === 'cowrie') {
                // Cowrie 特定展示
                if (log.command) {
                    eventContent = `<span class="badge badge-danger">CMD</span> <code>${escapeHtml(log.command)}</code>`;
                } else if (log.username || log.password) {
                    eventContent = `<span class="badge badge-warning">LOGIN</span> ${escapeHtml(log.username)} / ${escapeHtml(log.password)}`;
                } else {
                    eventContent = `<span class="badge badge-info">EVENT</span> ${escapeHtml(log.message || 'Connection')}`;
                }
                details = `Session: ${log.session_id ? log.session_id.substring(0,8) : '-'}`;
            } else if (type === 'heralding') {
                // Heralding 特定展示
                eventContent = `<span class="badge badge-warning">AUTH</span> ${escapeHtml(log.username)}:${escapeHtml(log.password)}`;
                details = `AuthID: ${log.auth_id ? log.auth_id.substring(0,8) : '-'}`;
            } else if (type === 'mysql') {
                // MySQL 特定展示
                eventContent = log.query ? `<code>${escapeHtml(log.query)}</code>` : 'Login Attempt';
                details = `User: ${log.username}`;
            }

            tableHtml += `
                <tr>
                    <td style="font-size:12px; white-space:nowrap">${time}</td>
                    <td>${ip}</td>
                    <td>${proto}</td>
                    <td>${eventContent}</td>
                    <td style="font-size:12px; color:#666">${details}</td>
                </tr>
            `;
        });

        tableHtml += `</tbody></table></div>`;
        display.innerHTML = tableHtml;
    } else {
        display.innerHTML = '<div style="padding:20px; text-align:center; color:#888">查询返回空</div>';
    }
};

function escapeHtml(text) {
    if (!text) return '';
    return text
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

window.pullLogs = async () => {
    const type = $('#log-type-select').value;
    const cid = ($('#log-cid').value.trim()) || $('#log-cid-select').value;
    // 后端拉取接口需要 container_id，缺失会 400
    if (!cid) {
        showToast('请选择或输入 Docker 容器ID 后再拉取', 'warning');
        return;
    }
    const endpoint = {
        'heralding': '/heralding/pull-logs',
        'cowrie': '/cowrie/pull-logs',
        'mysql': '/mysql-honeypot/pull-logs'
    }[type];
    
    await apiCall(endpoint, 'POST', {container_id: cid});
    showToast('日志拉取任务已触发', 'success');
    setTimeout(fetchLogs, 2000);
};


// 4. Malware
async function renderMalware(container) {
    container.innerHTML = `
        <div class="grid-2">
            <div class="section">
                <h3><i class="fas fa-upload"></i> 样本上传扫描</h3>
                <div class="form-group">
                    <input type="file" id="malware-file">
                </div>
                <button class="btn btn-danger" onclick="scanFile()">上传并扫描</button>
                <div id="scan-result" class="log-box" style="margin-top:15px; height:200px; display:none"></div>
            </div>
            <div class="section">
                <h3><i class="fas fa-database"></i> 病毒特征库</h3>
                <div class="button-group">
                    <button class="btn btn-sm btn-outline" onclick="loadSignatures()">刷新特征列表</button>
                    <button class="btn btn-sm btn-primary" onclick="importSignatures()">导入数据集</button>
                </div>
                <div id="sig-list" class="log-box" style="height:200px"></div>
            </div>
        </div>
        <div class="section">
            <h3><i class="fas fa-list-alt"></i> 扫描历史</h3>
            <div style="overflow-x:auto">
                <table class="table">
                    <thead><tr><th>Hash</th><th>文件名</th><th>结果</th><th>时间</th></tr></thead>
                    <tbody id="scan-history-body"></tbody>
                </table>
            </div>
        </div>
    `;
    refreshScanHistory();
    loadSignatures();
}

window.scanFile = async () => {
    const file = $('#malware-file').files[0];
    if (!file) return showToast('请选择文件', 'warning');
    
    const formData = new FormData();
    formData.append('file', file);
    
    $('#scan-result').style.display = 'block';
    $('#scan-result').textContent = '正在上传并扫描...';
    
    try {
        const res = await fetch(API_BASE + '/malware/scan/file', { method: 'POST', body: formData });
        const data = await res.json();
        
        if (data.code === 0 && data.data) {
            const result = data.data.scan_result || data.data;
            const infected = result.is_infected ? '已感染' : '干净';
            const level = result.threat_level || 'unknown';
            const detCount = result.detections ? result.detections.length : (result.detection_count || 0);
            
            $('#scan-result').innerHTML = `
                <div style="padding:15px">
                    <div style="margin-bottom:10px"><strong>文件:</strong> ${result.file_name || file.name}</div>
                    <div style="margin-bottom:10px"><strong>状态:</strong> <span class="badge badge-${result.is_infected ? 'danger' : 'success'}">${infected}</span></div>
                    <div style="margin-bottom:10px"><strong>威胁等级:</strong> ${level}</div>
                    <div style="margin-bottom:10px"><strong>检测数:</strong> ${detCount}</div>
                    <details style="margin-top:10px">
                        <summary style="cursor:pointer">查看原始响应</summary>
                        <pre style="margin-top:10px; font-size:11px">${JSON.stringify(data, null, 2)}</pre>
                    </details>
                </div>
            `;
        } else {
            $('#scan-result').textContent = JSON.stringify(data, null, 2);
        }
        
        // 延迟刷新以确保后端完成写入
        setTimeout(refreshScanHistory, 500);
    } catch(e) {
        $('#scan-result').textContent = 'Error: ' + e.message;
        showToast('扫描失败: ' + e.message, 'error');
    }
};

async function refreshScanHistory() {
    const res = await apiCall('/malware/scan/history');
    const tbody = $('#scan-history-body');
    if (!tbody) return;
    
    if (res && res.scan_history && res.scan_history.length > 0) {
        tbody.innerHTML = res.scan_history.map(h => `
            <tr>
                <td style="font-family:monospace; font-size:12px">${h.file_hash ? h.file_hash.substring(0,12)+'...' : (h.sha256 ? h.sha256.substring(0,12)+'...' : '-')}</td>
                <td>${h.file_name || '-'}</td>
                <td><span class="badge badge-${h.is_infected ? 'danger':'success'}">${h.is_infected ? 'Infected':'Clean'}</span></td>
                <td>${new Date(h.scan_time).toLocaleString() || '-'}</td>
            </tr>
        `).join('');
    } else {
        tbody.innerHTML = '<tr><td colspan="4" style="text-align:center; color:#999">暂无扫描记录</td></tr>';
    }
}

window.loadSignatures = async () => {
    const res = await apiCall('/malware/signatures');
    $('#sig-list').textContent = JSON.stringify(res, null, 2);
};

window.importSignatures = async () => {
    if(confirm('导入默认数据集签名?')) {
        await apiCall('/malware/signatures/import', 'POST');
        showToast('导入完成', 'success');
        loadSignatures();
    }
};


// 5. Network
async function renderNetwork(container) {
    container.innerHTML = `
        <div class="grid-2">
            <div class="section">
                <h3><i class="fas fa-search"></i> 端口检查</h3>
                <div class="form-row">
                    <input type="number" id="check-port" placeholder="端口号 (e.g. 80)">
                    <button class="btn btn-info" onclick="checkPort()">检查可用性</button>
                </div>
                <div id="port-check-res" style="margin-top:10px"></div>
            </div>
            <div class="section">
                <h3><i class="fas fa-plus"></i> 手动分配</h3>
                <div class="form-group">
                    <input type="text" id="alloc-cid" placeholder="容器ID">
                </div>
                <div class="form-row">
                    <select id="alloc-svc">
                        <option value="http">HTTP</option>
                        <option value="ssh">SSH</option>
                        <option value="mysql">MySQL</option>
                    </select>
                    <button class="btn btn-success" onclick="allocatePort()">分配</button>
                </div>
            </div>
        </div>
        <div class="section">
            <h3><i class="fas fa-project-diagram"></i> 已分配端口列表</h3>
            <button class="btn btn-sm btn-outline" onclick="refreshPorts()" style="margin-bottom:10px">刷新</button>
            <div id="port-list" class="grid-4"></div>
        </div>
    `;
    refreshPorts();
}

window.checkPort = async () => {
    const p = $('#check-port').value;
    if(!p) return;
    const res = await apiCall(`/ports/${p}/check`);
    $('#port-check-res').innerHTML = res ? 
        `<span class="badge badge-${res.available ? 'success':'danger'}">${res.message}</span>` : '';
};

window.allocatePort = async () => {
    const cid = $('#alloc-cid').value;
    const svc = $('#alloc-svc').value;
    if(!cid) return showToast('需填写容器ID', 'warning');
    await apiCall('/ports/allocate', 'POST', { container_id: cid, service_type: svc, description: 'Manual' });
    showToast('分配成功', 'success');
    refreshPorts();
};

async function refreshPorts() {
    const res = await apiCall('/ports/allocated');
    const list = $('#port-list');
    if(res && res.data) {
        list.innerHTML = res.data.map(p => `
            <div style="background:white; border:1px solid #eee; padding:10px; border-radius:4px; border-top:3px solid var(--info)">
                <div style="font-size:18px; font-weight:bold">${p.port}</div>
                <div style="color:#666; font-size:12px">${p.service_type}</div>
                <div style="color:#999; font-size:12px; margin-bottom:5px">CID: ${p.container_id}</div>
                <button class="btn btn-sm btn-danger" onclick="releasePort(${p.port})">释放</button>
            </div>
        `).join('');
    }
}

window.releasePort = async (p) => {
    if(confirm(`释放端口 ${p}?`)) {
        await apiCall(`/ports/${p}/release`, 'DELETE');
        refreshPorts();
    }
};


// 6. Threats
async function renderThreats(container) {
    container.innerHTML = `
        <div class="section">
            <h3><i class="fas fa-user-secret"></i> 攻击会话 (Sessions)</h3>
            <button class="btn btn-sm btn-outline" onclick="loadSessions()">刷新会话</button>
            <div id="session-list" class="log-box" style="margin-top:10px; height:200px"></div>
        </div>
        <div class="section">
            <h3><i class="fas fa-key"></i> 蜜签 (Honeytokens)</h3>
            <div class="button-group">
                <button class="btn btn-sm btn-primary" onclick="createHoneytoken()">生成新蜜签</button>
                <button class="btn btn-sm btn-outline" onclick="loadHoneytokens()">刷新列表</button>
            </div>
            <div id="token-list" class="log-box" style="margin-top:10px; height:200px"></div>
        </div>
    `;
    loadSessions();
    loadHoneytokens();
}

window.loadSessions = async () => {
    // Try to get recent sessions or stats
    const res = await apiCall('/sessions/statistics'); 
    $('#session-list').textContent = JSON.stringify(res, null, 2);
};

window.loadHoneytokens = async () => {
    const res = await apiCall('/honeytokens');
    $('#token-list').textContent = JSON.stringify(res, null, 2);
};

window.createHoneytoken = async () => {
    const type = prompt("蜜签类型 (file/db/url):", "file");
    if(type) {
        await apiCall('/honeytokens', 'POST', { type, value: "test-token-"+Date.now() });
        showToast('蜜签已创建', 'success');
        loadHoneytokens();
    }
};


// 7. System
async function renderSystem(container) {
    const health = await apiCall('/health');
    container.innerHTML = `
        <div class="grid-2">
            <div class="section">
                <h3><i class="fas fa-heartbeat"></i> 健康状态</h3>
                <pre class="log-box">${JSON.stringify(health, null, 2)}</pre>
            </div>
            <div class="section">
                <h3><i class="fas fa-gavel"></i> 安全规则</h3>
                <button class="btn btn-sm btn-outline" onclick="loadRules()">加载规则</button>
                <div id="rule-list" class="log-box" style="margin-top:10px; height:200px"></div>
            </div>
        </div>
    `;
    loadRules();
}

window.loadRules = async () => {
    const res = await apiCall('/rules');
    $('#rule-list').textContent = JSON.stringify(res, null, 2);
};


// --- Init ---
document.addEventListener('DOMContentLoaded', () => {
    // Nav handlers
    $$('.nav-item').forEach(link => {
        link.addEventListener('click', () => navigateTo(link.dataset.route));
    });
    
    $('#refresh-btn').addEventListener('click', () => {
        const active = $('.nav-item.active');
        if(active) navigateTo(active.dataset.route);
    });

    // Start
    navigateTo('dashboard');
    setInterval(updateTime, 1000);
    updateTime();
    
    // Check system status
    apiCall('/health').then(res => {
        const el = $('#system-status');
        if(res && res.status === 'ok') {
            el.className = 'badge badge-success';
            el.textContent = '系统在线';
        } else {
            el.className = 'badge badge-danger';
            el.textContent = '异常';
        }
    });
});
