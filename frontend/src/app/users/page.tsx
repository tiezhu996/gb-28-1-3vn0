'use client';
import { useCallback, useEffect, useState } from 'react';
import { userApi } from '@/api/user';
import { DataTable, type Column } from '@/components/DataTable';
import { Pagination } from '@/components/Pagination';
import { Modal } from '@/components/Modal';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { RoleBadge, StatusBadge } from '@/components/StatusBadge';
import { formatDateTime, roleText } from '@/utils/format';
import { ROLES } from '@/constants';
import type { User } from '@/types';

export default function UsersPage() {
  const [list, setList] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [role, setRole] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [confirmId, setConfirmId] = useState<string | null>(null);
  const [form, setForm] = useState<{ name: string; email: string; password: string; role: string }>({ name: '', email: '', password: '', role: ROLES.STUDENT });
  const [saving, setSaving] = useState(false);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const res = await userApi.list({ role, keyword, page, page_size: 10 });
      setList(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [role, keyword, page]);

  useEffect(() => {
    reload();
  }, [reload]);

  const onCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await userApi.create(form);
      setCreateOpen(false);
      setForm({ name: '', email: '', password: '', role: ROLES.STUDENT });
      reload();
    } catch (err) {
      alert((err as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const toggleStatus = async (u: User) => {
    await userApi.update(u.id, { status: u.status === 'active' ? 'disabled' : 'active' });
    reload();
  };

  const columns: Column<User>[] = [
    { key: 'name', title: '姓名', render: (u) => <span className="font-medium">{u.name}</span> },
    { key: 'email', title: '邮箱', render: (u) => <span className="text-xs">{u.email}</span> },
    { key: 'role', title: '角色', render: (u) => <RoleBadge role={u.role} /> },
    { key: 'status', title: '状态', render: (u) => (
        <StatusBadge text={u.status === 'active' ? '正常' : '已禁用'} color={u.status === 'active' ? 'green' : 'red'} />
      ) },
    { key: 'created_at', title: '创建时间', render: (u) => <span className="text-xs">{formatDateTime(u.created_at)}</span> },
    { key: 'actions', title: '操作', render: (u) => (
        <div className="flex gap-2">
          <button onClick={() => toggleStatus(u)} className="text-brand-600 hover:underline">
            {u.status === 'active' ? '禁用' : '启用'}
          </button>
          <button onClick={() => setConfirmId(u.id)} className="text-red-600 hover:underline">删除</button>
        </div>
      ) },
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-bold text-gray-800">用户管理</h1>
        <div className="flex gap-2">
          <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="搜索姓名/邮箱…"
            className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm" />
          <select value={role} onChange={(e) => setRole(e.target.value)}
            className="rounded-lg border border-gray-300 px-2 py-1.5 text-sm">
            <option value="">全部角色</option>
            <option value="student">学生</option>
            <option value="teacher">教师</option>
            <option value="admin">管理员</option>
          </select>
          <button onClick={() => setPage(1)} className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm hover:bg-gray-50">查询</button>
          <button onClick={() => setCreateOpen(true)}
            className="rounded-lg bg-brand-600 px-3 py-1.5 text-sm text-white hover:bg-brand-700">新增用户</button>
        </div>
      </div>
      <DataTable columns={columns} rows={list} loading={loading} emptyTitle="暂无用户" />
      <Pagination page={page} pageSize={10} total={total} onChange={setPage} />

      <Modal open={createOpen} title="新增用户" onClose={() => setCreateOpen(false)}>
        <form onSubmit={onCreate} className="space-y-3">
          <input required minLength={2} placeholder="姓名" value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          <input required type="email" placeholder="邮箱" value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          <input required type="password" minLength={6} placeholder="密码" value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm" />
          <select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
            {Object.values(ROLES).map((r) => <option key={r} value={r}>{roleText(r)}</option>)}
          </select>
          <button type="submit" disabled={saving}
            className="w-full rounded-lg bg-brand-600 py-2 text-sm text-white hover:bg-brand-700 disabled:opacity-60">
            {saving ? '创建中…' : '创建'}
          </button>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!confirmId}
        message="确定删除该用户？"
        onCancel={() => setConfirmId(null)}
        onConfirm={async () => {
          if (confirmId) {
            await userApi.remove(confirmId);
            setConfirmId(null);
            reload();
          }
        }}
      />
    </div>
  );
}
