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

// fakeRecordRepo 内存版考试记录仓储。
type fakeRecordRepo struct {
	records map[string]*model.ExamRecord
}

func newFakeRecordRepo() *fakeRecordRepo {
	return &fakeRecordRepo{records: make(map[string]*model.ExamRecord)}
}

func (f *fakeRecordRepo) Create(_ context.Context, r *model.ExamRecord) error {
	f.records[r.ID.Hex()] = r
	return nil
}
func (f *fakeRecordRepo) Update(_ context.Context, r *model.ExamRecord) error {
	f.records[r.ID.Hex()] = r
	return nil
}
func (f *fakeRecordRepo) FindByID(_ context.Context, id primitive.ObjectID) (*model.ExamRecord, error) {
	if r, ok := f.records[id.Hex()]; ok {
		return r, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeRecordRepo) FindActiveByExamAndStudent(_ context.Context, examID, studentID primitive.ObjectID) (*model.ExamRecord, error) {
	for _, r := range f.records {
		if r.ExamID == examID && r.StudentID == studentID && r.Status == constants.RecordStatusInProgress {
			return r, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeRecordRepo) List(_ context.Context, _ bson.M, _, _ int64) ([]*model.ExamRecord, int64, error) {
	var out []*model.ExamRecord
	for _, r := range f.records {
		out = append(out, r)
	}
	return out, int64(len(out)), nil
}
func (f *fakeRecordRepo) ListAll(_ context.Context, _ bson.M) ([]*model.ExamRecord, error) {
	var out []*model.ExamRecord
	for _, r := range f.records {
		out = append(out, r)
	}
	return out, nil
}
func (f *fakeRecordRepo) CountByExamAndStatus(_ context.Context, _ primitive.ObjectID, _ []string) (int64, error) {
	return int64(len(f.records)), nil
}

func newTestRecordSvc() *ExamRecordService {
	questionRepo := newFakeQuestionRepo()
	questionSvc := NewQuestionService(questionRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	examRepo := newFakeExamRepo()
	examSvc := NewExamService(examRepo, questionSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	recordRepo := newFakeRecordRepo()
	return NewExamRecordService(recordRepo, examSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestStartAndSubmit(t *testing.T) {
	svc := newTestRecordSvc()
	teacher := primitive.NewObjectID()
	student := primitive.NewObjectID()

	// 准备题目与试卷
	singleReq := &dto.CreateQuestionRequest{
		Type: "single", Subject: "数学", KnowledgePoints: []string{"代数"},
		Difficulty: "easy", Content: "1+1=?", Answer: "B", Score: 5,
		Options: []dto.OptionInput{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
	}
	q1, _ := svc.exam.question.Create(context.Background(), singleReq, teacher)
	fillReq := &dto.CreateQuestionRequest{
		Type: "fill", Subject: "数学", KnowledgePoints: []string{"代数"},
		Difficulty: "medium", Content: "圆周率前两位是？", Answer: "3.14", Score: 5,
	}
	q2, _ := svc.exam.question.Create(context.Background(), fillReq, teacher)

	now := time.Now()
	exam, err := svc.exam.Create(context.Background(), &dto.CreateExamRequest{
		Title: "单元测试卷", Subject: "数学", DurationMin: 30, PassScore: 60,
		StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour),
		Questions: []dto.ExamQuestionInput{{QuestionID: q1.ID.Hex()}, {QuestionID: q2.ID.Hex()}},
	}, teacher)
	if err != nil {
		t.Fatalf("create exam: %v", err)
	}
	if _, err := svc.exam.Publish(context.Background(), exam.ID, "t@example.com"); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// 开始考试
	rec, err := svc.StartExam(context.Background(), exam.ID, student, "李同学")
	if err != nil {
		t.Fatalf("StartExam() error = %v", err)
	}
	if rec.Status != constants.RecordStatusInProgress {
		t.Fatalf("status = %s", rec.Status)
	}
	// 再次开始应返回同一记录（幂等）
	rec2, err := svc.StartExam(context.Background(), exam.ID, student, "李同学")
	if err != nil {
		t.Fatalf("second StartExam() error = %v", err)
	}
	if rec2.ID != rec.ID {
		t.Fatal("开始考试不幂等")
	}

	// 提交答卷：客观题答对，主观题留空
	answers := []dto.AnswerInput{{QuestionID: q1.ID.Hex(), Answer: "B"}}
	submitted, err := svc.Submit(context.Background(), rec.ID, answers, 1, nil, false)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if submitted.Status != constants.RecordStatusSubmitted {
		t.Fatalf("status = %s", submitted.Status)
	}
	if submitted.ObjectiveScore != 5 {
		t.Fatalf("objective score = %f, want 5", submitted.ObjectiveScore)
	}
	if submitted.CheatCount != 1 {
		t.Fatalf("cheat count = %d, want 1", submitted.CheatCount)
	}

	// 教师批改主观题
	graded, err := svc.Grade(context.Background(), rec.ID, []dto.GradeItem{{QuestionID: q2.ID.Hex(), Score: 4, Comment: "部分正确"}}, "t@example.com")
	if err != nil {
		t.Fatalf("Grade() error = %v", err)
	}
	if graded.Status != constants.RecordStatusGraded {
		t.Fatalf("status = %s", graded.Status)
	}
	if graded.FinalScore != 9 {
		t.Fatalf("final score = %f, want 9", graded.FinalScore)
	}

	// 重复提交应失败
	if _, err := svc.Submit(context.Background(), rec.ID, answers, 0, nil, false); err == nil {
		t.Fatal("重复提交应失败")
	}
}

func TestReport(t *testing.T) {
	svc := newTestRecordSvc()
	teacher := primitive.NewObjectID()
	student := primitive.NewObjectID()

	q1, _ := svc.exam.question.Create(context.Background(), &dto.CreateQuestionRequest{
		Type: "single", Subject: "数学", KnowledgePoints: []string{"代数"},
		Difficulty: "easy", Content: "1+1=?", Answer: "B", Score: 5,
		Options: []dto.OptionInput{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
	}, teacher)
	now := time.Now()
	exam, _ := svc.exam.Create(context.Background(), &dto.CreateExamRequest{
		Title: "报告测试卷", Subject: "数学", DurationMin: 30, PassScore: 5,
		StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour),
		Questions: []dto.ExamQuestionInput{{QuestionID: q1.ID.Hex()}},
	}, teacher)
	_, _ = svc.exam.Publish(context.Background(), exam.ID, "t@example.com")

	rec, _ := svc.StartExam(context.Background(), exam.ID, student, "李同学")
	_, _ = svc.Submit(context.Background(), rec.ID, []dto.AnswerInput{{QuestionID: q1.ID.Hex(), Answer: "B"}}, 0, nil, false)

	report, err := svc.Report(context.Background(), exam.ID)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if report.TotalStudents != 1 {
		t.Fatalf("total students = %d, want 1", report.TotalStudents)
	}
	if report.MaxScore != 5 {
		t.Fatalf("max score = %f, want 5", report.MaxScore)
	}
}
