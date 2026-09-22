// 分页 hook：页码/页大小状态管理。
'use client';
import { useCallback, useState } from 'react';

export function usePagination(defaultPageSize = 10) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultPageSize);

  const reset = useCallback(() => setPage(1), []);
  const next = useCallback(() => setPage((p) => p + 1), []);
  const prev = useCallback(() => setPage((p) => Math.max(1, p - 1)), []);

  return { page, pageSize, setPage, setPageSize, reset, next, prev };
}
