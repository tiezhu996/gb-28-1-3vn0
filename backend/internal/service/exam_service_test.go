package service

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
)

// fakeExamRepo 内存版试卷仓储。
type fakeExamRepo struct {
	exams map[string]*model.Exam
}

func newFakeExamRepo() *fakeExamRepo {
	return &fakeExamRepo{exams: make(map[string]*model.Exam)}
}

func (f *fakeExamRepo) Create(_ context.Context, e *model.Exam) error {
	f.exams[e.ID.Hex()] = e
	return nil
}
func (f *fakeExamRepo) Update(_ context.Context, e *model.Exam) error {
	f.exams[e.ID.Hex()] = e
	return nil
}
func (f *fakeExamRepo) Delete(_ context.Context, id primitive.ObjectID) error {
	delete(f.exams, id.Hex())
	return nil
}
func (f *fakeExamRepo) FindByID(_ context.Context, id primitive.ObjectID) (*model.Exam, error) {
	if e, ok := f.exams[id.Hex()]; ok {
		return e, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeExamRepo) List(_ context.Context, _ bson.M, _, _ int64) ([]*model.Exam, int64, error) {
	var out []*model.Exam
	for _, e := range f.exams {
		out = append(out, e)
	}
	return out, int64(len(out)), nil
}
func (f *fakeExamRepo) CountByStatus(_ context.Context, _ string) (int64, error) {
	return int64(len(f.exams)), nil
}

func newTestExamSvc() *ExamService {
	questionRepo := newFakeQuestionRepo()
	questionSvc := NewQuestionService(questionRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	examRepo := newFakeExamRepo()
	return NewExamService(examRepo, questionSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func seedQuestion(t *testing.T, svc *ExamService) primitive.ObjectID {
	t.Helper()
	req := &dto.CreateQuestionRequest{
		Type: "single", Subject: "数学", KnowledgePoints: []string{"代数"},
		Difficulty: "easy", Content: "1+1=?", Answer: "B", Score: 5,
		Options: []dto.OptionInput{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
	}
	q, err := svc.question.Create(context.Background(), req, primitive.NewObjectID())
	if err != nil {
		t.Fatalf("seed question: %v", err)
	}
	return q.ID
}

func TestExamCreateAndPublish(t *testing.T) {
	svc := newTestExamSvc()
	qid := seedQuestion(t, svc)
	now := time.Now()
	exam, err := svc.Create(context.Background(), &dto.CreateExamRequest{
		Title: "期中考试", Subject: "数学", DurationMin: 60,
		StartAt: now.Add(-time.Hour), EndAt: now.Add(24 * time.Hour),
		Questions: []dto.ExamQuestionInput{{QuestionID: qid.Hex()}},
	}, primitive.NewObjectID())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if exam.Status != constants.ExamStatusDraft {
		t.Fatalf("status = %s, want draft", exam.Status)
	}
	// 发布
	pub, err := svc.Publish(context.Background(), exam.ID, "t@example.com")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if pub.Status != constants.ExamStatusPublished {
		t.Fatalf("status = %s, want published", pub.Status)
	}
	// 非法流转：published -> draft 应失败（状态机校验）
	_, err = svc.Update(context.Background(), exam.ID, &dto.UpdateExamRequest{Status: constants.ExamStatusDraft}, "t@example.com")
	if err == nil {
		t.Fatal("published -> draft 非法流转应失败")
	}
}

func TestExamAutoGenerate(t *testing.T) {
	svc := newTestExamSvc()
	_ = seedQuestion(t, svc)
	exam, err := svc.AutoGenerate(context.Background(), &dto.AutoGenerateRequest{
		Title: "自动组卷测试", Subject: "数学", KnowledgePoints: []string{"代数"},
		DifficultyDist: map[string]int{"easy": 1}, ScorePerQuestion: 5,
		DurationMin: 30, StartAt: time.Now(), EndAt: time.Now().Add(time.Hour),
	}, primitive.NewObjectID())
	if err != nil {
		t.Fatalf("AutoGenerate() error = %v", err)
	}
	if len(exam.Questions) == 0 {
		t.Fatal("自动组卷题目为空")
	}
}
