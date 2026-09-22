import { request, buildQuery } from '@/utils/request';
import type { ExamRecord, ExamReport, PageResult, SaveDraftResult } from '@/types';

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
  // 自动保存草稿：version 为客户端所基于的草稿版本，服务端忽略乱序到达的旧请求
  saveDraft(id: string, answers: AnswerInput[], currentIndex: number, version: number, keepalive = false) {
    return request<SaveDraftResult>(`/exam-records/${id}/save`, {
      method: 'POST',
      body: JSON.stringify({ answers, current_index: currentIndex, version }),
      keepalive,
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
