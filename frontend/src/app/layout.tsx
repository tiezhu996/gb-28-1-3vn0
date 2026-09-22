import type { Metadata } from 'next';
import './globals.css';
import { AuthProvider } from '@/hooks/useAuth';
import { Navbar } from '@/components/Navbar';

export const metadata: Metadata = {
  title: '在线考试系统',
  description: '支持题库管理、智能组卷、在线答题、自动阅卷的在线考试平台',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>
        <AuthProvider>
          <Navbar />
          <main className="mx-auto min-h-[calc(100vh-56px)] max-w-6xl px-4 py-6">{children}</main>
        </AuthProvider>
      </body>
    </html>
  );
}
