'use client';
import { Suspense, useCallback, useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { examApi } from '@/api/exam';
import { recordApi } from '@/api/record';
import { useExamStore } from '@/stores/examStore';
import { questionApi } from '@/api/question';
import { DataTable, type Column } from '@/components/DataTable';
import { Pagination } from '@/components/Pagination';
import { StatusBadge } from '@/components/StatusBadge';
import { examStatusColor, examStatusText, formatDateTime, questionTypeText } from '@/utils/format';
import { EXAM_STATUS } from '@/constants';
import type { Exam, ExamRecord, Question } from '@/types';

function ExamDetail() {
  const router = useRouter();
  const params = useSearchParams();
  const id = params.get('id') ?? '';
  const { isTeacher, isAdmin, isStudent } = useAuth();
  const canManage = isTeacher || isAdmin;
  const { publish, close, remove } = useExamStore();

  const [exam, setExam] = useState<Exam | null>(null);
  const [questions, setQuestions] = useState<Record<string, Question>>({});
  const [records, setRecords] = useState<ExamRecord[]>([]);
  const [recordsTotal, setRecordsTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [starting, setStarting] = useState(false);

  const load = useCallback(async () => {
    if (!id) return;
    const e = await examApi.get(id);
    setExam(e);
    const qids = e.questions.map((q) => q.question_id);
    const qmap: Record<string, Question> = {};
    for (const qid of qids) {
      try {
        const q = await questionApi.get(qid);
        qmap[qid] = q;
      } catch {
        // 忽略已删除题目
      }
    }
    setQuestions(qmap);
    if (canManage) {
      const res = await recordApi.listByExam(id, { page, page_size: 10 });
      setRecords(res.list);
      setRecordsTotal(res.total);
    }
  }, [id, canManage, page]);

  useEffect(() => {
    load();
  }, [load]);

  const onStart = async () => {
    if (!id) return;
    setStarting(true);
    try {
      const rec = await recordApi.start(id);
      router.push(`/exam-take?recordId=${rec.id}`);
    } catch (err) {
      alert((err as Error).message);
    } finally {
      setStarting(false);
    }
  };

  if (!exam) return <div className="p-10 text-center text-gray-400">加载中…</div>;

  const recordColumns: Column<ExamRecord>[] = [
    { key: 'student_name', title: '学生', render: (r) => <span>{r.student_name}</span> },
    { key: 'status', title: '状态', render: (r) => <StatusBadge text={r.status === 'graded' ? '已批改' : r.status === 'submitted' ? '已提交' : '答题中'} color={r.status === 'graded' ? 'green' : r.status === 'submitted' ? 'blue' : 'orange'} /> },
    { key: 'objective_score', title: '客观题分', render: (r) => <span>{r.objective_score}</span> },
    { key: 'final_score', title: '最终分', render: (r) => <span className="font-medium">{r.final_score || '-'}</span> },
    { key: 'cheat_count', title: '切屏次数', render: (r) => <span className={r.cheat_count > 0 ? 'text-red-600' : ''}>{r.cheat_count}</span> },
    { key: 'started_at', title: '开始时间', render: (r) => <span className="text-xs">{formatDateTime(r.started_at)}</span> },
    { key: 'actions', title: '操作', render: (r) => (
        <button onClick={() => router.push(`/records/review?recordId=${r.id}`)} className="text-brand-600 hover:underline">查看/批改</button>
      ) },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-xl font-bold text-gray-800">{exam.title}</h1>
          <p className="mt-1 text-sm text-gray-500">
            {exam.subject} · 总分 {exam.total_score} · 及格 {exam.pass_score} · 时长 {exam.duration_min} 分钟
          </p>
          <p className="mt-1 text-xs text-gray-400">
            {formatDateTime(exam.start_at)} ~ {formatDateTime(exam.end_at)}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <StatusBadge text={examStatusText(exam.status)} color={examStatusColor(exam.status)} />
          {isStudent && exam.status === EXAM_STATUS.PUBLISHED && (
            <button onClick={onStart} disabled={starting}
              className="rounded-lg bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700 disabled:opacity-60">
              {starting ? '进入中…' : '开始考试'}
            </button>
          )}
          {canManage && (
            <>
              {exam.status === EXAM_STATUS.DRAFT && (
                <button onClick={async () => { await publish(id); load(); }}
                  className="rounded-lg bg-brand-600 px-4 py-2 text-sm text-white hover:bg-brand-700">发布</button>
              )}
              {exam.status !== EXAM_STATUS.CLOSED && (
                <button onClick={async () => { await close(id); load(); }}
                  className="rounded-lg border border-gray-300 px-4 py-2 text-sm hover:bg-gray-50">关闭</button>
              )}
              <button onClick={async () => { if (confirm('确认删除？')) { await remove(id); router.push('/exams'); } }}
                className="rounded-lg border border-red-300 px-4 py-2 text-sm text-red-600 hover:bg-red-50">删除</button>
              <button onClick={() => router.push(`/reports?examId=${id}`)}
                className="rounded-lg border border-brand-600 px-4 py-2 text-sm text-brand-600 hover:bg-brand-50">成绩分析</button>
            </>
          )}
        </div>
      </div>

      <section className="rounded-xl border border-gray-200 bg-white p-5">
        <h2 className="font-semibold text-gray-800">试卷题目（{exam.questions.length} 题）</h2>
        <div className="mt-3 space-y-2">
          {exam.questions.map((eq) => {
            const q = questions[eq.question_id];
            return (
              <div key={eq.question_id} className="flex items-center gap-3 rounded-lg bg-gray-50 px-3 py-2 text-sm">
                <span className="w-6 text-center font-medium text-gray-400">{eq.order}</span>
                <span className="flex-1 truncate">{q ? q.content : eq.question_id}</span>
                <span className="text-xs text-gray-400">{q ? questionTypeText(q.type) : ''} · {eq.score}分</span>
              </div>
            );
          })}
        </div>
      </section>

      {canManage && (
        <section className="rounded-xl border border-gray-200 bg-white p-5">
          <h2 className="font-semibold text-gray-800">考生记录</h2>
          <div className="mt-3">
            <DataTable columns={recordColumns} rows={records} emptyTitle="暂无考生作答" />
            <Pagination page={page} pageSize={10} total={recordsTotal} onChange={setPage} />
          </div>
        </section>
      )}
    </div>
  );
}

export default function ExamDetailPage() {
  return (
    <Suspense fallback={<div className="p-10 text-center text-gray-400">加载中…</div>}>
      <ExamDetail />
    </Suspense>
  );
}
