// 认证状态：JWT 持久化 + 登录态管理（与后端 JWT 认证、前端路由守卫对应）。
import { create } from 'zustand';
import { authApi } from '@/api/auth';
import { setAuthToken } from '@/utils/request';
import type { User } from '@/types';

interface AuthState {
  token: string | null;
  user: User | null;
  init: () => void;
  login: (email: string, password: string) => Promise<User>;
  register: (data: { name: string; email: string; password: string; role?: string }) => Promise<User>;
  logout: () => void;
}

function readStorage<T>(key: string): T | null {
  if (typeof window === 'undefined') return null;
  const raw = window.localStorage.getItem(key);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

// 模块加载时即恢复请求令牌，避免首屏请求早于 React effect 导致 401。
if (typeof window !== 'undefined') {
  const rawToken = window.localStorage.getItem('onlineexam_token');
  if (rawToken) {
    try {
      setAuthToken(JSON.parse(rawToken) as string);
    } catch {
      // 忽略损坏的令牌
    }
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  init: () => {
    const token = readStorage<string>('onlineexam_token');
    const user = readStorage<User>('onlineexam_user');
    if (token) setAuthToken(token);
    set({ token, user });
  },
  login: async (email, password) => {
    const res = await authApi.login({ email, password });
    setAuthToken(res.token);
    window.localStorage.setItem('onlineexam_token', JSON.stringify(res.token));
    window.localStorage.setItem('onlineexam_user', JSON.stringify(res.user));
    set({ token: res.token, user: res.user });
    return res.user;
  },
  register: async (data) => {
    const user = await authApi.register(data);
    return user;
  },
  logout: () => {
    setAuthToken(null);
    window.localStorage.removeItem('onlineexam_token');
    window.localStorage.removeItem('onlineexam_user');
    set({ token: null, user: null });
  },
}));
