'use client';
import { Suspense, useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useQuestionStore } from '@/stores/questionStore';
import { questionApi, type QuestionInput } from '@/api/question';
import { DEFAULT_KNOWLEDGE_POINTS, QUESTION_TYPES, SUBJECTS } from '@/constants';
import { questionTypeText } from '@/utils/format';

function QuestionForm() {
  const router = useRouter();
  const params = useSearchParams();
  const editId = params.get('id');
  const { create, update } = useQuestionStore();

  const [form, setForm] = useState<QuestionInput>({
    type: QUESTION_TYPES.SINGLE,
    subject: SUBJECTS[0],
    knowledge_points: [DEFAULT_KNOWLEDGE_POINTS[0]],
    difficulty: 'easy',
    content: '',
    options: [
      { key: 'A', text: '' },
      { key: 'B', text: '' },
      { key: 'C', text: '' },
      { key: 'D', text: '' },
    ],
    answer: '',
    analysis: '',
    score: 5,
    status: 'published',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (editId) {
      questionApi.get(editId).then((q) => {
        setForm({
          type: q.type,
          subject: q.subject,
          knowledge_points: q.knowledge_points,
          difficulty: q.difficulty,
          content: q.content,
          options: (q.options ?? []).map((o) => ({ key: o.key, text: o.text })),
          answer: q.answer,
          analysis: q.analysis,
          score: q.score,
          status: q.status,
        });
      });
    }
  }, [editId]);

  const set = <K extends keyof QuestionInput>(k: K, v: QuestionInput[K]) => setForm((f) => ({ ...f, [k]: v }));

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      if (editId) {
        await update(editId, form);
      } else {
        await create(form);
      }
      router.push('/questions');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const needOptions = form.type === QUESTION_TYPES.SINGLE || form.type === QUESTION_TYPES.MULTIPLE;

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <h1 className="text-xl font-bold text-gray-800">{editId ? '编辑题目' : '新增题目'}</h1>
      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-gray-200 bg-white p-6">
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="text-sm text-gray-600">题型</label>
            <select value={form.type} onChange={(e) => set('type', e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
              {Object.values(QUESTION_TYPES).map((t) => <option key={t} value={t}>{questionTypeText(t)}</option>)}
            </select>
          </div>
          <div>
            <label className="text-sm text-gray-600">学科</label>
            <select value={form.subject} onChange={(e) => set('subject', e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
              {SUBJECTS.map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          <div>
            <label className="text-sm text-gray-600">难度</label>
            <select value={form.difficulty} onChange={(e) => set('difficulty', e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
              <option value="easy">容易</option>
              <option value="medium">中等</option>
              <option value="hard">困难</option>
            </select>
          </div>
          <div>
            <label className="text-sm text-gray-600">分值</label>
            <input type="number" step="0.5" min={0.5} value={form.score} onChange={(e) => set('score', Number(e.target.value))}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div className="sm:col-span-2">
            <label className="text-sm text-gray-600">知识点（逗号分隔）</label>
            <input value={form.knowledge_points.join(',')} onChange={(e) => set('knowledge_points', e.target.value.split(',').map((s) => s.trim()).filter(Boolean))}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
        </div>
        <div>
          <label className="text-sm text-gray-600">题干</label>
          <textarea value={form.content} onChange={(e) => set('content', e.target.value)} required rows={3}
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
        </div>
        {needOptions && (
          <div className="grid gap-2 sm:grid-cols-2">
            {form.options.map((o, i) => (
              <div key={o.key} className="flex items-center gap-2">
                <span className="w-5 text-sm font-medium text-gray-500">{o.key}</span>
                <input value={o.text} onChange={(e) => {
                  const opts = [...form.options];
                  opts[i] = { ...o, text: e.target.value };
                  set('options', opts);
                }} className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
              </div>
            ))}
          </div>
        )}
        <div>
          <label className="text-sm text-gray-600">答案</label>
          {form.type === QUESTION_TYPES.SINGLE || form.type === QUESTION_TYPES.JUDGE ? (
            <div className="mt-1 flex gap-4">
              {form.type === QUESTION_TYPES.JUDGE
                ? ['true', 'false'].map((v) => (
                    <label key={v} className="flex items-center gap-1 text-sm">
                      <input type="radio" checked={form.answer === v} onChange={() => set('answer', v)} /> {v === 'true' ? '正确' : '错误'}
                    </label>
                  ))
                : form.options.map((o) => (
                    <label key={o.key} className="flex items-center gap-1 text-sm">
                      <input type="radio" checked={form.answer === o.key} onChange={() => set('answer', o.key)} /> {o.key}
                    </label>
                  ))}
            </div>
          ) : (
            <input value={form.answer} onChange={(e) => set('answer', e.target.value)} required
              placeholder="多选题用逗号分隔，如 A,B；填空/简答填写参考答案"
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          )}
        </div>
        <div>
          <label className="text-sm text-gray-600">解析</label>
          <textarea value={form.analysis} onChange={(e) => set('analysis', e.target.value)} rows={2}
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex justify-end gap-3">
          <button type="button" onClick={() => router.push('/questions')}
            className="rounded-lg border border-gray-300 px-4 py-2 text-sm hover:bg-gray-50">取消</button>
          <button type="submit" disabled={loading}
            className="rounded-lg bg-brand-600 px-4 py-2 text-sm text-white hover:bg-brand-700 disabled:opacity-60">
            {loading ? '保存中…' : '保存'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default function NewQuestionPage() {
  return (
    <Suspense fallback={<div className="p-10 text-center text-gray-400">加载中…</div>}>
      <QuestionForm />
    </Suspense>
  );
}
