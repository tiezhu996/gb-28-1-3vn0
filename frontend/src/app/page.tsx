'use client';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';

export default function HomePage() {
  const { isAuthenticated, isTeacher, isStudent, isAdmin } = useAuth();

  const cards: { href: string; title: string; desc: string; show: boolean }[] = [
    { href: '/questions', title: '题库管理', desc: '录入单选/多选/判断/填空/简答，支持 Excel 批量导入', show: true },
    { href: '/exams', title: '智能组卷', desc: '手动选题或按知识点与难度分布自动组卷', show: true },
    { href: '/exam-take', title: '在线考试', desc: '倒计时答题、题目导航、切屏防作弊检测', show: isStudent },
    { href: '/records', title: '成绩与报告', desc: '自动阅卷、教师批改、成绩分析与错题回顾', show: true },
    { href: '/wrongbook', title: '错题本', desc: '按知识点归类复习错题', show: isStudent },
    { href: '/audit', title: '审计日志', desc: '操作审计留痕', show: isAdmin },
  ];

  return (
    <div>
      <section className="rounded-2xl bg-gradient-to-br from-brand-600 to-blue-800 p-8 text-white">
        <h1 className="text-3xl font-bold">在线考试系统</h1>
        <p className="mt-3 max-w-2xl text-blue-100">
          面向高校与培训机构的一站式在线考试平台：题库管理、智能组卷、在线答题、自动阅卷、防作弊与成绩分析。
        </p>
        <div className="mt-6 flex gap-3">
          {!isAuthenticated ? (
            <>
              <Link href="/register" className="rounded-lg bg-white px-5 py-2 text-sm font-medium text-brand-700">
                立即注册
              </Link>
              <Link href="/login" className="rounded-lg border border-white/60 px-5 py-2 text-sm font-medium text-white">
                登录
              </Link>
            </>
          ) : (
            <Link href="/exams" className="rounded-lg bg-white px-5 py-2 text-sm font-medium text-brand-700">
              进入考试中心
            </Link>
          )}
        </div>
      </section>

      <section className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {cards
          .filter((c) => c.show)
          .map((c) => (
            <Link
              key={c.href}
              href={c.href}
              className="rounded-xl border border-gray-200 bg-white p-5 transition hover:border-brand-500 hover:shadow-md"
            >
              <h3 className="font-semibold text-gray-800">{c.title}</h3>
              <p className="mt-2 text-sm text-gray-500">{c.desc}</p>
            </Link>
          ))}
      </section>
    </div>
  );
}
