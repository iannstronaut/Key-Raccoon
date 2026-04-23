import config from '../config';

class APIService {
  private token: string | null = localStorage.getItem("auth_token");
  private apiBase: string;

  constructor() {
    // Use config API base URL + /api path
    this.apiBase = config.apiBaseUrl ? `${config.apiBaseUrl}/api` : '/api';
  }

  setToken(token: string) {
    this.token = token;
    localStorage.setItem("auth_token", token);
  }

  getToken(): string | null {
    return this.token;
  }

  clearToken() {
    this.token = null;
    localStorage.removeItem("auth_token");
  }

  isAuthenticated(): boolean {
    return !!this.token;
  }

  private async request<T>(
    method: string,
    endpoint: string,
    data?: unknown,
  ): Promise<T | null> {
    const url = `${this.apiBase}${endpoint}`;
    const options: RequestInit = {
      method,
      headers: {
        "Content-Type": "application/json",
      },
    };

    if (this.token) {
      options.headers = {
        ...options.headers,
        Authorization: `Bearer ${this.token}`,
      };
    }

    if (data) {
      options.body = JSON.stringify(data);
    }

    try {
      const response = await fetch(url, options);

      if (response.status === 401) {
        this.clearToken();
        window.location.href = "/login";
        return null;
      }

      const contentType = response.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        return (await response.json()) as T;
      }
      return { status: response.status, ok: response.ok } as unknown as T;
    } catch (error) {
      console.error("Request error:", error);
      return { error: "Network error" } as unknown as T;
    }
  }

  // Auth
  async login(email: string, password: string) {
    return this.request<{ access_token?: string; error?: string }>(
      "POST",
      "/auth/login",
      { email, password },
    );
  }

  // Users
  async getUsers(limit = 50, offset = 0) {
    return this.request<{ users: unknown[]; total: number }>(
      "GET",
      `/users?limit=${limit}&offset=${offset}`,
    );
  }

  async createUser(user: {
    email: string;
    password: string;
    name: string;
    role: string;
  }) {
    return this.request<unknown>("POST", "/users", user);
  }

  async updateUser(id: number, data: Partial<unknown>) {
    return this.request<unknown>("PUT", `/users/${id}`, data);
  }

  async deleteUser(id: number) {
    return this.request<unknown>("DELETE", `/users/${id}`);
  }

  // Channels
  async getChannels(limit = 50, offset = 0) {
    return this.request<{ channels: unknown[]; total: number }>(
      "GET",
      `/channels?limit=${limit}&offset=${offset}`,
    );
  }

  async getChannel(id: number) {
    return this.request<unknown>("GET", `/channels/${id}`);
  }

  async createChannel(channel: {
    name: string;
    type: string;
    description?: string;
  }) {
    return this.request<unknown>("POST", "/channels", channel);
  }

  async deleteChannel(id: number) {
    return this.request<unknown>("DELETE", `/channels/${id}`);
  }

  // Channel API Keys
  async addChannelAPIKey(channelId: number, apiKey: string) {
    return this.request<unknown>("POST", `/channels/${channelId}/api-keys`, {
      api_key: apiKey,
    });
  }

  async getChannelAPIKeys(channelId: number) {
    return this.request<unknown>("GET", `/channels/${channelId}/api-keys`);
  }

  async deleteChannelAPIKey(channelId: number, keyId: number) {
    return this.request<unknown>("DELETE", `/channels/${channelId}/api-keys/${keyId}`);
  }

  // Channel Models
  async addChannelModel(channelId: number, model: {
    name: string;
    display_name?: string;
    token_price?: number;
    system_prompt?: string;
  }) {
    return this.request<unknown>("POST", `/channels/${channelId}/models`, model);
  }

  async getChannelModels(channelId: number) {
    return this.request<unknown>("GET", `/channels/${channelId}/models`);
  }

  async deleteChannelModel(channelId: number, modelId: number) {
    return this.request<unknown>("DELETE", `/channels/${channelId}/models/${modelId}`);
  }

  // Proxies
  async getProxies(limit = 50, offset = 0) {
    return this.request<{ proxies: unknown[]; total: number }>(
      "GET",
      `/proxies?limit=${limit}&offset=${offset}`,
    );
  }

  async addProxy(proxy: {
    proxy_url: string;
    type: string;
    username?: string;
    password?: string;
  }) {
    return this.request<unknown>("POST", "/proxies", proxy);
  }

  async deleteProxy(id: number) {
    return this.request<unknown>("DELETE", `/proxies/${id}`);
  }

  // Health
  async getHealth() {
    const healthUrl = config.apiBaseUrl ? `${config.apiBaseUrl}/health` : '/health';
    return fetch(healthUrl, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    }).then((response) => response.json());
  }
}

export const api = new APIService();
