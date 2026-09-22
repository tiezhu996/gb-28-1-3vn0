// 状态徽标：考试/记录状态、题型、难度、角色等（跨页面复用）。
'use client';

const COLOR_MAP: Record<string, string> = {
  gray: 'bg-gray-100 text-gray-700',
  blue: 'bg-blue-100 text-blue-700',
  green: 'bg-green-100 text-green-700',
  orange: 'bg-orange-100 text-orange-700',
  red: 'bg-red-100 text-red-700',
  purple: 'bg-purple-100 text-purple-700',
};

export function StatusBadge({ text, color = 'gray' }: { text: string; color?: string }) {
  const cls = COLOR_MAP[color] ?? COLOR_MAP.gray;
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${cls}`}>
      {text}
    </span>
  );
}

export function RoleBadge({ role }: { role: string }) {
  const map: Record<string, [string, string]> = {
    admin: ['管理员', 'purple'],
    teacher: ['教师', 'blue'],
    student: ['学生', 'green'],
  };
  const [text, color] = map[role] ?? [role, 'gray'];
  return <StatusBadge text={text} color={color} />;
}

export function QuestionTypeBadge({ type }: { type: string }) {
  const map: Record<string, [string, string]> = {
    single: ['单选题', 'blue'],
    multiple: ['多选题', 'purple'],
    judge: ['判断题', 'orange'],
    fill: ['填空题', 'green'],
    short: ['简答题', 'red'],
  };
  const [text, color] = map[type] ?? [type, 'gray'];
  return <StatusBadge text={text} color={color} />;
}

export function DifficultyBadge({ difficulty }: { difficulty: string }) {
  const map: Record<string, [string, string]> = {
    easy: ['容易', 'green'],
    medium: ['中等', 'orange'],
    hard: ['困难', 'red'],
  };
  const [text, color] = map[difficulty] ?? [difficulty, 'gray'];
  return <StatusBadge text={text} color={color} />;
}
