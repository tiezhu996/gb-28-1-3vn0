package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
	"github.com/onlineexam/onlineexam/internal/util"
	"github.com/onlineexam/onlineexam/pkg/sliceutil"
)

// QuestionService 题库服务：题目 CRUD、批量导入、自动组卷取题（复用 RandomPick）。
type QuestionService struct {
	repo   repository.QuestionRepository
	logger *slog.Logger
}

// NewQuestionService 构造题库服务。
func NewQuestionService(repo repository.QuestionRepository, logger *slog.Logger) *QuestionService {
	return &QuestionService{repo: repo, logger: logger}
}

// buildQuestionFromRow 将输入转换为题目模型（Create 与 Import 复用此方法）。
func buildQuestionFromRow(qt, subject string, kps []string, difficulty, content, answer, analysis string, options []model.QuestionOption, score float64, creatorID primitive.ObjectID, status string) *model.Question {
	now := time.Now()
	return &model.Question{
		ID:              primitive.NewObjectID(),
		Type:            qt,
		Subject:         subject,
		KnowledgePoints: sliceutil.Dedupe(kps),
		Difficulty:      difficulty,
		Content:         content,
		Options:         options,
		Answer:          answer,
		Analysis:        analysis,
		Score:           score,
		Status:          status,
		CreatorID:       creatorID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// validateQuestion 校验题型、难度与答案格式（Create/Import 复用）。
func validateQuestion(qt, difficulty, answer string, options []model.QuestionOption) error {
	if !constants.IsValidQuestionType(qt) {
		return util.NewAppError(constants.CodeQuestionTypeErr, fmt.Sprintf(constants.MsgQuestionTypeInvalid, qt))
	}
	if !constants.IsValidDifficulty(difficulty) {
		return util.NewAppError(constants.CodeBadRequest, fmt.Sprintf(constants.MsgValidationFailed, "difficulty"))
	}
	if qt == constants.QuestionTypeSingle || qt == constants.QuestionTypeMultiple {
		if len(options) == 0 {
			return util.NewAppError(constants.CodeBadRequest, fmt.Sprintf(constants.MsgValidationFailed, "options"))
		}
	}
	if strings.TrimSpace(answer) == "" {
		return util.NewAppError(constants.CodeBadRequest, fmt.Sprintf(constants.MsgValidationFailed, "answer"))
	}
	return nil
}

// Create 创建题目。
func (s *QuestionService) Create(ctx context.Context, req *dto.CreateQuestionRequest, creatorID primitive.ObjectID) (*model.Question, error) {
	opts := make([]model.QuestionOption, 0, len(req.Options))
	for _, o := range req.Options {
		opts = append(opts, model.QuestionOption{Key: o.Key, Text: o.Text})
	}
	if err := validateQuestion(req.Type, req.Difficulty, req.Answer, opts); err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = constants.QuestionStatusDraft
	}
	q := buildQuestionFromRow(req.Type, req.Subject, req.KnowledgePoints, req.Difficulty, req.Content, req.Answer, req.Analysis, opts, req.Score, creatorID, status)
	if err := s.repo.Create(ctx, q); err != nil {
		return nil, fmt.Errorf("question service create: %w", err)
	}
	s.logger.Info(constants.LogQuestionCreated, "type", q.Type, "difficulty", q.Difficulty, "subject", q.Subject, "creator", creatorID.Hex())
	return q, nil
}

// Update 更新题目。
func (s *QuestionService) Update(ctx context.Context, id primitive.ObjectID, req *dto.UpdateQuestionRequest, operator string) (*model.Question, error) {
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeQuestionNotFound, fmt.Sprintf(constants.MsgQuestionNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("question service update find: %w", err)
	}
	if req.Type != "" {
		q.Type = req.Type
	}
	if req.Subject != "" {
		q.Subject = req.Subject
	}
	if req.KnowledgePoints != nil {
		q.KnowledgePoints = req.KnowledgePoints
	}
	if req.Difficulty != "" {
		q.Difficulty = req.Difficulty
	}
	if req.Content != "" {
		q.Content = req.Content
	}
	if req.Options != nil {
		opts := make([]model.QuestionOption, 0, len(req.Options))
		for _, o := range req.Options {
			opts = append(opts, model.QuestionOption{Key: o.Key, Text: o.Text})
		}
		q.Options = opts
	}
	if req.Answer != "" {
		q.Answer = req.Answer
	}
	if req.Analysis != "" {
		q.Analysis = req.Analysis
	}
	if req.Score > 0 {
		q.Score = req.Score
	}
	if req.Status != "" {
		q.Status = req.Status
	}
	if err := validateQuestion(q.Type, q.Difficulty, q.Answer, q.Options); err != nil {
		return nil, err
	}
	q.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, q); err != nil {
		return nil, fmt.Errorf("question service update: %w", err)
	}
	s.logger.Info(constants.LogQuestionUpdated, "type", q.Type, "difficulty", q.Difficulty, "subject", q.Subject, "operator", operator)
	return q, nil
}

// Delete 删除题目。
func (s *QuestionService) Delete(ctx context.Context, id primitive.ObjectID, operator string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeQuestionNotFound, fmt.Sprintf(constants.MsgQuestionNotFound, id.Hex()))
		}
		return fmt.Errorf("question service delete: %w", err)
	}
	s.logger.Info(constants.LogQuestionDeleted, "question_id", id.Hex(), "operator", operator)
	return nil
}

// GetByID 查询单个题目。
func (s *QuestionService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Question, error) {
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeQuestionNotFound, fmt.Sprintf(constants.MsgQuestionNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("question service get: %w", err)
	}
	return q, nil
}

// List 分页查询题目。
func (s *QuestionService) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.Question, int64, error) {
	list, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("question service list: %w", err)
	}
	return list, total, nil
}

// Import 批量导入题目（与 Create 共用 buildQuestionFromRow）。
func (s *QuestionService) Import(ctx context.Context, rows []util.ExcelQuestionRow, creatorID primitive.ObjectID) (int, error) {
	questions := make([]*model.Question, 0, len(rows))
	for _, row := range rows {
		var opts []model.QuestionOption
		keys := []string{"A", "B", "C", "D"}
		for i, text := range row.Options {
			if strings.TrimSpace(text) == "" {
				continue
			}
			opts = append(opts, model.QuestionOption{Key: keys[i], Text: text})
		}
		kps := strings.Split(row.KnowledgePoint, ",")
		q := buildQuestionFromRow(row.Type, row.Subject, kps, row.Difficulty, row.Content, row.Answer, row.Analysis, opts, row.Score, creatorID, constants.QuestionStatusPublished)
		if err := validateQuestion(q.Type, q.Difficulty, q.Answer, q.Options); err != nil {
			return 0, util.WrapAppError(constants.CodeQuestionImportErr, fmt.Sprintf("题库模块：第 %d 行导入校验失败（字段 type/difficulty/answer 非法）", len(questions)+1), err)
		}
		questions = append(questions, q)
	}
	if err := s.repo.CreateMany(ctx, questions); err != nil {
		return 0, fmt.Errorf("question service import: %w", err)
	}
	subject := ""
	if len(questions) > 0 {
		subject = questions[0].Subject
	}
	s.logger.Info(constants.LogQuestionImported, "count", len(questions), "subject", subject, "operator", creatorID.Hex())
	return len(questions), nil
}

// PickByDifficulty 按学科+知识点+难度随机抽取题目（自动组卷复用 RandomPick）。
func (s *QuestionService) PickByDifficulty(ctx context.Context, subject string, knowledgePoints []string, difficulty string, limit int) ([]*model.Question, error) {
	filter := bson.M{"difficulty": difficulty, "status": constants.QuestionStatusPublished}
	if subject != "" {
		filter["subject"] = subject
	}
	if len(knowledgePoints) > 0 {
		filter["knowledge_points"] = bson.M{"$in": knowledgePoints}
	}
	return s.repo.RandomPick(ctx, filter, limit)
}
