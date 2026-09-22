import { request } from '@/utils/request';
import type { LoginResponse, User } from '@/types';

export const authApi = {
  register(data: { name: string; email: string; password: string; role?: string }) {
    return request<User>('/auth/register', { method: 'POST', body: JSON.stringify(data) });
  },
  login(data: { email: string; password: string }) {
    return request<LoginResponse>('/auth/login', { method: 'POST', body: JSON.stringify(data) });
  },
  me() {
    return request<User>('/auth/me');
  },
  changePassword(data: { old_password: string; new_password: string }) {
    return request<null>('/auth/password', { method: 'PUT', body: JSON.stringify(data) });
  },
};
