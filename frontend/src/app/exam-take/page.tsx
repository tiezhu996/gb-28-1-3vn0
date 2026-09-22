'use client';
import { Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useCountdown } from '@/hooks/useCountdown';
import { useAutoSave, type SaveStatus } from '@/hooks/useAutoSave';
import { recordApi, type AnswerInput } from '@/api/record';
import { examApi } from '@/api/exam';
import { QuestionTypeBadge } from '@/components/StatusBadge';
import { questionTypeText } from '@/utils/format';
import type { Exam, ExamRecord } from '@/types';

function SaveIndicator({ status, onRetry }: { status: SaveStatus; onRetry: () => void }) {
  if (status === 'saving') {
    return <span className="text-xs text-gray-400">⏳ 自动保存中…</span>;
  }
  if (status === 'saved') {
    return <span className="text-xs text-green-600">✓ 已自动保存</span>;
  }
  if (status === 'error') {
    return (
      <button onClick={onRetry} className="text-xs text-red-600 hover:underline">
        ⚠ 自动保存失败，点击重试（页面答案未改动）
      </button>
    );
  }
  if (status === 'conflict') {
    return <span className="text-xs text-red-600">⚠ 检测到其他设备有更新的作答，请刷新页面后继续</span>;
  }
  return null;
}

function ExamTake() {
  const router = useRouter();
  const params = useSearchParams();
  const recordId = params.get('recordId') ?? '';
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

  const { status: saveStatus, scheduleSave, flush, retry } = useAutoSave(record);

  // 始终保存最新作答/题号的引用，供自动保存、交卷、倒计时回调读取
  const answersRef = useRef<Record<string, string>>({});
  const currentRef = useRef(0);
  const recordRef = useRef<ExamRecord | null>(null);
  const readyRef = useRef(false);
  // 恢复后的初始签名：与基线一致时不触发自动保存（只有真实作答/切题变化才保存）
  const baselineSigRef = useRef('');
  answersRef.current = answers;
  currentRef.current = current;
  recordRef.current = record;

  const load = useCallback(async () => {
    if (!recordId && !examId) return;
    let rec: ExamRecord;
    if (recordId) {
      rec = await recordApi.get(recordId);
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
    // 断点续答：恢复最后已保存的答案与当前题号（剩余时间由 started_at + 时长决定，倒计时自然恢复）
    const ans: Record<string, string> = {};
    rec.questions.forEach((q) => {
      if (q.user_answer) ans[q.question_id] = q.user_answer;
    });
    setAnswers(ans);
    answersRef.current = ans;
    const restoredIndex = Math.min(Math.max(rec.current_index ?? 0, 0), Math.max(rec.questions.length - 1, 0));
    setCurrent(restoredIndex);
    currentRef.current = restoredIndex;
    baselineSigRef.current = `${JSON.stringify(Object.entries(ans).sort())}:${restoredIndex}`;
    readyRef.current = true;
  }, [recordId, examId]);

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

  // 作答或题号变化后短暂停顿自动保存；与恢复时的基线签名相同则跳过（不产生无意义保存）
  useEffect(() => {
    if (!readyRef.current) return;
    const sig = `${JSON.stringify(Object.entries(answers).sort())}:${current}`;
    if (sig === baselineSigRef.current) return;
    const rec = recordRef.current;
    if (!rec || rec.status !== 'in_progress') return;
    const ansList: AnswerInput[] = Object.entries(answers).map(([question_id, answer]) => ({ question_id, answer }));
    scheduleSave(ansList, current);
  }, [answers, current, scheduleSave]);

  const doSubmit = useCallback(
    async (auto = false) => {
      const rec = recordRef.current;
      if (!rec) return;
      setSubmitting(true);
      try {
        // 交卷前先把最后一次未落盘的草稿写上去（截止时用最后已保存答案交卷）；
        // 即使保存失败（如刚好超时），也继续走原交卷/自动交卷流程，由服务端拒绝或定时任务兜底
        await flush();
        const ansList: AnswerInput[] = Object.entries(answersRef.current).map(([question_id, answer]) => ({ question_id, answer }));
        await recordApi.submit(rec.id, ansList, cheatRef.current, eventsRef.current);
        alert(auto ? '考试时间到，答卷已自动提交' : '答卷提交成功，客观题已自动评分');
        router.push('/records');
      } catch (err) {
        alert((err as Error).message);
      } finally {
        setSubmitting(false);
      }
    },
    [router, flush],
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

  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_260px]">
      <div className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-gray-200 bg-white px-5 py-3">
          <div>
            <h1 className="font-bold text-gray-800">{record.exam_title}</h1>
            <p className="text-xs text-gray-400">共 {record.questions.length} 题 · 已答 {answeredCount} 题</p>
          </div>
          <div className="flex items-center gap-3">
            <SaveIndicator status={saveStatus} onRetry={retry} />
            {warn5 && <span className="text-sm font-medium text-red-600">⚠️ 剩余不足 5 分钟</span>}
            <span className={`rounded-lg px-3 py-1 font-mono text-lg font-bold ${warn5 ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-700'}`}>
              {text}
            </span>
          </div>
        </div>

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
