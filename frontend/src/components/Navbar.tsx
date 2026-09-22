// 顶部导航：按角色显示菜单（与后端 RBAC 对应，按钮显隐联动）。
'use client';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import { useAuth } from '@/hooks/useAuth';
import { ROLES } from '@/constants';

export function Navbar() {
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const { isAuthenticated, isTeacher, isStudent, isAdmin } = useAuth();

  const links: { href: string; label: string; show: boolean }[] = [
    { href: '/', label: '首页', show: true },
    { href: '/questions', label: '题库', show: true },
    { href: '/exams', label: '考试', show: true },
    { href: '/records', label: '我的考试', show: isStudent },
    { href: '/wrongbook', label: '错题本', show: isStudent },
    { href: '/users', label: '用户管理', show: isAdmin },
    { href: '/audit', label: '审计日志', show: isAdmin },
  ];

  return (
    <header className="sticky top-0 z-40 border-b border-gray-200 bg-white">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        <div className="flex items-center gap-6">
          <Link href="/" className="text-lg font-bold text-brand-600">
            在线考试系统
          </Link>
          <nav className="hidden items-center gap-1 md:flex">
            {links
              .filter((l) => l.show)
              .map((l) => (
                <Link
                  key={l.href}
                  href={l.href}
                  className="rounded-lg px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                >
                  {l.label}
                </Link>
              ))}
          </nav>
        </div>
        <div className="flex items-center gap-3">
          {isAuthenticated && user ? (
            <>
              <span className="text-sm text-gray-600">
                {user.name}
                <span className="ml-1 text-xs text-gray-400">
                  {user.role === ROLES.ADMIN ? '管理员' : user.role === ROLES.TEACHER ? '教师' : '学生'}
                </span>
              </span>
              <button
                onClick={() => {
                  logout();
                  router.push('/login');
                }}
                className="rounded-lg border border-gray-300 px-3 py-1 text-sm text-gray-600 hover:bg-gray-50"
              >
                退出
              </button>
            </>
          ) : (
            <Link
              href="/login"
              className="rounded-lg bg-brand-600 px-4 py-1.5 text-sm text-white hover:bg-brand-700"
            >
              登录
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
