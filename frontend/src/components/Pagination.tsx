// 分页控件（跨页面复用）。
'use client';
export function Pagination({
  page,
  pageSize,
  total,
  onChange,
}: {
  page: number;
  pageSize: number;
  total: number;
  onChange: (page: number) => void;
}) {
  const totalPage = Math.max(1, Math.ceil(total / pageSize));
  return (
    <div className="mt-4 flex items-center justify-between text-sm text-gray-600">
      <span>
        共 {total} 条 · 第 {page}/{totalPage} 页
      </span>
      <div className="flex gap-2">
        <button
          disabled={page <= 1}
          onClick={() => onChange(page - 1)}
          className="rounded-lg border border-gray-300 px-3 py-1 hover:bg-gray-50 disabled:opacity-40"
        >
          上一页
        </button>
        <button
          disabled={page >= totalPage}
          onClick={() => onChange(page + 1)}
          className="rounded-lg border border-gray-300 px-3 py-1 hover:bg-gray-50 disabled:opacity-40"
        >
          下一页
        </button>
      </div>
    </div>
  );
}
