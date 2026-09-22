'use client';
import { Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useCountdown } from '@/hooks/useCountdown';
import { useAutoSave } from '@/hooks/useAutoSave';
import { recordApi, type AnswerInput } from '@/api/record';
import { examApi } from '@/api/exam';
import { QuestionTypeBadge } from '@/components/StatusBadge';
import { questionTypeText } from '@/utils/format';
import type { Exam, ExamRecord } from '@/types';

function ExamTake() {
  const router = useRouter();
  const params = useSearchParams();
  const recordIdParam = params.get('recordId') ?? '';
  const examId = params.get('examId') ?? '';

  const [exam, setExam] = useState<Exam | null>(null);
  const [record, setRecord] = useState<ExamRecord | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [marked, setMarked] = useState<Set<string>>(new Set());
  const [current, setCurrent] = useState(0);
  const [cheatCount, setCheatCount] = useState(0);
  const [cheatEvents, setCheatEvents] = useState<{ type: string; detail: string }[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [notified5, setNotified5] = useState(false);
  const cheatRef = useRef(0);
  const eventsRef = useRef<{ type: string; detail: string }[]>([]);
  const submittingRef = useRef(false);

  const load = useCallback(async () => {
    if (!recordIdParam && !examId) return;
    let rec: ExamRecord;
    if (recordIdParam) {
      rec = await recordApi.get(recordIdParam);
    } else if (examId) {
      rec = await recordApi.start(examId);
    } else {
      return;
    }
    setRecord(rec);
    try {
      const e = await examApi.get(rec.exam_id);
      setExam(e);
    } catch {
      setExam(null);
    }
    // 断点续答：恢复上次已保存的答案与当前题号（未交卷状态不变，答卷仍为 in_progress）
    const ans: Record<string, string> = {};
    rec.questions.forEach((q) => {
      if (q.user_answer) ans[q.question_id] = q.user_answer;
    });
    setAnswers(ans);
    const restoredIndex = Math.min(Math.max(rec.current_index ?? 0, 0), Math.max(rec.questions.length - 1, 0));
    setCurrent(restoredIndex);
  }, [recordIdParam, examId]);

  useEffect(() => {
    load();
  }, [load]);

  // 防作弊：切屏检测（visibilitychange）+ 禁止复制粘贴
  useEffect(() => {
    const onVisibility = () => {
      if (document.hidden) {
        cheatRef.current += 1;
        setCheatCount(cheatRef.current);
        eventsRef.current.push({ type: 'switch_tab', detail: `第${cheatRef.current}次切屏` });
        setCheatEvents([...eventsRef.current]);
      }
    };
    const onCopy = (e: ClipboardEvent) => {
      e.preventDefault();
      cheatRef.current += 1;
      setCheatCount(cheatRef.current);
      eventsRef.current.push({ type: 'copy_paste', detail: '尝试复制' });
      setCheatEvents([...eventsRef.current]);
    };
    const onPaste = (e: ClipboardEvent) => {
      e.preventDefault();
      cheatRef.current += 1;
      setCheatCount(cheatRef.current);
      eventsRef.current.push({ type: 'copy_paste', detail: '尝试粘贴' });
      setCheatEvents([...eventsRef.current]);
    };
    const onBlur = () => {
      cheatRef.current += 1;
      setCheatCount(cheatRef.current);
      eventsRef.current.push({ type: 'blur', detail: '窗口失焦' });
      setCheatEvents([...eventsRef.current]);
    };
    document.addEventListener('visibilitychange', onVisibility);
    document.addEventListener('copy', onCopy);
    document.addEventListener('paste', onPaste);
    window.addEventListener('blur', onBlur);
    return () => {
      document.removeEventListener('visibilitychange', onVisibility);
      document.removeEventListener('copy', onCopy);
      document.removeEventListener('paste', onPaste);
      window.removeEventListener('blur', onBlur);
    };
  }, []);

  const endAt = useMemo(() => {
    if (!record) return 0;
    const duration = exam?.duration_min ?? 60;
    return new Date(record.started_at).getTime() + duration * 60 * 1000;
  }, [record, exam]);

  // 自动保存草稿（作答停顿后写入，刷新/换设备重进时恢复答案、剩余时间、当前题号）
  const draft = useAutoSave({
    recordId: record && record.status === 'in_progress' ? record.id : null,
    answers,
    current,
    version: record?.answer_version ?? 0,
  });

  const doSubmit = useCallback(
    async (auto = false) => {
      if (!record || submittingRef.current) return;
      submittingRef.current = true;
      setSubmitting(true);
      // 交卷前先尽量把未落盘的答案写入草稿；失败（网络/已截止/版本冲突）也不阻断交卷，
      // 截止时以最后已保存的答案为准（与定时自动提交口径一致）
      await draft.flush().catch(() => false);
      try {
        const ansList: AnswerInput[] = Object.entries(answers).map(([question_id, answer]) => ({ question_id, answer }));
        await recordApi.submit(record.id, ansList, cheatRef.current, eventsRef.current);
        alert(auto ? '考试时间到，答卷已自动提交' : '答卷提交成功，客观题已自动评分');
        router.push('/records');
      } catch (err) {
        submittingRef.current = false;
        setSubmitting(false);
        alert((err as Error).message);
      }
    },
    [record, answers, router, draft],
  );

  const { text, warn5 } = useCountdown(endAt, () => {
    doSubmit(true);
  });

  useEffect(() => {
    if (warn5 && !notified5 && record && record.status === 'in_progress') {
      setNotified5(true);
      alert('距离考试结束还有 5 分钟，请抓紧作答！');
    }
  }, [warn5, notified5, record]);

  if (!record) return <div className="p-10 text-center text-gray-400">加载中…</div>;

  if (record.status !== 'in_progress') {
    return (
      <div className="mx-auto max-w-md rounded-xl border border-gray-200 bg-white p-8 text-center">
        <h1 className="text-lg font-bold text-gray-800">该答卷已提交</h1>
        <p className="mt-2 text-sm text-gray-500">最终得分：{record.final_score || record.objective_score}</p>
        <button onClick={() => router.push(`/records/review?recordId=${record.id}`)}
          className="mt-4 rounded-lg bg-brand-600 px-4 py-2 text-sm text-white">查看答卷</button>
      </div>
    );
  }

  const q = record.questions[current];

  const setAnswer = (val: string) => {
    setAnswers((a) => ({ ...a, [q.question_id]: val }));
  };

  const toggleMark = () => {
    setMarked((m) => {
      const next = new Set(m);
      if (next.has(q.question_id)) next.delete(q.question_id);
      else next.add(q.question_id);
      return next;
    });
  };

  const answeredCount = Object.keys(answers).filter((k) => answers[k] && answers[k].trim() !== '').length;

  const saveHint = (() => {
    switch (draft.status) {
      case 'saving':
        return <span className="text-xs text-gray-400">正在保存…</span>;
      case 'pending':
        return <span className="text-xs text-gray-400">待保存…</span>;
      case 'saved':
        return <span className="text-xs text-green-600">✓ 答案已自动保存</span>;
      case 'error':
        return (
          <span className="flex items-center gap-2 text-xs text-red-600">
            ⚠️ 保存失败，作答仍保留在本页
            <button onClick={draft.retry} className="rounded border border-red-300 px-1.5 py-0.5 hover:bg-red-50">重试</button>
          </span>
        );
      default:
        return null;
    }
  })();

  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_260px]">
      <div className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-gray-200 bg-white px-5 py-3">
          <div>
            <h1 className="font-bold text-gray-800">{record.exam_title}</h1>
            <p className="text-xs text-gray-400">共 {record.questions.length} 题 · 已答 {answeredCount} 题</p>
          </div>
          <div className="flex items-center gap-3">
            {saveHint}
            {warn5 && <span className="text-sm font-medium text-red-600">⚠️ 剩余不足 5 分钟</span>}
            <span className={`rounded-lg px-3 py-1 font-mono text-lg font-bold ${warn5 ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-700'}`}>
              {text}
            </span>
          </div>
        </div>

        {draft.status === 'conflict' && (
          <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-800">
            <span>⚠️ {draft.errorMsg}</span>
            <button onClick={() => window.location.reload()}
              className="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-700">刷新加载最新答案</button>
          </div>
        )}
        {draft.status === 'error' && (
          <div className="rounded-xl border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-700">
            ⚠️ {draft.errorMsg}。页面上的作答不会被改动，恢复网络后可自动重试，或点击
            <button onClick={draft.retry} className="mx-1 font-medium underline">立即重试</button>。
          </div>
        )}

        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-400">第 {current + 1} 题 / 共 {record.questions.length} 题</span>
            <QuestionTypeBadge type={q.type} />
          </div>
          <p className="mt-3 text-base font-medium text-gray-800">{q.content}</p>
          <p className="mt-1 text-xs text-gray-400">{q.score} 分 · {questionTypeText(q.type)}</p>

          <div className="mt-5 space-y-2">
            {q.type === 'judge' ? (
              <div className="flex gap-6">
                {['true', 'false'].map((v) => (
                  <label key={v} className="flex items-center gap-2 text-sm">
                    <input type="radio" name={q.question_id} checked={answers[q.question_id] === v} onChange={() => setAnswer(v)} />
                    {v === 'true' ? '正确' : '错误'}
                  </label>
                ))}
              </div>
            ) : (q.options?.length ?? 0) > 0 && (q.type === 'single' || q.type === 'multiple') ? (
              q.options.map((o) => (
                <label key={o.key} className="flex items-start gap-3 rounded-lg border border-gray-200 px-3 py-2 hover:bg-gray-50">
                  <input
                    type={q.type === 'single' ? 'radio' : 'checkbox'}
                    name={q.question_id}
                    checked={q.type === 'single' ? answers[q.question_id] === o.key : (answers[q.question_id] ?? '').split(',').includes(o.key)}
                    onChange={() => {
                      if (q.type === 'single') {
                        setAnswer(o.key);
                      } else {
                        const cur = (answers[q.question_id] ?? '').split(',').filter(Boolean);
                        const next = cur.includes(o.key) ? cur.filter((k) => k !== o.key) : [...cur, o.key];
                        setAnswer(next.sort().join(','));
                      }
                    }}
                  />
                  <span className="text-sm text-gray-700">{o.key}. {o.text}</span>
                </label>
              ))
            ) : (
              <textarea
                value={answers[q.question_id] ?? ''}
                onChange={(e) => setAnswer(e.target.value)}
                onCopy={(e) => e.preventDefault()}
                onPaste={(e) => e.preventDefault()}
                rows={4}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
                placeholder="请输入答案…"
              />
            )}
          </div>

          <div className="mt-6 flex items-center justify-between">
            <button onClick={toggleMark}
              className={`rounded-lg px-3 py-1.5 text-sm ${marked.has(q.question_id) ? 'bg-amber-100 text-amber-700' : 'border border-gray-300 text-gray-600 hover:bg-gray-50'}`}>
              {marked.has(q.question_id) ? '★ 已标记稍后作答' : '标记稍后作答'}
            </button>
            <div className="flex gap-2">
              <button onClick={() => setCurrent((c) => Math.max(0, c - 1))} disabled={current === 0}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm hover:bg-gray-50 disabled:opacity-40">上一题</button>
              {current < record.questions.length - 1 ? (
                <button onClick={() => setCurrent((c) => c + 1)}
                  className="rounded-lg bg-brand-600 px-4 py-2 text-sm text-white hover:bg-brand-700">下一题</button>
              ) : (
                <button onClick={() => doSubmit(false)} disabled={submitting}
                  className="rounded-lg bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700 disabled:opacity-60">
                  {submitting ? '提交中…' : '交卷'}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      <aside className="rounded-xl border border-gray-200 bg-white p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-700">题目导航</h3>
          <span className="text-xs text-red-500">切屏 {cheatCount} 次</span>
        </div>
        <div className="mt-3 grid grid-cols-6 gap-1.5">
          {record.questions.map((item, i) => {
            const answered = answers[item.question_id] && answers[item.question_id].trim() !== '';
            const isMarked = marked.has(item.question_id);
            return (
              <button
                key={item.question_id}
                onClick={() => setCurrent(i)}
                className={`h-8 rounded text-xs font-medium ${
                  i === current
                    ? 'bg-brand-600 text-white'
                    : isMarked
                    ? 'bg-amber-100 text-amber-700'
                    : answered
                    ? 'bg-green-100 text-green-700'
                    : 'bg-gray-100 text-gray-500'
                }`}
              >
                {i + 1}
              </button>
            );
          })}
        </div>
        <div className="mt-3 space-y-1 text-xs text-gray-400">
          <p>■ 当前题 <span className="ml-1 text-gray-300">■</span> 已作答 <span className="ml-1 text-amber-500">★</span> 标记</p>
        </div>
        <button
          onClick={() => {
            if (confirm(`已作答 ${answeredCount}/${record.questions.length} 题，确定交卷？`)) doSubmit(false);
          }}
          disabled={submitting}
          className="mt-4 w-full rounded-lg bg-red-600 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-60"
        >
          {submitting ? '提交中…' : '交卷'}
        </button>
      </aside>
    </div>
  );
}

export default function ExamTakePage() {
  return (
    <Suspense fallback={<div className="p-10 text-center text-gray-400">加载中…</div>}>
      <ExamTake />
    </Suspense>
  );
}
