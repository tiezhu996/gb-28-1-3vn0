import { request, buildQuery } from '@/utils/request';
import type { AuditLog, PageResult } from '@/types';

export const auditApi = {
  list(query: { module?: string; action?: string; keyword?: string; page?: number; page_size?: number }) {
    return request<PageResult<AuditLog>>(`/audit-logs${buildQuery({ ...query })}`);
  },
};
