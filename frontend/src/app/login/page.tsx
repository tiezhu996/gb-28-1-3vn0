'use client';
import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';

export default function LoginPage() {
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const quickFill = (em: string) => {
    setEmail(em);
    setPassword(em.startsWith('admin') ? 'admin123456' : em.startsWith('teacher') ? 'teacher123456' : 'student123456');
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      router.push('/');
    } catch (err) {
      setError((err as Error).message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto mt-10 max-w-md rounded-2xl border border-gray-200 bg-white p-8 shadow-sm">
      <h1 className="text-xl font-bold text-gray-800">登录</h1>
      <p className="mt-1 text-sm text-gray-400">欢迎回到在线考试系统</p>
      <form onSubmit={onSubmit} className="mt-6 space-y-4">
        <div>
          <label className="text-sm text-gray-600">邮箱</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none"
            placeholder="student@onlineexam.com"
          />
        </div>
        <div>
          <label className="text-sm text-gray-600">密码</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none"
            placeholder="••••••"
          />
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-lg bg-brand-600 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-60"
        >
          {loading ? '登录中…' : '登录'}
        </button>
      </form>
      <div className="mt-5 rounded-lg bg-gray-50 p-3 text-xs text-gray-500">
        <p className="mb-1 font-medium text-gray-600">演示账号（点击快速填充）</p>
        <div className="flex flex-wrap gap-2">
          <button onClick={() => quickFill('admin@onlineexam.com')} className="rounded bg-purple-100 px-2 py-1 text-purple-700">管理员</button>
          <button onClick={() => quickFill('teacher@onlineexam.com')} className="rounded bg-blue-100 px-2 py-1 text-blue-700">教师</button>
          <button onClick={() => quickFill('student@onlineexam.com')} className="rounded bg-green-100 px-2 py-1 text-green-700">学生</button>
        </div>
      </div>
      <p className="mt-5 text-center text-sm text-gray-500">
        还没有账号？<Link href="/register" className="text-brand-600 hover:underline">去注册</Link>
      </p>
    </div>
  );
}
