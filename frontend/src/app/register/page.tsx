'use client';
import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';

export default function RegisterPage() {
  const router = useRouter();
  const register = useAuthStore((s) => s.register);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('student');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await register({ name, email, password, role });
      router.push('/login');
    } catch (err) {
      setError((err as Error).message || '注册失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto mt-10 max-w-md rounded-2xl border border-gray-200 bg-white p-8 shadow-sm">
      <h1 className="text-xl font-bold text-gray-800">注册</h1>
      <p className="mt-1 text-sm text-gray-400">创建你的在线考试账号</p>
      <form onSubmit={onSubmit} className="mt-6 space-y-4">
        <div>
          <label className="text-sm text-gray-600">姓名</label>
          <input value={name} onChange={(e) => setName(e.target.value)} required minLength={2}
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none" />
        </div>
        <div>
          <label className="text-sm text-gray-600">邮箱</label>
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none" />
        </div>
        <div>
          <label className="text-sm text-gray-600">密码（至少 6 位）</label>
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={6}
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none" />
        </div>
        <div>
          <label className="text-sm text-gray-600">角色</label>
          <div className="mt-1 flex gap-4">
            {[
              { v: 'student', t: '学生' },
              { v: 'teacher', t: '教师' },
            ].map((r) => (
              <label key={r.v} className="flex items-center gap-1 text-sm text-gray-600">
                <input type="radio" checked={role === r.v} onChange={() => setRole(r.v)} />
                {r.t}
              </label>
            ))}
          </div>
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button type="submit" disabled={loading}
          className="w-full rounded-lg bg-brand-600 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-60">
          {loading ? '注册中…' : '注册'}
        </button>
      </form>
      <p className="mt-5 text-center text-sm text-gray-500">
        已有账号？<Link href="/login" className="text-brand-600 hover:underline">去登录</Link>
      </p>
    </div>
  );
}
