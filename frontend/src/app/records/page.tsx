'use client';
import { useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import { useRecordStore } from '@/stores/recordStore';
import { DataTable, type Column } from '@/components/DataTable';
import { Pagination } from '@/components/Pagination';
import { StatusBadge } from '@/components/StatusBadge';
import { formatDateTime, recordStatusColor, recordStatusText } from '@/utils/format';
import type { ExamRecord } from '@/types';

export default function RecordsPage() {
  const { mine, total, loading, fetchMine } = useRecordStore();
  const [page, setPage] = useState(1);

  const reload = useCallback(() => {
    fetchMine({ page, page_size: 10 });
  }, [fetchMine, page]);

  useEffect(() => {
    reload();
  }, [reload]);

  const columns: Column<ExamRecord>[] = [
    { key: 'exam_title', title: '考试', render: (r) => <span className="font-medium">{r.exam_title}</span> },
    { key: 'status', title: '状态', render: (r) => <StatusBadge text={recordStatusText(r.status)} color={recordStatusColor(r.status)} /> },
    { key: 'objective_score', title: '客观题分', render: (r) => <span>{r.objective_score}</span> },
    { key: 'final_score', title: '最终分', render: (r) => <span className="font-semibold">{r.final_score || '-'}</span> },
    { key: 'cheat_count', title: '切屏次数', render: (r) => <span className={r.cheat_count > 0 ? 'text-red-600' : ''}>{r.cheat_count}</span> },
    { key: 'started_at', title: '开始时间', render: (r) => <span className="text-xs">{formatDateTime(r.started_at)}</span> },
    { key: 'actions', title: '操作', render: (r) => (
        <div className="flex gap-2">
          {r.status === 'in_progress' ? (
            <Link href={`/exam-take?recordId=${r.id}`} className="text-brand-600 hover:underline">继续作答</Link>
          ) : (
            <Link href={`/records/review?recordId=${r.id}`} className="text-brand-600 hover:underline">查看答卷</Link>
          )}
        </div>
      ) },
  ];

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-bold text-gray-800">我的考试记录</h1>
      <DataTable columns={columns} rows={mine} loading={loading} emptyTitle="还没有参加过考试" />
      <Pagination page={page} pageSize={10} total={total} onChange={setPage} />
    </div>
  );
}
