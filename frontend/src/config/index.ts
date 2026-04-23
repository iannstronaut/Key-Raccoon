/**
 * Application configuration
 * Reads from environment variables with fallback defaults
 */

interface AppConfig {
  apiBaseUrl: string;
  isDevelopment: boolean;
}

/**
 * Get the API base URL from environment or use default
 */
function getApiBaseUrl(): string {
  // Check Vite environment variable
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL;
  }

  // Check if running in development mode with proxy
  if (import.meta.env.DEV) {
    // In development, use relative path to leverage Vite proxy
    return '';
  }

  // Production fallback - try to detect from window location
  if (typeof window !== 'undefined') {
    const { protocol, hostname, port } = window.location;
    // If frontend is served from same origin as backend
    if (port === '8080' || port === '3000') {
      return `${protocol}//${hostname}:${port}`;
    }
    // Default production backend port
    return `${protocol}//${hostname}:8080`;
  }

  // Final fallback
  return 'http://localhost:8080';
}

/**
 * Application configuration object
 */
export const config: AppConfig = {
  apiBaseUrl: getApiBaseUrl(),
  isDevelopment: import.meta.env.DEV || import.meta.env.VITE_DEV_MODE === 'true',
};

/**
 * Log configuration on startup (only in development)
 */
if (config.isDevelopment) {
  console.log('🔧 App Configuration:', {
    apiBaseUrl: config.apiBaseUrl,
    isDevelopment: config.isDevelopment,
    env: import.meta.env.MODE,
  });
}

export default config;
