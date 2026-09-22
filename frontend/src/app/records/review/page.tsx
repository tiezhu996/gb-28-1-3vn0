'use client';
import { Suspense, useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { recordApi } from '@/api/record';
import { wrongBookApi } from '@/api/wrongBook';
import { QuestionTypeBadge, StatusBadge } from '@/components/StatusBadge';
import { answerResultText, formatDateTime, recordStatusColor, recordStatusText } from '@/utils/format';
import { ANSWER_RESULT } from '@/constants';
import type { ExamRecord } from '@/types';

function Review() {
  const params = useSearchParams();
  const recordId = params.get('recordId') ?? '';
  const { isTeacher, isAdmin, isStudent } = useAuth();
  const canGrade = isTeacher || isAdmin;

  const [record, setRecord] = useState<ExamRecord | null>(null);
  const [scores, setScores] = useState<Record<string, number>>({});
  const [comments, setComments] = useState<Record<string, string>>({});
  const [grading, setGrading] = useState(false);
  const [adding, setAdding] = useState(false);

  const load = useCallback(async () => {
    if (!recordId) return;
    const r = await recordApi.get(recordId);
    setRecord(r);
    const sc: Record<string, number> = {};
    const cm: Record<string, string> = {};
    r.questions.forEach((q) => {
      if (q.subjective_score) sc[q.question_id] = q.subjective_score;
      if (q.comment) cm[q.question_id] = q.comment;
    });
    setScores(sc);
    setComments(cm);
  }, [recordId]);

  useEffect(() => {
    load();
  }, [load]);

  if (!record) return <div className="p-10 text-center text-gray-400">加载中…</div>;

  const resultColor = (r: string) => {
    switch (r) {
      case ANSWER_RESULT.CORRECT: return 'green';
      case ANSWER_RESULT.WRONG: return 'red';
      case ANSWER_RESULT.PARTIAL: return 'orange';
      default: return 'gray';
    }
  };

  const onGrade = async () => {
    setGrading(true);
    try {
      const grades = record.questions
        .filter((q) => q.type === 'fill' || q.type === 'short')
        .map((q) => ({ question_id: q.question_id, score: scores[q.question_id] ?? 0, comment: comments[q.question_id] ?? '' }));
      await recordApi.grade(record.id, grades);
      alert('批改完成');
      load();
    } catch (err) {
      alert((err as Error).message);
    } finally {
      setGrading(false);
    }
  };

  const onAddWrong = async (q: { question_id: string }) => {
    setAdding(true);
    try {
      await wrongBookApi.add({ question_id: q.question_id, exam_id: record.exam_id, exam_record_id: record.id });
      alert('已加入错题本');
    } catch (err) {
      alert((err as Error).message);
    } finally {
      setAdding(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-bold text-gray-800">{record.exam_title} · 答卷详情</h1>
          <p className="mt-1 text-sm text-gray-500">
            学生 {record.student_name} · 客观题 {record.objective_score} 分 · 最终 {record.final_score || '-'} 分 · 切屏 {record.cheat_count} 次
          </p>
          <p className="text-xs text-gray-400">开始 {formatDateTime(record.started_at)}</p>
        </div>
        <StatusBadge text={recordStatusText(record.status)} color={recordStatusColor(record.status)} />
      </div>

      {record.questions.map((q, i) => (
        <div key={q.question_id} className="rounded-xl border border-gray-200 bg-white p-5">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-400">第 {i + 1} 题</span>
            <div className="flex items-center gap-2">
              <QuestionTypeBadge type={q.type} />
              <StatusBadge text={answerResultText(q.result)} color={resultColor(q.result)} />
            </div>
          </div>
          <p className="mt-2 font-medium text-gray-800">{q.content}</p>
          <p className="mt-1 text-xs text-gray-400">{q.score} 分 · 本题得分 {q.got_score || 0}</p>

          {(q.type === 'single' || q.type === 'multiple' || q.type === 'judge') && (
            <div className="mt-3 space-y-1">
              {q.type === 'judge'
                ? ['true', 'false'].map((v) => (
                    <p key={v} className="text-sm">
                      {v === 'true' ? '正确' : '错误'}
                      {q.correct_answer === v && <span className="ml-2 text-green-600">✓ 正确答案</span>}
                      {q.user_answer === v && <span className={`ml-2 ${q.user_answer === q.correct_answer ? 'text-green-600' : 'text-red-600'}`}>我的答案</span>}
                    </p>
                  ))
                : (q.options ?? []).map((o) => (
                    <p key={o.key} className="text-sm">
                      {o.key}. {o.text}
                      {q.correct_answer.split(',').includes(o.key) && <span className="ml-2 text-green-600">✓ 正确答案</span>}
                      {(q.user_answer ?? '').split(',').includes(o.key) && (
                        <span className={`ml-2 ${(q.correct_answer ?? '').split(',').includes(o.key) ? 'text-green-600' : 'text-red-600'}`}>我的答案</span>
                      )}
                    </p>
                  ))}
            </div>
          )}

          {q.type === 'fill' || q.type === 'short' ? (
            <div className="mt-3 rounded-lg bg-gray-50 p-3 text-sm">
              <p><span className="text-gray-500">我的答案：</span>{q.user_answer || '（未作答）'}</p>
              <p className="mt-1"><span className="text-gray-500">参考答案：</span>{q.correct_answer}</p>
              {canGrade && (
                <div className="mt-3 grid gap-2 sm:grid-cols-2">
                  <input type="number" min={0} max={q.score} step={0.5} value={scores[q.question_id] ?? ''}
                    placeholder="给分"
                    onChange={(e) => setScores({ ...scores, [q.question_id]: Number(e.target.value) })}
                    className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm" />
                  <input value={comments[q.question_id] ?? ''} placeholder="评语（可选）"
                    onChange={(e) => setComments({ ...comments, [q.question_id]: e.target.value })}
                    className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm" />
                </div>
              )}
            </div>
          ) : (
            <div className="mt-2 text-sm">
              <p><span className="text-gray-500">我的答案：</span>{q.user_answer || '（未作答）'}</p>
            </div>
          )}

          <div className="mt-3 flex items-center justify-between border-t border-gray-100 pt-3">
            <p className="text-xs text-gray-400">本题得分 {q.got_score || 0}/{q.score}</p>
            {isStudent && q.result === ANSWER_RESULT.WRONG && (
              <button onClick={() => onAddWrong(q)} disabled={adding}
                className="rounded-lg border border-amber-500 px-3 py-1 text-xs text-amber-600 hover:bg-amber-50 disabled:opacity-60">
                加入错题本
              </button>
            )}
          </div>
        </div>
      ))}

      {canGrade && record.status !== 'graded' && (
        <div className="flex justify-end">
          <button onClick={onGrade} disabled={grading}
            className="rounded-lg bg-brand-600 px-6 py-2 text-sm text-white hover:bg-brand-700 disabled:opacity-60">
            {grading ? '批改中…' : '保存批改'}
          </button>
        </div>
      )}
    </div>
  );
}

export default function ReviewPage() {
  return (
    <Suspense fallback={<div className="p-10 text-center text-gray-400">加载中…</div>}>
      <Review />
    </Suspense>
  );
}
