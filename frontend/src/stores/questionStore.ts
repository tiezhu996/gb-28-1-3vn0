// 题库状态（教师维护、学生查询复用）。
import { create } from 'zustand';
import { questionApi, type QuestionInput, type QuestionQuery } from '@/api/question';
import type { Question } from '@/types';

interface QuestionState {
  list: Question[];
  total: number;
  loading: boolean;
  fetch: (query: QuestionQuery) => Promise<void>;
  create: (data: QuestionInput) => Promise<Question>;
  update: (id: string, data: Partial<QuestionInput>) => Promise<Question>;
  remove: (id: string) => Promise<void>;
  importExcel: (file: File) => Promise<number>;
}

export const useQuestionStore = create<QuestionState>((set) => ({
  list: [],
  total: 0,
  loading: false,
  fetch: async (query) => {
    set({ loading: true });
    try {
      const res = await questionApi.list(query);
      set({ list: res.list, total: res.total });
    } finally {
      set({ loading: false });
    }
  },
  create: async (data) => {
    const q = await questionApi.create(data);
    return q;
  },
  update: async (id, data) => {
    const q = await questionApi.update(id, data);
    return q;
  },
  remove: async (id) => {
    await questionApi.remove(id);
  },
  importExcel: async (file) => {
    const res = await questionApi.importExcel(file);
    return res.count;
  },
}));
