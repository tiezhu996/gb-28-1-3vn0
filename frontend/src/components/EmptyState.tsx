// 空状态占位（跨页面复用）。
'use client';
export function EmptyState({ title = '暂无数据', description }: { title?: string; description?: string }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-gray-300 bg-gray-50 py-14 text-center">
      <div className="text-4xl">📭</div>
      <p className="mt-3 text-sm font-medium text-gray-600">{title}</p>
      {description && <p className="mt-1 text-xs text-gray-400">{description}</p>}
    </div>
  );
}
