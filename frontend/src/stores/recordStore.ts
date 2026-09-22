// 考试记录状态（学生答题、教师阅卷/报告复用）。
import { create } from 'zustand';
import { recordApi, type AnswerInput } from '@/api/record';
import type { ExamRecord } from '@/types';

interface RecordState {
  mine: ExamRecord[];
  total: number;
  loading: boolean;
  fetchMine: (query: { status?: string; page?: number; page_size?: number }) => Promise<void>;
  start: (examId: string) => Promise<ExamRecord>;
  submit: (id: string, answers: AnswerInput[], cheatCount: number, cheatEvents: { type: string; detail: string }[]) => Promise<ExamRecord>;
  get: (id: string) => Promise<ExamRecord>;
}

export const useRecordStore = create<RecordState>((set) => ({
  mine: [],
  total: 0,
  loading: false,
  fetchMine: async (query) => {
    set({ loading: true });
    try {
      const res = await recordApi.mine(query);
      set({ mine: res.list, total: res.total });
    } finally {
      set({ loading: false });
    }
  },
  start: (examId) => recordApi.start(examId),
  submit: (id, answers, cheatCount, cheatEvents) => recordApi.submit(id, answers, cheatCount, cheatEvents),
  get: (id) => recordApi.get(id),
}));
