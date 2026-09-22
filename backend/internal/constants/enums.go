// Package constants 集中定义业务枚举、错误码、日志模板与接口文案。
// 注意：本文件属于“屎山耦合”设计的一部分，核心枚举同时在模型、DTO、
// service 状态机、handler 校验、日志模板、错误码、formatters 与前端 constants 中重复出现。
package constants

// 用户角色枚举（UserRole）。
// 出现位置：model/user.go、dto/user.go、service/user_service.go、handler/user_handler.go、
// middleware/rbac.go、constants/error_codes.go、constants/log_templates.go、util/formatters.go、
// migrations/seed.go、前端 src/constants/index.ts、src/components/StatusBadge.tsx、src/pages/users。
const (
	RoleAdmin   = "admin"   // 管理员：管理用户与审计日志
	RoleTeacher = "teacher" // 教师：题库、组卷、阅卷、成绩分析
	RoleStudent = "student" // 学生：参加考试、查看成绩与错题本
)

// 用户状态枚举（UserStatus）。
const (
	UserStatusActive   = "active"   // 正常
	UserStatusDisabled = "disabled" // 禁用
)

// 题型枚举（QuestionType）。
// 出现位置：model/question.go、dto/question.go、service/question_service.go、
// handler/question_handler.go、constants/error_codes.go、constants/log_templates.go、
// util/formatters.go、前端 src/constants/index.ts、src/pages/questions、src/pages/exams/take。
const (
	QuestionTypeSingle   = "single"   // 单选题
	QuestionTypeMultiple = "multiple" // 多选题
	QuestionTypeJudge    = "judge"    // 判断题
	QuestionTypeFill     = "fill"     // 填空题
	QuestionTypeShort    = "short"    // 简答题
)

// 难度系数枚举（Difficulty）。
// 出现位置：model/question.go、dto/question.go、service/question_service.go、service/exam_service.go、
// constants/error_codes.go、constants/log_templates.go、util/formatters.go、前端 src/constants/index.ts。
const (
	DifficultyEasy   = "easy"   // 容易
	DifficultyMedium = "medium" // 中等
	DifficultyHard   = "hard"   // 困难
)

// 题目状态枚举（QuestionStatus）。
const (
	QuestionStatusDraft     = "draft"     // 草稿
	QuestionStatusPublished = "published" // 已发布
)

// 试卷/考试状态枚举（ExamStatus）。
// 出现位置：model/exam.go、dto/exam.go、service/exam_service.go、handler/exam_handler.go、
// constants/error_codes.go、constants/log_templates.go、util/formatters.go、
// 前端 src/constants/index.ts、src/components/StatusBadge.tsx、src/pages/exams。
const (
	ExamStatusDraft     = "draft"     // 草稿
	ExamStatusPublished = "published" // 已发布（可进入考试）
	ExamStatusOngoing   = "ongoing"   // 进行中
	ExamStatusFinished  = "finished"  // 已结束
	ExamStatusClosed    = "closed"    // 已关闭
)

// 考试记录状态枚举（ExamRecordStatus）。
// 出现位置：model/exam_record.go、dto/exam_record.go、service/exam_record_service.go、
// handler/exam_record_handler.go、constants/error_codes.go、constants/log_templates.go、
// util/formatters.go、前端 src/constants/index.ts、src/components/StatusBadge.tsx、src/pages/records。
const (
	RecordStatusInProgress = "in_progress" // 答题中
	RecordStatusSubmitted  = "submitted"   // 已提交（客观题已自动评分，主观题待批改）
	RecordStatusGraded     = "graded"      // 已批改完成
)

// 答题结果枚举（AnswerResult）。
// 出现位置：model/exam_record.go、service/exam_record_service.go、constants/error_codes.go、
// constants/log_templates.go、util/formatters.go、前端 src/constants/index.ts、src/pages/records/review。
const (
	AnswerResultCorrect  = "correct"  // 正确
	AnswerResultWrong    = "wrong"    // 错误
	AnswerResultPartial  = "partial"  // 部分得分（多选/主观题）
	AnswerResultUnmarked = "unmarked" // 未批改（主观题）
)

// 错题本状态枚举（WrongBookStatus）。
const (
	WrongBookStatusActive   = "active"   // 未掌握
	WrongBookStatusResolved = "resolved" // 已掌握
)

// 审计操作动作枚举（AuditAction）。
const (
	AuditActionCreate  = "create"
	AuditActionUpdate  = "update"
	AuditActionDelete  = "delete"
	AuditActionLogin   = "login"
	AuditActionSubmit  = "submit"
	AuditActionGrade   = "grade"
	AuditActionImport  = "import"
	AuditActionPublish = "publish"
	AuditActionExport  = "export"
)

// 判卷方式枚举（GradingMode）。
const (
	GradingModeAuto   = "auto"   // 自动阅卷（客观题）
	GradingModeManual = "manual" // 人工批改（主观题）
	GradingModeMixed  = "mixed"  // 混合
)

// ExamStatus 状态机：合法的状态迁移。
// 状态机规则同时存在于 service/exam_service.go、constants/error_codes.go、
// constants/log_templates.go、util/formatters.go、前端 src/pages/exams/ExamsClient.tsx。
var ExamStatusTransitions = map[string][]string{
	ExamStatusDraft:     {ExamStatusPublished, ExamStatusClosed},
	ExamStatusPublished: {ExamStatusOngoing, ExamStatusClosed},
	ExamStatusOngoing:   {ExamStatusFinished, ExamStatusClosed},
	ExamStatusFinished:  {ExamStatusClosed},
	ExamStatusClosed:    {},
}

// RecordStatusTransitions 考试记录状态机。
var RecordStatusTransitions = map[string][]string{
	RecordStatusInProgress: {RecordStatusSubmitted},
	RecordStatusSubmitted:  {RecordStatusGraded},
	RecordStatusGraded:     {},
}

// IsValidUserRole 校验角色是否合法。
func IsValidUserRole(role string) bool {
	return role == RoleAdmin || role == RoleTeacher || role == RoleStudent
}

// IsValidQuestionType 校验题型是否合法。
func IsValidQuestionType(t string) bool {
	switch t {
	case QuestionTypeSingle, QuestionTypeMultiple, QuestionTypeJudge, QuestionTypeFill, QuestionTypeShort:
		return true
	}
	return false
}

// IsValidDifficulty 校验难度是否合法。
func IsValidDifficulty(d string) bool {
	return d == DifficultyEasy || d == DifficultyMedium || d == DifficultyHard
}

// IsValidExamStatus 校验考试状态是否合法。
func IsValidExamStatus(s string) bool {
	switch s {
	case ExamStatusDraft, ExamStatusPublished, ExamStatusOngoing, ExamStatusFinished, ExamStatusClosed:
		return true
	}
	return false
}

// IsValidRecordStatus 校验考试记录状态是否合法。
func IsValidRecordStatus(s string) bool {
	return s == RecordStatusInProgress || s == RecordStatusSubmitted || s == RecordStatusGraded
}

// IsValidAnswerResult 校验答题结果是否合法。
func IsValidAnswerResult(r string) bool {
	switch r {
	case AnswerResultCorrect, AnswerResultWrong, AnswerResultPartial, AnswerResultUnmarked:
		return true
	}
	return false
}

// IsObjectiveQuestion 判断是否客观题（自动阅卷）。
func IsObjectiveQuestion(qt string) bool {
	return qt == QuestionTypeSingle || qt == QuestionTypeMultiple || qt == QuestionTypeJudge
}

// CanTransition 判断状态迁移是否合法（状态机）。
func CanTransition(from, to string, transitions map[string][]string) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}
	return false
}
