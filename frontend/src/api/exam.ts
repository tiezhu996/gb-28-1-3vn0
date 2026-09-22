import { request, buildQuery } from '@/utils/request';
import type { Exam, PageResult } from '@/types';

export interface ExamQuestionInput {
  question_id: string;
  score?: number;
}

export interface ExamInput {
  title: string;
  subject: string;
  description?: string;
  total_score?: number;
  pass_score?: number;
  duration_min: number;
  start_at: string;
  end_at: string;
  shuffle_question?: boolean;
  shuffle_option?: boolean;
  questions?: ExamQuestionInput[];
}

export interface AutoGenerateInput {
  title: string;
  subject: string;
  description?: string;
  pass_score?: number;
  duration_min: number;
  start_at: string;
  end_at: string;
  shuffle_question?: boolean;
  shuffle_option?: boolean;
  knowledge_points: string[];
  difficulty_dist: Record<string, number>;
  score_per_question: number;
}

export const examApi = {
  list(query: { title?: string; subject?: string; status?: string; page?: number; page_size?: number }) {
    return request<PageResult<Exam>>(`/exams${buildQuery({ ...query })}`);
  },
  get(id: string) {
    return request<Exam>(`/exams/${id}`);
  },
  create(data: ExamInput) {
    return request<Exam>('/exams', { method: 'POST', body: JSON.stringify(data) });
  },
  autoGenerate(data: AutoGenerateInput) {
    return request<Exam>('/exams/auto-generate', { method: 'POST', body: JSON.stringify(data) });
  },
  update(id: string, data: Partial<ExamInput>) {
    return request<Exam>(`/exams/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  publish(id: string) {
    return request<Exam>(`/exams/${id}/publish`, { method: 'POST' });
  },
  close(id: string) {
    return request<Exam>(`/exams/${id}/close`, { method: 'POST' });
  },
  remove(id: string) {
    return request<null>(`/exams/${id}`, { method: 'DELETE' });
  },
};
