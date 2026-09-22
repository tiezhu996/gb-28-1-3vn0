package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
	"github.com/onlineexam/onlineexam/internal/util"
)

// WrongBookService 错题本服务：加入/复习/移除。
type WrongBookService struct {
	repo     repository.WrongBookRepository
	question *QuestionService // 复用题库服务（加载题目快照）
	record   *ExamRecordService // 复用考试记录服务（获取学生作答）
	logger   *slog.Logger
}

// NewWrongBookService 构造错题本服务。
func NewWrongBookService(repo repository.WrongBookRepository, question *QuestionService, record *ExamRecordService, logger *slog.Logger) *WrongBookService {
	return &WrongBookService{repo: repo, question: question, record: record, logger: logger}
}

// Add 将错题加入错题本（幂等：已存在则直接返回）。
func (s *WrongBookService) Add(ctx context.Context, studentID, questionID, examID, examRecordID primitive.ObjectID, note string) (*model.WrongBook, error) {
	if existing, err := s.repo.FindByStudentAndQuestion(ctx, studentID, questionID); err == nil {
		return existing, nil
	}
	q, err := s.question.GetByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	myAnswer := ""
	if !examRecordID.IsZero() {
		rec, err := s.record.GetByID(ctx, examRecordID)
		if err == nil {
			for i := range rec.Questions {
				if rec.Questions[i].QuestionID == questionID {
					myAnswer = rec.Questions[i].UserAnswer
					break
				}
			}
		}
	}
	now := time.Now()
	entry := &model.WrongBook{
		ID:              primitive.NewObjectID(),
		StudentID:       studentID,
		QuestionID:      questionID,
		ExamID:          examID,
		ExamRecordID:    examRecordID,
		Subject:         q.Subject,
		KnowledgePoints: q.KnowledgePoints,
		QuestionContent: q.Content,
		MyAnswer:        myAnswer,
		CorrectAnswer:   q.Answer,
		Analysis:        q.Analysis,
		Note:            note,
		Status:          constants.WrongBookStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("wrong book service add: %w", err)
	}
	s.logger.Info(constants.LogWrongBookAdded, "student", studentID.Hex(), "question_id", questionID.Hex())
	return entry, nil
}

// Update 更新错题本（标记已掌握/修改备注）。
func (s *WrongBookService) Update(ctx context.Context, id, studentID primitive.ObjectID, status, note string) (*model.WrongBook, error) {
	entry, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeWrongBookNotFound, fmt.Sprintf("错题本模块：id=%s 的条目不存在", id.Hex()))
		}
		return nil, fmt.Errorf("wrong book service update find: %w", err)
	}
	if entry.StudentID != studentID {
		return nil, util.NewAppError(constants.CodeForbidden, constants.MsgForbidden)
	}
	if status != "" {
		if status != constants.WrongBookStatusActive && status != constants.WrongBookStatusResolved {
			return nil, util.NewAppError(constants.CodeBadRequest, fmt.Sprintf(constants.MsgValidationFailed, "status"))
		}
		entry.Status = status
	}
	if note != "" {
		entry.Note = note
	}
	entry.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("wrong book service update: %w", err)
	}
	if entry.Status == constants.WrongBookStatusResolved {
		s.logger.Info(constants.LogWrongBookResolved, "student", studentID.Hex(), "question_id", entry.QuestionID.Hex())
	}
	return entry, nil
}

// Delete 移除错题。
func (s *WrongBookService) Delete(ctx context.Context, id, studentID primitive.ObjectID) error {
	entry, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeWrongBookNotFound, fmt.Sprintf("错题本模块：id=%s 的条目不存在", id.Hex()))
		}
		return fmt.Errorf("wrong book service delete find: %w", err)
	}
	if entry.StudentID != studentID {
		return util.NewAppError(constants.CodeForbidden, constants.MsgForbidden)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("wrong book service delete: %w", err)
	}
	return nil
}

// List 分页查询学生错题本。
func (s *WrongBookService) List(ctx context.Context, studentID primitive.ObjectID, filter bson.M, page, pageSize int64) ([]*model.WrongBook, int64, error) {
	filter["student_id"] = studentID
	list, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("wrong book service list: %w", err)
	}
	return list, total, nil
}

// GetByID 查询单个错题（仅本人）。
func (s *WrongBookService) GetByID(ctx context.Context, id, studentID primitive.ObjectID) (*model.WrongBook, error) {
	entry, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeWrongBookNotFound, fmt.Sprintf("错题本模块：id=%s 的条目不存在", id.Hex()))
		}
		return nil, fmt.Errorf("wrong book service get: %w", err)
	}
	if entry.StudentID != studentID {
		return nil, util.NewAppError(constants.CodeForbidden, constants.MsgForbidden)
	}
	return entry, nil
}
