package constants

// 统一错误码定义。
// 约定：code=0 表示成功；业务错误码 1001-1999 为通用，2001-2999 用户，3001-3999 题库，
// 4001-4999 试卷/考试，5001-5999 考试记录，6001-6999 错题本，7001-7999 审计。
// 注意：每个 service/handler 在抛出错误时必须手动拼接包含实体名、字段名、角色名的 message。
const (
	CodeOK               = 0    // 成功
	CodeBadRequest       = 1001 // 参数错误
	CodeUnauthorized     = 1002 // 未认证/登录过期
	CodeForbidden        = 1003 // 无权限
	CodeNotFound         = 1004 // 资源不存在
	CodeConflict         = 1005 // 状态冲突
	CodeInternalError    = 1006 // 内部错误
	CodeValidationFailed = 1007 // 参数校验失败
	CodeRateLimited      = 1008 // 请求过于频繁
	CodeDuplicateKey     = 1009 // 唯一键冲突

	// 用户模块
	CodeUserEmailExists = 2001 // 用户邮箱已存在
	CodeUserNotFound    = 2002 // 用户不存在
	CodeUserBadPassword = 2003 // 用户密码错误
	CodeUserDisabled    = 2004 // 用户已被禁用
	CodeUserInvalidRole = 2005 // 用户角色非法

	// 题库模块
	CodeQuestionNotFound  = 3001 // 题目不存在
	CodeQuestionTypeErr   = 3002 // 题目类型非法
	CodeQuestionImportErr = 3003 // 题目导入失败

	// 试卷模块
	CodeExamNotFound    = 4001 // 试卷不存在
	CodeExamStatusErr   = 4002 // 试卷状态非法/状态流转非法
	CodeExamNotInWindow = 4003 // 不在考试时间窗口内
	CodeExamNoQuestions = 4004 // 试卷未配置题目

	// 考试记录模块
	CodeRecordNotFound    = 5001 // 考试记录不存在
	CodeRecordStatusErr   = 5002 // 考试记录状态非法/不可提交
	CodeRecordExpired     = 5003 // 考试记录已超时
	CodeRecordAlreadyDone = 5004 // 考试记录已提交

	// 错题本模块
	CodeWrongBookNotFound = 6001 // 错题本条目不存在
	CodeWrongBookExists   = 6002 // 错题已存在错题本

	// 审计模块
	CodeAuditNotFound = 7001 // 审计日志不存在
)

// ErrorCodeText 返回错误码对应的默认文案（供 messages 与 handler 包装使用）。
func ErrorCodeText(code int) string {
	switch code {
	case CodeOK:
		return "ok"
	case CodeBadRequest:
		return "请求参数错误"
	case CodeUnauthorized:
		return "未登录或登录已过期"
	case CodeForbidden:
		return "无权限执行该操作"
	case CodeNotFound:
		return "资源不存在"
	case CodeConflict:
		return "资源状态冲突"
	case CodeInternalError:
		return "服务器内部错误"
	case CodeValidationFailed:
		return "参数校验失败"
	case CodeRateLimited:
		return "请求过于频繁，请稍后再试"
	case CodeDuplicateKey:
		return "数据已存在"
	case CodeUserEmailExists:
		return "用户邮箱已存在"
	case CodeUserNotFound:
		return "用户不存在"
	case CodeUserBadPassword:
		return "用户密码错误"
	case CodeUserDisabled:
		return "用户已被禁用"
	case CodeUserInvalidRole:
		return "用户角色非法"
	case CodeQuestionNotFound:
		return "题目不存在"
	case CodeQuestionTypeErr:
		return "题目类型非法"
	case CodeQuestionImportErr:
		return "题目导入失败"
	case CodeExamNotFound:
		return "试卷不存在"
	case CodeExamStatusErr:
		return "试卷状态非法或状态流转不允许"
	case CodeExamNotInWindow:
		return "不在考试时间窗口内"
	case CodeExamNoQuestions:
		return "试卷未配置题目"
	case CodeRecordNotFound:
		return "考试记录不存在"
	case CodeRecordStatusErr:
		return "考试记录状态非法或不可执行该操作"
	case CodeRecordExpired:
		return "考试记录已超时"
	case CodeRecordAlreadyDone:
		return "考试记录已提交"
	case CodeWrongBookNotFound:
		return "错题本条目不存在"
	case CodeWrongBookExists:
		return "错题已存在于错题本"
	case CodeAuditNotFound:
		return "审计日志不存在"
	default:
		return "未知错误"
	}
}
