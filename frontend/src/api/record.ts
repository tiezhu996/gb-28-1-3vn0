import { request, buildQuery } from '@/utils/request';
import type { ExamRecord, ExamReport, PageResult } from '@/types';

export interface AnswerInput {
  question_id: string;
  answer: string;
}

export const recordApi = {
  start(examId: string) {
    return request<ExamRecord>(`/exam-records/${examId}/start`, { method: 'POST' });
  },
  mine(query: { status?: string; page?: number; page_size?: number }) {
    return request<PageResult<ExamRecord>>(`/exam-records/mine${buildQuery({ ...query })}`);
  },
  submit(id: string, answers: AnswerInput[], cheatCount: number, cheatEvents: { type: string; detail: string }[]) {
    return request<ExamRecord>(`/exam-records/${id}/submit`, {
      method: 'POST',
      body: JSON.stringify({ answers, cheat_count: cheatCount, cheat_events: cheatEvents }),
    });
  },
  get(id: string) {
    return request<ExamRecord>(`/exam-records/${id}`);
  },
  listByExam(examId: string, query: { status?: string; page?: number; page_size?: number }) {
    return request<PageResult<ExamRecord>>(`/exams/${examId}/records${buildQuery({ ...query })}`);
  },
  report(examId: string) {
    return request<ExamReport>(`/exams/${examId}/report`);
  },
  grade(id: string, grades: { question_id: string; score: number; comment?: string }[]) {
    return request<ExamRecord>(`/exam-records/${id}/grade`, {
      method: 'POST',
      body: JSON.stringify({ grades }),
    });
  },
  autoSubmit(id: string) {
    return request<ExamRecord>(`/exam-records/${id}/auto-submit`, { method: 'POST' });
  },
};
