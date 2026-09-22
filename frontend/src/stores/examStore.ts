// 试卷状态（教师组卷、学生查询复用）。
import { create } from 'zustand';
import { examApi, type AutoGenerateInput, type ExamInput } from '@/api/exam';
import type { Exam } from '@/types';

interface ExamState {
  list: Exam[];
  total: number;
  loading: boolean;
  fetch: (query: { title?: string; subject?: string; status?: string; page?: number; page_size?: number }) => Promise<void>;
  create: (data: ExamInput) => Promise<Exam>;
  autoGenerate: (data: AutoGenerateInput) => Promise<Exam>;
  update: (id: string, data: Partial<ExamInput>) => Promise<Exam>;
  publish: (id: string) => Promise<Exam>;
  close: (id: string) => Promise<Exam>;
  remove: (id: string) => Promise<void>;
}

export const useExamStore = create<ExamState>((set) => ({
  list: [],
  total: 0,
  loading: false,
  fetch: async (query) => {
    set({ loading: true });
    try {
      const res = await examApi.list(query);
      set({ list: res.list, total: res.total });
    } finally {
      set({ loading: false });
    }
  },
  create: (data) => examApi.create(data),
  autoGenerate: (data) => examApi.autoGenerate(data),
  update: (id, data) => examApi.update(id, data),
  publish: (id) => examApi.publish(id),
  close: (id) => examApi.close(id),
  remove: async (id) => {
    await examApi.remove(id);
  },
}));
