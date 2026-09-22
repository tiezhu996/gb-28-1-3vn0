package constants

// 统一文案集中定义：接口返回文案、日志文案、错误提示文案混合维护（屎山设计之一）。
// 注意：修改任何文案时，需要同步检查 handler 返回、service 日志与前端提示。
const (
	MsgOK                    = "ok"
	MsgUnauthorized          = "请先登录后再操作"
	MsgForbidden             = "当前角色无权执行该操作"
	MsgLoginSuccess          = "登录成功"
	MsgRegisterSuccess       = "注册成功"
	MsgLogoutSuccess         = "退出登录成功"
	MsgQuestionImportSuccess = "题目批量导入成功"
	MsgExamAutoGenerate      = "自动组卷完成"
	MsgExamPublishSuccess    = "试卷发布成功"
	MsgRecordSubmitSuccess   = "答卷提交成功"
	MsgRecordGradedSuccess   = "主观题批改完成"
	MsgWrongBookAdded        = "已加入错题本"
	MsgWrongBookResolved     = "已标记为已掌握"

	// 错误提示文案（与 error_codes.go 对应，但由 service/handler 手动拼接实体名、字段名、角色名）
	MsgValidationFailed    = "参数校验失败：字段 %s 不符合要求"
	MsgUserEmailExists     = "用户模块：邮箱字段 %s 已被注册"
	MsgUserNotFound        = "用户模块：id=%s 的用户不存在"
	MsgUserBadPassword     = "用户模块：密码字段不正确"
	MsgUserDisabled        = "用户模块：角色 %s 的用户已被禁用"
	MsgQuestionNotFound    = "题库模块：id=%s 的题目不存在"
	MsgQuestionTypeInvalid = "题库模块：题型字段 %s 非法"
	MsgExamNotFound        = "试卷模块：id=%s 的试卷不存在"
	MsgExamStatusInvalid   = "试卷模块：状态字段 %s 非法，无法从 %s 流转到 %s"
	MsgExamNotInWindow     = "试卷模块：考试时间窗口校验失败"
	MsgExamNoQuestions     = "试卷模块：题目列表为空，无法开始考试"
	MsgRecordNotFound      = "考试记录模块：id=%s 的记录不存在"
	MsgRecordStatusInvalid = "考试记录模块：状态字段 %s 非法，无法执行该操作"
	MsgRecordExpired       = "考试记录模块：考试时长已超时"
	MsgRecordVersionStale  = "考试记录模块：草稿版本 version=%d 已过期（当前版本 %d），旧保存请求已忽略"
	MsgRecordNotOwner      = "考试记录模块：id=%s 的答卷不属于当前学生，无权操作"
	MsgWrongBookExists     = "错题本模块：question_id=%s 已在错题本中"
)
