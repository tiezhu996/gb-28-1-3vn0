'use client';
import { useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useExamStore } from '@/stores/examStore';
import { DataTable, type Column } from '@/components/DataTable';
import { Pagination } from '@/components/Pagination';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { StatusBadge } from '@/components/StatusBadge';
import { examStatusColor, examStatusText, formatDateTime } from '@/utils/format';
import { EXAM_STATUS, SUBJECTS } from '@/constants';
import type { Exam } from '@/types';

export default function ExamsPage() {
  const { isTeacher, isAdmin, isStudent } = useAuth();
  const canWrite = isTeacher || isAdmin;
  const { list, total, loading, fetch, remove } = useExamStore();
  const [page, setPage] = useState(1);
  const [subject, setSubject] = useState('');
  const [status, setStatus] = useState('');
  const [confirmId, setConfirmId] = useState<string | null>(null);

  const reload = useCallback(() => {
    fetch({ subject, status, page, page_size: 10 });
  }, [fetch, subject, status, page]);

  useEffect(() => {
    reload();
  }, [reload]);

  const columns: Column<Exam>[] = [
    { key: 'title', title: '考试名称', render: (e) => (
        <Link href={`/exams/detail?id=${e.id}`} className="font-medium text-brand-600 hover:underline">{e.title}</Link>
      ) },
    { key: 'subject', title: '学科', render: (e) => <span className="text-xs text-gray-500">{e.subject}</span> },
    { key: 'total_score', title: '总分', render: (e) => <span>{e.total_score}</span> },
    { key: 'duration_min', title: '时长(分)', render: (e) => <span>{e.duration_min}</span> },
    { key: 'questions', title: '题数', render: (e) => <span>{e.questions.length}</span> },
    { key: 'status', title: '状态', render: (e) => <StatusBadge text={examStatusText(e.status)} color={examStatusColor(e.status)} /> },
    { key: 'start_at', title: '开始时间', render: (e) => <span className="text-xs">{formatDateTime(e.start_at)}</span> },
    { key: 'end_at', title: '结束时间', render: (e) => <span className="text-xs">{formatDateTime(e.end_at)}</span> },
    ...(canWrite
      ? [{
          key: 'actions',
          title: '操作',
          render: (e: Exam) => (
            <div className="flex gap-2 text-sm">
              <Link href={`/exams/detail?id=${e.id}`} className="text-brand-600 hover:underline">管理</Link>
              <button onClick={() => setConfirmId(e.id)} className="text-red-600 hover:underline">删除</button>
            </div>
          ),
        } as Column<Exam>]
      : []),
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-bold text-gray-800">考试管理</h1>
        <div className="flex flex-wrap gap-2">
          <select value={subject} onChange={(e) => setSubject(e.target.value)}
            className="rounded-lg border border-gray-300 px-2 py-1.5 text-sm">
            <option value="">全部学科</option>
            {SUBJECTS.map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
          <select value={status} onChange={(e) => setStatus(e.target.value)}
            className="rounded-lg border border-gray-300 px-2 py-1.5 text-sm">
            <option value="">全部状态</option>
            {Object.values(EXAM_STATUS).map((s) => <option key={s} value={s}>{examStatusText(s)}</option>)}
          </select>
          <button onClick={() => setPage(1)} className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm hover:bg-gray-50">查询</button>
          {canWrite && (
            <Link href="/exams/new" className="rounded-lg bg-brand-600 px-3 py-1.5 text-sm text-white hover:bg-brand-700">
              {isStudent ? '' : '创建考试'}
            </Link>
          )}
        </div>
      </div>

      <DataTable columns={columns} rows={list} loading={loading} emptyTitle="暂无考试" />
      <Pagination page={page} pageSize={10} total={total} onChange={setPage} />

      <ConfirmDialog
        open={!!confirmId}
        message="确定删除该考试？"
        onCancel={() => setConfirmId(null)}
        onConfirm={async () => {
          if (confirmId) {
            await remove(confirmId);
            setConfirmId(null);
            reload();
          }
        }}
      />
    </div>
  );
}
