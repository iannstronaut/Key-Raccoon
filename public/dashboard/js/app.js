// Check authentication
if (!API.getToken()) {
	window.location.href = '/login';
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
	setupNavigation();
	loadDashboardStats();
	loadHealthStatus();
	setupAutoRefresh();
});

function setupNavigation() {
	const navLinks = document.querySelectorAll('.sidebar nav a[data-page]');
	navLinks.forEach(link => {
		link.addEventListener('click', (e) => {
			e.preventDefault();
			const pageName = link.dataset.page;
			showPage(pageName);

			// Update active state
			navLinks.forEach(l => l.classList.remove('active'));
			link.classList.add('active');

			// Update page title
			const title = link.textContent.trim();
			document.getElementById('page-title').textContent = title;
		});
	});
}

function showPage(pageName) {
	document.querySelectorAll('.page').forEach(page => {
		page.classList.remove('active');
	});

	const page = document.getElementById(`${pageName}-page`);
	if (page) {
		page.classList.add('active');
		if (pageName === 'users') loadUsers();
		if (pageName === 'channels') loadChannels();
		if (pageName === 'proxies') loadProxies();
	}
}

/* ============================================
   Dashboard Stats
   ============================================ */

async function loadDashboardStats() {
	try {
		const [users, channels, proxies] = await Promise.all([
			API.getUsers(1),
			API.getChannels(1),
			API.getProxies(1),
		]);

		const usersEl = document.getElementById('total-users');
		const channelsEl = document.getElementById('total-channels');
		const proxiesEl = document.getElementById('active-proxies');

		if (usersEl) usersEl.textContent = users?.total ?? (Array.isArray(users) ? users.length : 0);
		if (channelsEl) channelsEl.textContent = channels?.total ?? (Array.isArray(channels) ? channels.length : 0);

		let activeProxies = 0;
		if (proxies?.data) {
			activeProxies = proxies.data.filter(p => p.status === 'healthy').length;
		} else if (Array.isArray(proxies)) {
			activeProxies = proxies.filter(p => p.status === 'healthy').length;
		}
		if (proxiesEl) proxiesEl.textContent = activeProxies;
	} catch (err) {
		console.error('Failed to load dashboard stats:', err);
	}
}

async function loadHealthStatus() {
	try {
		const health = await API.getHealth();
		const dbStatus = document.getElementById('db-status');
		const redisStatus = document.getElementById('redis-status');
		const healthBadge = document.getElementById('health-badge');

		if (dbStatus) {
			const dbOk = health?.database_ok;
			dbStatus.textContent = dbOk ? 'Connected' : 'Disconnected';
			dbStatus.className = 'health-status ' + (dbOk ? 'healthy' : 'unhealthy');
		}

		if (redisStatus) {
			const redisOk = health?.redis_ok;
			redisStatus.textContent = redisOk ? 'Connected' : 'Disconnected';
			redisStatus.className = 'health-status ' + (redisOk ? 'healthy' : 'unhealthy');
		}

		if (healthBadge) {
			const allOk = health?.database_ok && health?.redis_ok;
			healthBadge.textContent = allOk ? 'Healthy' : 'Degraded';
			healthBadge.className = 'badge ' + (allOk ? 'badge-success' : 'badge-warning');
		}
	} catch (err) {
		console.error('Failed to load health status:', err);
	}
}

/* ============================================
   Users
   ============================================ */

async function loadUsers() {
	const response = await API.getUsers();
	const tbody = document.getElementById('users-tbody');
	if (!tbody) return;
	tbody.innerHTML = '';

	const users = response?.users ?? response ?? [];
	if (!Array.isArray(users) || users.length === 0) {
		tbody.innerHTML = '<tr><td colspan="6" class="placeholder-text" style="text-align:center;padding:32px;">No users found</td></tr>';
		return;
	}

	users.forEach(user => {
		const row = document.createElement('tr');
		const lastLogin = user.last_login ? new Date(user.last_login).toLocaleString() : 'Never';
		const statusBadge = user.is_active
			? '<span class="badge badge-success">Active</span>'
			: '<span class="badge badge-danger">Inactive</span>';

		row.innerHTML = `
			<td>${escapeHtml(user.email)}</td>
			<td>${escapeHtml(user.name || '')}</td>
			<td><span class="badge">${escapeHtml(user.role)}</span></td>
			<td>${statusBadge}</td>
			<td>${escapeHtml(lastLogin)}</td>
			<td>
				<button onclick="editUser(${user.id})" class="btn-small">Edit</button>
				<button onclick="deleteUser(${user.id})" class="btn-small btn-danger">Delete</button>
			</td>
		`;
		tbody.appendChild(row);
	});
}

async function createUser(event) {
	event.preventDefault();
	const email = document.getElementById('user-email').value;
	const name = document.getElementById('user-name').value;
	const password = document.getElementById('user-password').value;
	const role = document.getElementById('user-role').value;

	const response = await API.createUser(email, password, name, role);
	if (response?.id || response?.email) {
		closeModal('create-user-modal');
		event.target.reset();
		loadUsers();
		loadDashboardStats();
	} else {
		alert('Error: ' + (response?.error || 'Failed to create user'));
	}
}

async function deleteUser(id) {
	if (!confirm('Are you sure you want to delete this user?')) return;
	const response = await API.deleteUser(id);
	if (response?.error) {
		alert('Error: ' + response.error);
	} else {
		loadUsers();
		loadDashboardStats();
	}
}

function editUser(id) {
	alert('Edit user ' + id + ' — implementation pending');
}

/* ============================================
   Channels
   ============================================ */

async function loadChannels() {
	const response = await API.getChannels();
	const tbody = document.getElementById('channels-tbody');
	if (!tbody) return;
	tbody.innerHTML = '';

	const channels = response?.channels ?? response ?? [];
	if (!Array.isArray(channels) || channels.length === 0) {
		tbody.innerHTML = '<tr><td colspan="6" class="placeholder-text" style="text-align:center;padding:32px;">No channels found</td></tr>';
		return;
	}

	channels.forEach(channel => {
		const row = document.createElement('tr');
		const statusBadge = channel.is_active
			? '<span class="badge badge-success">Active</span>'
			: '<span class="badge badge-danger">Inactive</span>';

		row.innerHTML = `
			<td>${escapeHtml(channel.name)}</td>
			<td><span class="badge">${escapeHtml(channel.type)}</span></td>
			<td>${statusBadge}</td>
			<td>${channel.models?.length || 0}</td>
			<td>${channel.api_keys?.length || 0}</td>
			<td>
				<button onclick="editChannel(${channel.id})" class="btn-small">Edit</button>
				<button onclick="deleteChannel(${channel.id})" class="btn-small btn-danger">Delete</button>
			</td>
		`;
		tbody.appendChild(row);
	});
}

async function createChannel(event) {
	event.preventDefault();
	const name = document.getElementById('channel-name').value;
	const type = document.getElementById('channel-type').value;
	const description = document.getElementById('channel-description').value;

	const response = await API.createChannel(name, type, description);
	if (response?.id || response?.name) {
		closeModal('create-channel-modal');
		event.target.reset();
		loadChannels();
		loadDashboardStats();
	} else {
		alert('Error: ' + (response?.error || 'Failed to create channel'));
	}
}

async function deleteChannel(id) {
	if (!confirm('Are you sure you want to delete this channel?')) return;
	const response = await API.deleteChannel(id);
	if (response?.error) {
		alert('Error: ' + response.error);
	} else {
		loadChannels();
		loadDashboardStats();
	}
}

function editChannel(id) {
	alert('Edit channel ' + id + ' — implementation pending');
}

/* ============================================
   Proxies
   ============================================ */

async function loadProxies() {
	const response = await API.getProxies();
	const tbody = document.getElementById('proxies-tbody');
	if (!tbody) return;
	tbody.innerHTML = '';

	const proxies = response?.proxies ?? response ?? [];
	if (!Array.isArray(proxies) || proxies.length === 0) {
		tbody.innerHTML = '<tr><td colspan="5" class="placeholder-text" style="text-align:center;padding:32px;">No proxies found</td></tr>';
		return;
	}

	proxies.forEach(proxy => {
		const row = document.createElement('tr');
		const lastCheck = proxy.last_check ? new Date(proxy.last_check).toLocaleString() : 'Never';
		const statusClass = proxy.status === 'healthy' ? 'badge-success' : proxy.status === 'unhealthy' ? 'badge-danger' : 'badge-warning';

		row.innerHTML = `
			<td>${escapeHtml(proxy.url)}</td>
			<td><span class="badge">${escapeHtml(proxy.type)}</span></td>
			<td><span class="badge ${statusClass}">${escapeHtml(proxy.status)}</span></td>
			<td>${escapeHtml(lastCheck)}</td>
			<td>
				<button onclick="checkProxy(${proxy.id})" class="btn-small">Check</button>
				<button onclick="deleteProxy(${proxy.id})" class="btn-small btn-danger">Delete</button>
			</td>
		`;
		tbody.appendChild(row);
	});
}

async function addProxy(event) {
	event.preventDefault();
	const url = document.getElementById('proxy-url').value;
	const type = document.getElementById('proxy-type').value;
	const username = document.getElementById('proxy-username').value;
	const password = document.getElementById('proxy-password').value;

	const response = await API.addProxy(url, type, username || null, password || null);
	if (response?.id || response?.url) {
		closeModal('add-proxy-modal');
		event.target.reset();
		loadProxies();
		loadDashboardStats();
	} else {
		alert('Error: ' + (response?.error || 'Failed to add proxy'));
	}
}

async function deleteProxy(id) {
	if (!confirm('Are you sure you want to delete this proxy?')) return;
	const response = await API.deleteProxy(id);
	if (response?.error) {
		alert('Error: ' + response.error);
	} else {
		loadProxies();
		loadDashboardStats();
	}
}

function checkProxy(id) {
	alert('Check proxy ' + id + ' — implementation pending');
}

/* ============================================
   Modal Helpers
   ============================================ */

function openCreateUserModal() {
	document.getElementById('create-user-modal').classList.add('active');
}

function openCreateChannelModal() {
	document.getElementById('create-channel-modal').classList.add('active');
}

function openAddProxyModal() {
	document.getElementById('add-proxy-modal').classList.add('active');
}

function closeModal(modalId) {
	document.getElementById(modalId).classList.remove('active');
}

// Close modal on backdrop click
document.addEventListener('click', (e) => {
	if (e.target.classList.contains('modal')) {
		e.target.classList.remove('active');
	}
});

/* ============================================
   Utilities
   ============================================ */

function logout() {
	API.clearToken();
	window.location.href = '/login';
}

function setupAutoRefresh() {
	setInterval(() => {
		loadDashboardStats();
		loadHealthStatus();
	}, 30000);
}

function escapeHtml(text) {
	if (text == null) return '';
	const div = document.createElement('div');
	div.textContent = text;
	return div.innerHTML;
}
