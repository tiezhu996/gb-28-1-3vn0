// 认证上下文：登录态初始化 + 路由守卫（ProtectedRoute）。
'use client';
import { createContext, useContext, useEffect, useMemo, type ReactNode } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import { ROLES } from '@/constants';

interface AuthContextValue {
  isAuthenticated: boolean;
  isAdmin: boolean;
  isTeacher: boolean;
  isStudent: boolean;
  requireRoles: (...roles: string[]) => boolean;
}

const AuthContext = createContext<AuthContextValue>({
  isAuthenticated: false,
  isAdmin: false,
  isTeacher: false,
  isStudent: false,
  requireRoles: () => false,
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const init = useAuthStore((s) => s.init);
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);

  useEffect(() => {
    init();
  }, [init]);

  const value = useMemo<AuthContextValue>(() => {
    const role = user?.role ?? '';
    return {
      isAuthenticated: !!token && !!user,
      isAdmin: role === ROLES.ADMIN,
      isTeacher: role === ROLES.TEACHER,
      isStudent: role === ROLES.STUDENT,
      requireRoles: (...roles: string[]) => roles.includes(role),
    };
  }, [token, user]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}

// 路由守卫：未登录跳登录页；角色不匹配跳首页。
export function ProtectedRoute({
  children,
  roles,
  fallback = '/login',
}: {
  children: ReactNode;
  roles?: string[];
  fallback?: string;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, requireRoles } = useAuth();

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace(fallback);
      return;
    }
    if (roles && !requireRoles(...roles)) {
      router.replace('/');
    }
  }, [isAuthenticated, roles, requireRoles, router, pathname, fallback]);

  if (!isAuthenticated) {
    return <div className="flex min-h-screen items-center justify-center text-gray-400">请先登录…</div>;
  }
  if (roles && !requireRoles(...roles)) {
    return <div className="flex min-h-screen items-center justify-center text-gray-400">无权限访问…</div>;
  }
  return <>{children}</>;
}
