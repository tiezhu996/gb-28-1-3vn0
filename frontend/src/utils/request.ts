// 统一请求封装：注入 JWT、统一错误码处理、请求追踪（与后端 middleware/request_id.go、error_handler.go 对应）。
const BASE = '/api/v1';

export class ApiError extends Error {
  code: number;
  status: number;
  constructor(code: number, message: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

let authToken: string | null = null;
export function setAuthToken(token: string | null) {
  authToken = token;
}
export function getAuthToken() {
  return authToken;
}

interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

async function handle<T>(res: Response): Promise<T> {
  let body: ApiResponse<T>;
  try {
    body = await res.json();
  } catch {
    throw new ApiError(-1, '服务器响应格式异常', res.status);
  }
  if (body.code !== 0) {
    if (body.code === 1002) {
      // 登录过期：清理本地登录态
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem('onlineexam_token');
        window.localStorage.removeItem('onlineexam_user');
      }
    }
    throw new ApiError(body.code, body.message || '请求失败', res.status);
  }
  return body.data;
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  if (authToken) {
    headers.Authorization = `Bearer ${authToken}`;
  }
  const res = await fetch(`${BASE}${path}`, { ...options, headers, keepalive: options.keepalive });
  return handle<T>(res);
}

export async function upload<T>(path: string, file: File, extra: Record<string, string> = {}): Promise<T> {
  const form = new FormData();
  form.append('file', file);
  const headers: Record<string, string> = { ...extra };
  if (authToken) {
    headers.Authorization = `Bearer ${authToken}`;
  }
  const res = await fetch(`${BASE}${path}`, { method: 'POST', headers, body: form });
  return handle<T>(res);
}

export function buildQuery(params: Record<string, string | number | undefined>): string {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== '') qs.append(k, String(v));
  });
  const s = qs.toString();
  return s ? `?${s}` : '';
}
