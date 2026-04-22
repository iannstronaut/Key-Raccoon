const API_BASE = '/api';

let authToken = localStorage.getItem('auth_token');

class API {
	static setToken(token) {
		authToken = token;
		localStorage.setItem('auth_token', token);
	}

	static getToken() {
		return authToken;
	}

	static clearToken() {
		authToken = null;
		localStorage.removeItem('auth_token');
	}

	static async request(method, endpoint, data = null) {
		const url = `${API_BASE}${endpoint}`;
		const options = {
			method,
			headers: {
				'Content-Type': 'application/json',
			},
		};

		if (authToken) {
			options.headers['Authorization'] = `Bearer ${authToken}`;
		}

		if (data) {
			options.body = JSON.stringify(data);
		}

		try {
			const response = await fetch(url, options);

			if (response.status === 401) {
				API.clearToken();
				window.location.href = '/login';
				return null;
			}

			const contentType = response.headers.get('content-type');
			if (contentType && contentType.includes('application/json')) {
				return await response.json();
			}
			return { status: response.status, ok: response.ok };
		} catch (error) {
			console.error('Request error:', error);
			return { error: 'Network error' };
		}
	}

	// Users
	static async getUsers(limit = 50, offset = 0) {
		return this.request('GET', `/users?limit=${limit}&offset=${offset}`);
	}

	static async createUser(email, password, name, role) {
		return this.request('POST', '/users', {
			email, password, name, role
		});
	}

	static async updateUser(id, data) {
		return this.request('PUT', `/users/${id}`, data);
	}

	static async deleteUser(id) {
		return this.request('DELETE', `/users/${id}`);
	}

	// Channels
	static async getChannels(limit = 50, offset = 0) {
		return this.request('GET', `/channels?limit=${limit}&offset=${offset}`);
	}

	static async createChannel(name, type, description) {
		return this.request('POST', '/channels', {
			name, type, description
		});
	}

	static async deleteChannel(id) {
		return this.request('DELETE', `/channels/${id}`);
	}

	// Proxies
	static async getProxies(limit = 50, offset = 0) {
		return this.request('GET', `/proxies?limit=${limit}&offset=${offset}`);
	}

	static async addProxy(proxyUrl, type, username, password) {
		return this.request('POST', '/proxies', {
			proxy_url: proxyUrl,
			type,
			username,
			password
		});
	}

	static async deleteProxy(id) {
		return this.request('DELETE', `/proxies/${id}`);
	}

	// Health
	static async getHealth() {
		return this.request('GET', '/health');
	}
}
