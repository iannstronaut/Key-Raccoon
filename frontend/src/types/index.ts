export interface User {
  id: number;
  email: string;
  name: string;
  role: string;
  is_active: boolean;
  last_login?: string;
}

export interface Channel {
  id: number;
  name: string;
  type: string;
  description?: string;
  is_active: boolean;
  models?: string[];
  api_keys?: string[];
}

export interface Proxy {
  id: number;
  url: string;
  type: string;
  status: string;
  username?: string;
  last_check?: string;
}

export interface HealthStatus {
  database_ok: boolean;
  redis_ok: boolean;
}

export interface PaginatedResponse<T> {
  data?: T[];
  users?: T[];
  channels?: T[];
  proxies?: T[];
  total?: number;
}

export interface ApiError {
  error: string;
}
