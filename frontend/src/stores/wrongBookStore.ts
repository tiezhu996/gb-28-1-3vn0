// 错题本状态（学生复习）。
import { create } from 'zustand';
import { wrongBookApi } from '@/api/wrongBook';
import type { WrongBook } from '@/types';

interface WrongBookState {
  list: WrongBook[];
  total: number;
  loading: boolean;
  fetch: (query: { subject?: string; knowledge_point?: string; status?: string; page?: number; page_size?: number }) => Promise<void>;
  add: (data: { question_id: string; exam_id?: string; exam_record_id?: string; note?: string }) => Promise<WrongBook>;
  update: (id: string, data: { status?: string; note?: string }) => Promise<WrongBook>;
  remove: (id: string) => Promise<void>;
}

export const useWrongBookStore = create<WrongBookState>((set) => ({
  list: [],
  total: 0,
  loading: false,
  fetch: async (query) => {
    set({ loading: true });
    try {
      const res = await wrongBookApi.list(query);
      set({ list: res.list, total: res.total });
    } finally {
      set({ loading: false });
    }
  },
  add: (data) => wrongBookApi.add(data),
  update: (id, data) => wrongBookApi.update(id, data),
  remove: async (id) => {
    await wrongBookApi.remove(id);
  },
}));
