import { request, buildQuery } from '@/utils/request';
import type { PageResult, User } from '@/types';

export const userApi = {
  list(query: { role?: string; status?: string; keyword?: string; page?: number; page_size?: number }) {
    return request<PageResult<User>>(`/users${buildQuery({ ...query })}`);
  },
  create(data: { name: string; email: string; password: string; role: string }) {
    return request<User>('/users', { method: 'POST', body: JSON.stringify(data) });
  },
  update(id: string, data: { name?: string; role?: string; status?: string }) {
    return request<User>(`/users/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  remove(id: string) {
    return request<null>(`/users/${id}`, { method: 'DELETE' });
  },
};
