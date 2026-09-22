import { request, upload, buildQuery } from '@/utils/request';
import type { PageResult, Question } from '@/types';

export interface QuestionQuery {
  subject?: string;
  type?: string;
  difficulty?: string;
  knowledge_point?: string;
  keyword?: string;
  status?: string;
  page?: number;
  page_size?: number;
}

export interface QuestionInput {
  type: string;
  subject: string;
  knowledge_points: string[];
  difficulty: string;
  content: string;
  options: { key: string; text: string }[];
  answer: string;
  analysis?: string;
  score: number;
  status?: string;
}

export const questionApi = {
  list(query: QuestionQuery) {
    return request<PageResult<Question>>(`/questions${buildQuery({ ...query })}`);
  },
  get(id: string) {
    return request<Question>(`/questions/${id}`);
  },
  create(data: QuestionInput) {
    return request<Question>('/questions', { method: 'POST', body: JSON.stringify(data) });
  },
  update(id: string, data: Partial<QuestionInput>) {
    return request<Question>(`/questions/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  remove(id: string) {
    return request<null>(`/questions/${id}`, { method: 'DELETE' });
  },
  importExcel(file: File) {
    return upload<{ count: number }>('/questions/import', file);
  },
  templateUrl() {
    return '/api/v1/questions/template';
  },
};
