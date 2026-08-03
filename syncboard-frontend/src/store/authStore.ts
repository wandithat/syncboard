import { create } from 'zustand';
import type { User } from '../types';
import { authApi } from '../api/auth';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
  init: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: localStorage.getItem('token'),
  isAuthenticated: !!localStorage.getItem('token'),

  login: async (email, password) => {
    const token = await authApi.login(email, password);
    localStorage.setItem('token', token);
    set({ token, isAuthenticated: true });
  },

  register: async (email, password) => {
    await authApi.register(email, password);
    // После регистрации сразу логиним или просим войти (здесь просто выходим)
  },

  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null, isAuthenticated: false });
  },

  init: () => {
    const token = localStorage.getItem('token');
    set({ token, isAuthenticated: !!token });
  },
}));
