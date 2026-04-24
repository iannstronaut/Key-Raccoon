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
  endpoint?: string;
  description?: string;
  is_active: boolean;
  budget: number;
  budget_used: number;
  models?: Model[];
  api_keys?: ChannelAPIKey[];
}

export interface ChannelAPIKey {
  id: number;
  channel_id: number;
  api_key: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Model {
  id: number;
  name: string;
  display_name: string;
  channel_id?: number;
  is_active?: boolean;
  token_price?: number;
  system_prompt?: string;
  created_at?: string;
  updated_at?: string;
}

export interface UserAPIKey {
  id: number;
  user_id: number;
  name: string;
  key: string;
  is_active: boolean;
  usage_limit: number;
  usage_count: number;
  token_limit?: number;
  token_used?: number;
  expires_at?: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
  user?: User;
  channels?: Channel[];
  models?: Model[];
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

export interface RequestLog {
  id: number;
  user_api_key_id: number;
  user_id: number;
  channel_id: number;
  model_name: string;
  channel_name: string;
  user_email: string;
  token_price: number;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  cost: number;
  status: string;
  error_message?: string;
  latency_ms: number;
  request_ip: string;
  created_at: string;
}

export interface UsageStats {
  total_requests: number;
  total_tokens: number;
  total_cost: number;
  success_count: number;
  failed_count: number;
  avg_latency_ms: number;
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
