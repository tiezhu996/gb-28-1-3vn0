'use client';
import { useCallback, useEffect, useState } from 'react';
import { auditApi } from '@/api/audit';
import { DataTable, type Column } from '@/components/DataTable';
import { Pagination } from '@/components/Pagination';
import { RoleBadge, StatusBadge } from '@/components/StatusBadge';
import { formatDateTime } from '@/utils/format';
import type { AuditLog } from '@/types';

export default function AuditPage() {
  const [list, setList] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [module, setModule] = useState('');
  const [keyword, setKeyword] = useState('');

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const res = await auditApi.list({ module, keyword, page, page_size: 10 });
      setList(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [module, keyword, page]);

  useEffect(() => {
    reload();
  }, [reload]);

  const columns: Column<AuditLog>[] = [
    { key: 'username', title: '操作人', render: (a) => <span>{a.username || '-'}</span> },
    { key: 'role', title: '角色', render: (a) => <RoleBadge role={a.role} /> },
    { key: 'module', title: '模块', render: (a) => <span className="text-xs">{a.module}</span> },
    { key: 'action', title: '动作', render: (a) => (
        <StatusBadge text={a.action} color={a.status_code >= 400 ? 'red' : 'blue'} />
      ) },
    { key: 'method', title: '方法', render: (a) => <span className="font-mono text-xs">{a.method}</span> },
    { key: 'path', title: '路径', render: (a) => <span className="font-mono text-xs">{a.path}</span> },
    { key: 'status_code', title: '状态码', render: (a) => (
        <span className={a.status_code >= 400 ? 'text-red-600' : 'text-green-600'}>{a.status_code}</span>
      ) },
    { key: 'client_ip', title: 'IP', render: (a) => <span className="text-xs">{a.client_ip}</span> },
    { key: 'created_at', title: '时间', render: (a) => <span className="text-xs">{formatDateTime(a.created_at)}</span> },
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-bold text-gray-800">操作审计日志</h1>
        <div className="flex gap-2">
          <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="搜索操作人/路径…"
            className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm" />
          <select value={module} onChange={(e) => setModule(e.target.value)}
            className="rounded-lg border border-gray-300 px-2 py-1.5 text-sm">
            <option value="">全部模块</option>
            <option value="users">用户</option>
            <option value="questions">题库</option>
            <option value="exams">试卷</option>
            <option value="exam-records">考试记录</option>
            <option value="wrong-books">错题本</option>
            <option value="audit-logs">审计</option>
          </select>
          <button onClick={() => setPage(1)} className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm hover:bg-gray-50">查询</button>
        </div>
      </div>
      <DataTable columns={columns} rows={list} loading={loading} emptyTitle="暂无审计日志" />
      <Pagination page={page} pageSize={10} total={total} onChange={setPage} />
    </div>
  );
}
