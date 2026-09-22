// 与后端 constants/enums.go 对应的业务枚举（前后端必须同步修改，新增枚举值需联动 ≥10 处文件）。
export const ROLES = {
  ADMIN: 'admin',
  TEACHER: 'teacher',
  STUDENT: 'student',
} as const;

export type Role = (typeof ROLES)[keyof typeof ROLES];

export const USER_STATUS = {
  ACTIVE: 'active',
  DISABLED: 'disabled',
} as const;

export const QUESTION_TYPES = {
  SINGLE: 'single',
  MULTIPLE: 'multiple',
  JUDGE: 'judge',
  FILL: 'fill',
  SHORT: 'short',
} as const;

export type QuestionType = (typeof QUESTION_TYPES)[keyof typeof QUESTION_TYPES];

export const DIFFICULTY = {
  EASY: 'easy',
  MEDIUM: 'medium',
  HARD: 'hard',
} as const;

export type Difficulty = (typeof DIFFICULTY)[keyof typeof DIFFICULTY];

export const EXAM_STATUS = {
  DRAFT: 'draft',
  PUBLISHED: 'published',
  ONGOING: 'ongoing',
  FINISHED: 'finished',
  CLOSED: 'closed',
} as const;

export type ExamStatus = (typeof EXAM_STATUS)[keyof typeof EXAM_STATUS];

// 状态机：与后端 constants.ExamStatusTransitions 对应（前端按钮显隐依赖此表）
export const EXAM_STATUS_TRANSITIONS: Record<string, string[]> = {
  draft: ['published', 'closed'],
  published: ['ongoing', 'closed'],
  ongoing: ['finished', 'closed'],
  finished: ['closed'],
  closed: [],
};

export const RECORD_STATUS = {
  IN_PROGRESS: 'in_progress',
  SUBMITTED: 'submitted',
  GRADED: 'graded',
} as const;

export type RecordStatus = (typeof RECORD_STATUS)[keyof typeof RECORD_STATUS];

export const ANSWER_RESULT = {
  CORRECT: 'correct',
  WRONG: 'wrong',
  PARTIAL: 'partial',
  UNMARKED: 'unmarked',
} as const;

export type AnswerResult = (typeof ANSWER_RESULT)[keyof typeof ANSWER_RESULT];

export const WRONG_BOOK_STATUS = {
  ACTIVE: 'active',
  RESOLVED: 'resolved',
} as const;

export const QUESTION_STATUS = {
  DRAFT: 'draft',
  PUBLISHED: 'published',
} as const;

export const SUBJECTS = ['计算机基础', '数学', '语文', '英语', '物理', '化学', '生物', '历史', '政治', '地理'];

export const DEFAULT_KNOWLEDGE_POINTS = ['数据结构', '代数', '几何', '文言文', '阅读理解', '词汇', '力学', '电学', '化学平衡', '遗传学'];
