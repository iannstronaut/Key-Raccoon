import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api } from '../services/api';

export type UserRole = 'superadmin' | 'admin' | 'user';

interface User {
  id: number;
  email: string;
  name: string;
  role: UserRole;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isSuperAdmin: boolean;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    // Load user from token on mount
    const token = api.getToken();
    if (token) {
      // Decode JWT to get user info
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        setUser({
          id: payload.user_id,
          email: payload.email,
          name: payload.name || payload.email,
          role: payload.role || 'user',
        });
      } catch (error) {
        console.error('Failed to decode token:', error);
        api.clearToken();
      }
    }
  }, []);

  const login = async (email: string, password: string): Promise<boolean> => {
    try {
      const response = await api.login(email, password);
      if (response?.access_token) {
        api.setToken(response.access_token);
        
        // Decode token to get user info
        const payload = JSON.parse(atob(response.access_token.split('.')[1]));
        setUser({
          id: payload.user_id,
          email: payload.email,
          name: payload.name || payload.email,
          role: payload.role || 'user',
        });
        
        return true;
      }
      return false;
    } catch (error) {
      console.error('Login failed:', error);
      return false;
    }
  };

  const logout = () => {
    api.clearToken();
    setUser(null);
  };

  const isAdmin = user?.role === 'admin' || user?.role === 'superadmin';
  const isSuperAdmin = user?.role === 'superadmin';

  const hasPermission = (permission: string): boolean => {
    if (!user) return false;

    switch (permission) {
      case 'view:dashboard':
        return true; // All authenticated users
      case 'view:channels':
        return true; // All authenticated users
      case 'edit:channels':
        return isAdmin;
      case 'delete:channels':
        return isAdmin;
      case 'view:users':
        return isAdmin;
      case 'edit:users':
        return isAdmin;
      case 'view:proxies':
        return isAdmin;
      case 'edit:proxies':
        return isAdmin;
      case 'view:analytics':
        return true; // All authenticated users
      case 'view:logs':
        return isAdmin;
      case 'view:my-channels':
        return true; // All authenticated users
      case 'create:own-api-keys':
        return true; // All authenticated users
      default:
        return false;
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isAdmin,
        isSuperAdmin,
        login,
        logout,
        hasPermission,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
