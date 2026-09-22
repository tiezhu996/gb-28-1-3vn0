package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
	"github.com/onlineexam/onlineexam/internal/util"
)

// ExamService 试卷服务：手动组卷、自动组卷、状态流转、CRUD。
type ExamService struct {
	repo     repository.ExamRepository
	question *QuestionService // 复用题库服务（自动组卷取题）
	logger   *slog.Logger
}

// NewExamService 构造试卷服务。
func NewExamService(repo repository.ExamRepository, question *QuestionService, logger *slog.Logger) *ExamService {
	return &ExamService{repo: repo, question: question, logger: logger}
}

// applyStatusTransition 状态机流转公共方法（Create/Update/Publish/Close 复用）。
// 状态机规则见 constants.ExamStatusTransitions。
func (s *ExamService) applyStatusTransition(ctx context.Context, exam *model.Exam, to string, operator string) error {
	if !constants.IsValidExamStatus(to) {
		return util.NewAppError(constants.CodeExamStatusErr, fmt.Sprintf(constants.MsgExamStatusInvalid, to, exam.Status, to))
	}
	if exam.Status == to {
		return nil
	}
	if !constants.CanTransition(exam.Status, to, constants.ExamStatusTransitions) {
		return util.NewAppError(constants.CodeExamStatusErr, fmt.Sprintf(constants.MsgExamStatusInvalid, to, exam.Status, to))
	}
	from := exam.Status
	exam.Status = to
	exam.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, exam); err != nil {
		return fmt.Errorf("exam service apply status transition: %w", err)
	}
	s.logger.Info(constants.LogExamStatusChanged, "from", from, "to", to, "exam_id", exam.ID.Hex(), "operator", operator)
	return nil
}

// resolveQuestions 根据输入题目 ID 列表加载题目并组装（Create/Update 复用）。
func (s *ExamService) resolveQuestions(ctx context.Context, items []dto.ExamQuestionInput) ([]model.ExamQuestion, float64, error) {
	if len(items) == 0 {
		return nil, 0, nil
	}
	ids := make([]primitive.ObjectID, 0, len(items))
	for _, it := range items {
		oid, err := primitive.ObjectIDFromHex(it.QuestionID)
		if err != nil {
			return nil, 0, util.NewAppError(constants.CodeBadRequest, fmt.Sprintf("试卷模块：题目 id 非法 %s", it.QuestionID))
		}
		ids = append(ids, oid)
	}
	qs, err := s.question.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, 0, fmt.Errorf("exam service resolve questions: %w", err)
	}
	byID := make(map[string]*model.Question, len(qs))
	for _, q := range qs {
		byID[q.ID.Hex()] = q
	}
	out := make([]model.ExamQuestion, 0, len(items))
	var total float64
	for i, it := range items {
		q, ok := byID[it.QuestionID]
		if !ok {
			return nil, 0, util.NewAppError(constants.CodeQuestionNotFound, fmt.Sprintf(constants.MsgQuestionNotFound, it.QuestionID))
		}
		score := it.Score
		if score <= 0 {
			score = q.Score
		}
		out = append(out, model.ExamQuestion{QuestionID: q.ID, Score: score, Order: i + 1})
		total += score
	}
	return out, total, nil
}

// Create 手动创建试卷。
func (s *ExamService) Create(ctx context.Context, req *dto.CreateExamRequest, creatorID primitive.ObjectID) (*model.Exam, error) {
	questions, total, err := s.resolveQuestions(ctx, req.Questions)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	exam := &model.Exam{
		ID:              primitive.NewObjectID(),
		Title:           req.Title,
		Subject:         req.Subject,
		Description:     req.Description,
		TotalScore:      total,
		PassScore:       req.PassScore,
		DurationMin:     req.DurationMin,
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		Status:          constants.ExamStatusDraft,
		ShuffleQuestion: req.ShuffleQuestion,
		ShuffleOption:   req.ShuffleOption,
		Questions:       questions,
		CreatedBy:       creatorID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, exam); err != nil {
		return nil, fmt.Errorf("exam service create: %w", err)
	}
	s.logger.Info(constants.LogExamCreated, "title", exam.Title, "total_score", exam.TotalScore, "subject", exam.Subject, "creator", creatorID.Hex())
	return exam, nil
}

// AutoGenerate 自动组卷：按学科、知识点覆盖与难度分布随机取题。
func (s *ExamService) AutoGenerate(ctx context.Context, req *dto.AutoGenerateRequest, creatorID primitive.ObjectID) (*model.Exam, error) {
	var picked []*model.Question
	// 难度分布 map 遍历顺序不确定，先按 easy/medium/hard 固定顺序取题
	order := []string{constants.DifficultyEasy, constants.DifficultyMedium, constants.DifficultyHard}
	for _, diff := range order {
		n := req.DifficultyDist[diff]
		if n <= 0 {
			continue
		}
		qs, err := s.question.PickByDifficulty(ctx, req.Subject, req.KnowledgePoints, diff, n)
		if err != nil {
			return nil, fmt.Errorf("exam service auto generate pick: %w", err)
		}
		picked = append(picked, qs...)
	}
	if len(picked) == 0 {
		return nil, util.NewAppError(constants.CodeExamNoQuestions, constants.MsgExamNoQuestions)
	}
	// 随机题序
	util.Shuffle(picked)
	sort.SliceStable(picked, func(i, j int) bool { return picked[i].Difficulty < picked[j].Difficulty })
	questions := make([]model.ExamQuestion, 0, len(picked))
	var total float64
	for i, q := range picked {
		score := req.ScorePerQuestion
		if score <= 0 {
			score = q.Score
		}
		questions = append(questions, model.ExamQuestion{QuestionID: q.ID, Score: score, Order: i + 1})
		total += score
	}
	now := time.Now()
	exam := &model.Exam{
		ID:              primitive.NewObjectID(),
		Title:           req.Title,
		Subject:         req.Subject,
		Description:     req.Description,
		TotalScore:      total,
		PassScore:       req.PassScore,
		DurationMin:     req.DurationMin,
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		Status:          constants.ExamStatusDraft,
		ShuffleQuestion: req.ShuffleQuestion,
		ShuffleOption:   req.ShuffleOption,
		Questions:       questions,
		CreatedBy:       creatorID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, exam); err != nil {
		return nil, fmt.Errorf("exam service auto generate: %w", err)
	}
	s.logger.Info(constants.LogExamAutoGenerated, "title", exam.Title, "question_count", len(questions), "difficulty", req.Subject, "creator", creatorID.Hex())
	return exam, nil
}

// Update 更新草稿试卷（或按请求做状态流转）。
func (s *ExamService) Update(ctx context.Context, id primitive.ObjectID, req *dto.UpdateExamRequest, operator string) (*model.Exam, error) {
	exam, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeExamNotFound, fmt.Sprintf(constants.MsgExamNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("exam service update find: %w", err)
	}
	if exam.Status != constants.ExamStatusDraft && req.Status == "" && req.Questions != nil {
		return nil, util.NewAppError(constants.CodeExamStatusErr, fmt.Sprintf(constants.MsgExamStatusInvalid, exam.Status, exam.Status, "edit"))
	}
	if req.Title != "" {
		exam.Title = req.Title
	}
	if req.Subject != "" {
		exam.Subject = req.Subject
	}
	if req.Description != "" {
		exam.Description = req.Description
	}
	if req.TotalScore > 0 {
		exam.TotalScore = req.TotalScore
	}
	if req.PassScore > 0 {
		exam.PassScore = req.PassScore
	}
	if req.DurationMin > 0 {
		exam.DurationMin = req.DurationMin
	}
	if req.StartAt != nil {
		exam.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		exam.EndAt = *req.EndAt
	}
	if req.ShuffleQuestion != nil {
		exam.ShuffleQuestion = *req.ShuffleQuestion
	}
	if req.ShuffleOption != nil {
		exam.ShuffleOption = *req.ShuffleOption
	}
	if req.Questions != nil {
		questions, total, err := s.resolveQuestions(ctx, req.Questions)
		if err != nil {
			return nil, err
		}
		exam.Questions = questions
		exam.TotalScore = total
	}
	exam.UpdatedAt = time.Now()
	if req.Status != "" {
		if err := s.applyStatusTransition(ctx, exam, req.Status, operator); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(ctx, exam); err != nil {
			return nil, fmt.Errorf("exam service update: %w", err)
		}
	}
	s.logger.Info(constants.LogExamUpdated, "title", exam.Title, "exam_id", exam.ID.Hex(), "operator", operator)
	return exam, nil
}

// Publish 发布试卷（draft → published）。
func (s *ExamService) Publish(ctx context.Context, id primitive.ObjectID, operator string) (*model.Exam, error) {
	exam, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeExamNotFound, fmt.Sprintf(constants.MsgExamNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("exam service publish find: %w", err)
	}
	if len(exam.Questions) == 0 {
		return nil, util.NewAppError(constants.CodeExamNoQuestions, constants.MsgExamNoQuestions)
	}
	if err := s.applyStatusTransition(ctx, exam, constants.ExamStatusPublished, operator); err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogExamPublishSuccess, "exam_id", exam.ID.Hex(), "operator", operator)
	return exam, nil
}

// Close 关闭试卷（任意状态 → closed）。
func (s *ExamService) Close(ctx context.Context, id primitive.ObjectID, operator string) (*model.Exam, error) {
	exam, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeExamNotFound, fmt.Sprintf(constants.MsgExamNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("exam service close find: %w", err)
	}
	if err := s.applyStatusTransition(ctx, exam, constants.ExamStatusClosed, operator); err != nil {
		return nil, err
	}
	return exam, nil
}

// List 分页查询试卷。
func (s *ExamService) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.Exam, int64, error) {
	list, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("exam service list: %w", err)
	}
	return list, total, nil
}

// GetByID 查询单个试卷。
func (s *ExamService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Exam, error) {
	exam, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeExamNotFound, fmt.Sprintf(constants.MsgExamNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("exam service get: %w", err)
	}
	return exam, nil
}

// Delete 删除试卷。
func (s *ExamService) Delete(ctx context.Context, id primitive.ObjectID, operator string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeExamNotFound, fmt.Sprintf(constants.MsgExamNotFound, id.Hex()))
		}
		return fmt.Errorf("exam service delete: %w", err)
	}
	s.logger.Info(constants.LogExamDeleted, "exam_id", id.Hex(), "operator", operator)
	return nil
}
