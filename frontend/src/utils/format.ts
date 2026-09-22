// 格式化工具（与后端 util/formatters.go 对应，枚举文案需前后端同步）。
import {
  ANSWER_RESULT,
  DIFFICULTY,
  EXAM_STATUS,
  QUESTION_TYPES,
  RECORD_STATUS,
  ROLES,
  USER_STATUS,
} from '@/constants';

export function formatDateTime(v?: string | null): string {
  if (!v) return '-';
  const d = new Date(v);
  if (Number.isNaN(d.getTime())) return '-';
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

export function roleText(role: string): string {
  switch (role) {
    case ROLES.ADMIN: return '管理员';
    case ROLES.TEACHER: return '教师';
    case ROLES.STUDENT: return '学生';
    default: return role;
  }
}

export function questionTypeText(t: string): string {
  switch (t) {
    case QUESTION_TYPES.SINGLE: return '单选题';
    case QUESTION_TYPES.MULTIPLE: return '多选题';
    case QUESTION_TYPES.JUDGE: return '判断题';
    case QUESTION_TYPES.FILL: return '填空题';
    case QUESTION_TYPES.SHORT: return '简答题';
    default: return t;
  }
}

export function difficultyText(d: string): string {
  switch (d) {
    case DIFFICULTY.EASY: return '容易';
    case DIFFICULTY.MEDIUM: return '中等';
    case DIFFICULTY.HARD: return '困难';
    default: return d;
  }
}

export function examStatusText(s: string): string {
  switch (s) {
    case EXAM_STATUS.DRAFT: return '草稿';
    case EXAM_STATUS.PUBLISHED: return '已发布';
    case EXAM_STATUS.ONGOING: return '进行中';
    case EXAM_STATUS.FINISHED: return '已结束';
    case EXAM_STATUS.CLOSED: return '已关闭';
    default: return s;
  }
}

export function recordStatusText(s: string): string {
  switch (s) {
    case RECORD_STATUS.IN_PROGRESS: return '答题中';
    case RECORD_STATUS.SUBMITTED: return '已提交';
    case RECORD_STATUS.GRADED: return '已批改';
    default: return s;
  }
}

export function answerResultText(r: string): string {
  switch (r) {
    case ANSWER_RESULT.CORRECT: return '正确';
    case ANSWER_RESULT.WRONG: return '错误';
    case ANSWER_RESULT.PARTIAL: return '部分得分';
    case ANSWER_RESULT.UNMARKED: return '未批改';
    default: return '未作答';
  }
}

export function userStatusText(s: string): string {
  return s === USER_STATUS.ACTIVE ? '正常' : '已禁用';
}

export function examStatusColor(s: string): string {
  switch (s) {
    case EXAM_STATUS.DRAFT: return 'gray';
    case EXAM_STATUS.PUBLISHED: return 'blue';
    case EXAM_STATUS.ONGOING: return 'green';
    case EXAM_STATUS.FINISHED: return 'orange';
    case EXAM_STATUS.CLOSED: return 'red';
    default: return 'gray';
  }
}

export function recordStatusColor(s: string): string {
  switch (s) {
    case RECORD_STATUS.IN_PROGRESS: return 'orange';
    case RECORD_STATUS.SUBMITTED: return 'blue';
    case RECORD_STATUS.GRADED: return 'green';
    default: return 'gray';
  }
}
