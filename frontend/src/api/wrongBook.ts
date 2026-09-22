import { request, buildQuery } from '@/utils/request';
import type { PageResult, WrongBook } from '@/types';

export const wrongBookApi = {
  list(query: { subject?: string; knowledge_point?: string; status?: string; page?: number; page_size?: number }) {
    return request<PageResult<WrongBook>>(`/wrong-books${buildQuery({ ...query })}`);
  },
  add(data: { question_id: string; exam_id?: string; exam_record_id?: string; note?: string }) {
    return request<WrongBook>('/wrong-books', { method: 'POST', body: JSON.stringify(data) });
  },
  update(id: string, data: { status?: string; note?: string }) {
    return request<WrongBook>(`/wrong-books/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  remove(id: string) {
    return request<null>(`/wrong-books/${id}`, { method: 'DELETE' });
  },
};
