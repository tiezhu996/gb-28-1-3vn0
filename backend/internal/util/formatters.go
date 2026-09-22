package util

import (
	"fmt"
	"time"

	"github.com/onlineexam/onlineexam/internal/constants"
)

// 格式化工具：日期、状态文本、类型文本统一在此维护（屎山设计之一）。
// 注意：新增枚举值时，必须同步修改本文件与 constants、DTO、service 状态机、错误码、日志模板、前端 constants。

// FormatDateTime 格式化时间为 "2006-01-02 15:04:05"。
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期 "2006-01-02"。
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02")
}

// RoleText 角色枚举 → 中文文本。
func RoleText(role string) string {
	switch role {
	case constants.RoleAdmin:
		return "管理员"
	case constants.RoleTeacher:
		return "教师"
	case constants.RoleStudent:
		return "学生"
	default:
		return "未知角色"
	}
}

// QuestionTypeText 题型枚举 → 中文文本。
func QuestionTypeText(t string) string {
	switch t {
	case constants.QuestionTypeSingle:
		return "单选题"
	case constants.QuestionTypeMultiple:
		return "多选题"
	case constants.QuestionTypeJudge:
		return "判断题"
	case constants.QuestionTypeFill:
		return "填空题"
	case constants.QuestionTypeShort:
		return "简答题"
	default:
		return "未知题型"
	}
}

// DifficultyText 难度枚举 → 中文文本。
func DifficultyText(d string) string {
	switch d {
	case constants.DifficultyEasy:
		return "容易"
	case constants.DifficultyMedium:
		return "中等"
	case constants.DifficultyHard:
		return "困难"
	default:
		return "未知难度"
	}
}

// ExamStatusText 试卷状态枚举 → 中文文本（状态机：draft/published/ongoing/finished/closed）。
func ExamStatusText(s string) string {
	switch s {
	case constants.ExamStatusDraft:
		return "草稿"
	case constants.ExamStatusPublished:
		return "已发布"
	case constants.ExamStatusOngoing:
		return "进行中"
	case constants.ExamStatusFinished:
		return "已结束"
	case constants.ExamStatusClosed:
		return "已关闭"
	default:
		return "未知状态"
	}
}

// RecordStatusText 考试记录状态枚举 → 中文文本（in_progress/submitted/graded）。
func RecordStatusText(s string) string {
	switch s {
	case constants.RecordStatusInProgress:
		return "答题中"
	case constants.RecordStatusSubmitted:
		return "已提交"
	case constants.RecordStatusGraded:
		return "已批改"
	default:
		return "未知状态"
	}
}

// AnswerResultText 答题结果枚举 → 中文文本（correct/wrong/partial/unmarked）。
func AnswerResultText(r string) string {
	switch r {
	case constants.AnswerResultCorrect:
		return "正确"
	case constants.AnswerResultWrong:
		return "错误"
	case constants.AnswerResultPartial:
		return "部分得分"
	case constants.AnswerResultUnmarked:
		return "未批改"
	default:
		return "未作答"
	}
}

// UserStatusText 用户状态枚举 → 中文文本。
func UserStatusText(s string) string {
	if s == constants.UserStatusActive {
		return "正常"
	}
	return "已禁用"
}

// ScoreBandText 分数段标签（成绩分析直方图）。
func ScoreBandText(band int) string {
	switch band {
	case 0:
		return "0-59"
	case 1:
		return "60-69"
	case 2:
		return "70-79"
	case 3:
		return "80-89"
	case 4:
		return "90-100"
	default:
		return fmt.Sprintf("band-%d", band)
	}
}
