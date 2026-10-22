// src/context/AuthContext.tsx
import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { clearToken, saveToken } from '../services/api';

// 1. Tipos de rol que maneja la aplicación
export type Rol = 'admin' | 'cliente';

// 2. Usuario autenticado: correo + rol (lo entrega la API en /api/login)
export interface Usuario {
  email: string;
  rol: Rol;
  referido_id?: number | null;
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: Usuario | null;
  login: (usuario: Usuario, token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<Usuario | null>(() => {
    try {
      return JSON.parse(localStorage.getItem('multicatalogo_user') ?? 'null') as Usuario | null;
    } catch { return null; }
  });
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => Boolean(localStorage.getItem('multicatalogo_token') && user));

  const login = (usuario: Usuario, token: string) => {
    saveToken(token);
    localStorage.setItem('multicatalogo_user', JSON.stringify(usuario));
    setIsAuthenticated(true);
    setUser(usuario);
  };

  const logout = () => {
    clearToken();
    localStorage.removeItem('multicatalogo_user');
    setIsAuthenticated(false);
    setUser(null);
  };

  useEffect(() => {
    const expire = () => {
      clearToken();
      localStorage.removeItem('multicatalogo_user');
      setIsAuthenticated(false);
      setUser(null);
    };
    window.addEventListener('session-expired', expire);
    return () => window.removeEventListener('session-expired', expire);
  }, []);

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
