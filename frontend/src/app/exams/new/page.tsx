'use client';
import { Suspense, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useExamStore } from '@/stores/examStore';
import { questionApi } from '@/api/question';
import { SUBJECTS } from '@/constants';
import type { Question } from '@/types';

function ExamForm() {
  const router = useRouter();
  const { create, autoGenerate } = useExamStore();
  const [mode, setMode] = useState<'manual' | 'auto'>('manual');
  const [questions, setQuestions] = useState<Question[]>([]);
  const [selected, setSelected] = useState<Record<string, number>>({});

  const [form, setForm] = useState({
    title: '',
    subject: SUBJECTS[0],
    description: '',
    pass_score: 60,
    duration_min: 60,
    start_at: '',
    end_at: '',
    shuffle_question: true,
    shuffle_option: true,
    knowledge_points: ['数据结构'],
    difficulty_dist: { easy: 3, medium: 3, hard: 2 },
    score_per_question: 5,
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    questionApi.list({ page: 1, page_size: 100 }).then((res) => setQuestions(res.list)).catch(() => undefined);
  }, []);

  const set = <K extends keyof typeof form>(k: K, v: (typeof form)[K]) => setForm((f) => ({ ...f, [k]: v }));

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    const startAt = new Date(form.start_at).toISOString();
    const endAt = new Date(form.end_at).toISOString();
    try {
      if (mode === 'manual') {
        const items = Object.entries(selected).map(([qid, score]) => ({ question_id: qid, score }));
        await create({
          title: form.title,
          subject: form.subject,
          description: form.description,
          pass_score: form.pass_score,
          duration_min: form.duration_min,
          start_at: startAt,
          end_at: endAt,
          shuffle_question: form.shuffle_question,
          shuffle_option: form.shuffle_option,
          questions: items,
        });
      } else {
        await autoGenerate({
          title: form.title,
          subject: form.subject,
          description: form.description,
          pass_score: form.pass_score,
          duration_min: form.duration_min,
          start_at: startAt,
          end_at: endAt,
          shuffle_question: form.shuffle_question,
          shuffle_option: form.shuffle_option,
          knowledge_points: form.knowledge_points,
          difficulty_dist: form.difficulty_dist,
          score_per_question: form.score_per_question,
        });
      }
      router.push('/exams');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const diffSum = Object.values(form.difficulty_dist).reduce((a, b) => a + b, 0);

  return (
    <div className="mx-auto max-w-4xl space-y-4">
      <h1 className="text-xl font-bold text-gray-800">创建考试</h1>
      <div className="flex gap-2">
        {(['manual', 'auto'] as const).map((m) => (
          <button key={m} onClick={() => setMode(m)}
            className={`rounded-lg px-4 py-2 text-sm ${mode === m ? 'bg-brand-600 text-white' : 'border border-gray-300 text-gray-600 hover:bg-gray-50'}`}>
            {m === 'manual' ? '手动选题' : '自动组卷'}
          </button>
        ))}
      </div>
      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-gray-200 bg-white p-6">
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="text-sm text-gray-600">考试名称</label>
            <input value={form.title} onChange={(e) => set('title', e.target.value)} required minLength={2}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label className="text-sm text-gray-600">学科</label>
            <select value={form.subject} onChange={(e) => set('subject', e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
              {SUBJECTS.map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          <div>
            <label className="text-sm text-gray-600">考试时长（分钟）</label>
            <input type="number" min={1} value={form.duration_min} onChange={(e) => set('duration_min', Number(e.target.value))}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label className="text-sm text-gray-600">及格分</label>
            <input type="number" min={0} value={form.pass_score} onChange={(e) => set('pass_score', Number(e.target.value))}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label className="text-sm text-gray-600">开始时间</label>
            <input type="datetime-local" value={form.start_at} onChange={(e) => set('start_at', e.target.value)} required
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label className="text-sm text-gray-600">结束时间</label>
            <input type="datetime-local" value={form.end_at} onChange={(e) => set('end_at', e.target.value)} required
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label className="text-sm text-gray-600">描述</label>
            <input value={form.description} onChange={(e) => set('description', e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          </div>
        </div>

        <div className="flex flex-wrap gap-6">
          <label className="flex items-center gap-2 text-sm text-gray-600">
            <input type="checkbox" checked={form.shuffle_question} onChange={(e) => set('shuffle_question', e.target.checked)} />
            随机打乱题目顺序（防作弊）
          </label>
          <label className="flex items-center gap-2 text-sm text-gray-600">
            <input type="checkbox" checked={form.shuffle_option} onChange={(e) => set('shuffle_option', e.target.checked)} />
            随机打乱选项顺序（防作弊）
          </label>
        </div>

        {mode === 'manual' ? (
          <div>
            <label className="text-sm text-gray-600">选择题库题目（点击勾选，可设置每题分值）</label>
            <div className="mt-2 max-h-80 space-y-1 overflow-y-auto rounded-lg border border-gray-200 p-2">
              {questions.length === 0 && <p className="p-3 text-sm text-gray-400">题库暂无题目，请先到题库页录入或导入</p>}
              {questions.map((q) => (
                <label key={q.id} className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm hover:bg-gray-50">
                  <input type="checkbox" checked={q.id in selected} onChange={(e) => {
                    const s = { ...selected };
                    if (e.target.checked) s[q.id] = q.score;
                    else delete s[q.id];
                    setSelected(s);
                  }} />
                  <span className="flex-1 truncate">{q.content}</span>
                  <span className="text-xs text-gray-400">{q.type} · {q.score}分</span>
                  {q.id in selected && (
                    <input type="number" min={0.5} step={0.5} value={selected[q.id]} onChange={(ev) => setSelected({ ...selected, [q.id]: Number(ev.target.value) })}
                      className="w-20 rounded border border-gray-300 px-2 py-1 text-xs" />
                  )}
                </label>
              ))}
            </div>
            <p className="mt-1 text-xs text-gray-400">已选 {Object.keys(selected).length} 题，总分 {Object.values(selected).reduce((a, b) => a + b, 0)}</p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <label className="text-sm text-gray-600">知识点覆盖（逗号分隔）</label>
              <input value={form.knowledge_points.join(',')} onChange={(e) => set('knowledge_points', e.target.value.split(',').map((s) => s.trim()).filter(Boolean))}
                className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label className="text-sm text-gray-600">每题分值</label>
              <input type="number" min={0.5} step={0.5} value={form.score_per_question} onChange={(e) => set('score_per_question', Number(e.target.value))}
                className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
            </div>
            {(['easy', 'medium', 'hard'] as const).map((d) => (
              <div key={d}>
                <label className="text-sm text-gray-600">{d === 'easy' ? '容易' : d === 'medium' ? '中等' : '困难'} 题量</label>
                <input type="number" min={0} value={form.difficulty_dist[d]} onChange={(e) => set('difficulty_dist', { ...form.difficulty_dist, [d]: Number(e.target.value) })}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
              </div>
            ))}
            <p className="text-xs text-gray-400 sm:col-span-2">预计抽取 {diffSum} 题，总分约 {diffSum * form.score_per_question}</p>
          </div>
        )}

        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex justify-end gap-3">
          <button type="button" onClick={() => router.push('/exams')}
            className="rounded-lg border border-gray-300 px-4 py-2 text-sm hover:bg-gray-50">取消</button>
          <button type="submit" disabled={loading}
            className="rounded-lg bg-brand-600 px-4 py-2 text-sm text-white hover:bg-brand-700 disabled:opacity-60">
            {loading ? '创建中…' : '创建考试'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default function NewExamPage() {
  return (
    <Suspense fallback={<div className="p-10 text-center text-gray-400">加载中…</div>}>
      <ExamForm />
    </Suspense>
  );
}
