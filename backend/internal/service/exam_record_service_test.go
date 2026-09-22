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
func (f *fakeRecordRepo) SaveDraft(_ context.Context, id, studentID primitive.ObjectID, questions []model.AttemptQuestion, currentIndex int, version int64, savedAt time.Time) error {
	r, ok := f.records[id.Hex()]
	if !ok || r.StudentID != studentID || r.Status != constants.RecordStatusInProgress || version <= r.SaveVersion {
		return repository.ErrConflict
	}
	r.Questions = questions
	r.CurrentIndex = currentIndex
	r.SaveVersion = version
	r.LastSavedAt = &savedAt
	r.UpdatedAt = savedAt
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
	submitted, err := svc.Submit(context.Background(), rec.ID, student, answers, 1, nil, false)
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
	if _, err := svc.Submit(context.Background(), rec.ID, student, answers, 0, nil, false); err == nil {
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
	_, _ = svc.Submit(context.Background(), rec.ID, student, []dto.AnswerInput{{QuestionID: q1.ID.Hex(), Answer: "B"}}, 0, nil, false)

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

// prepareRecordForSave 构造一份含两道题、已发布开考的进行中答卷。
func prepareRecordForSave(t *testing.T) (*ExamRecordService, *model.ExamRecord, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID) {
	t.Helper()
	svc := newTestRecordSvc()
	teacher := primitive.NewObjectID()
	student := primitive.NewObjectID()
	q1, _ := svc.exam.question.Create(context.Background(), &dto.CreateQuestionRequest{
		Type: "single", Subject: "数学", KnowledgePoints: []string{"代数"},
		Difficulty: "easy", Content: "1+1=?", Answer: "B", Score: 5,
		Options: []dto.OptionInput{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
	}, teacher)
	q2, _ := svc.exam.question.Create(context.Background(), &dto.CreateQuestionRequest{
		Type: "single", Subject: "数学", KnowledgePoints: []string{"几何"},
		Difficulty: "easy", Content: "三角形有几条边?", Answer: "C", Score: 5,
		Options: []dto.OptionInput{{Key: "C", Text: "3"}, {Key: "D", Text: "4"}},
	}, teacher)
	now := time.Now()
	exam, err := svc.exam.Create(context.Background(), &dto.CreateExamRequest{
		Title: "自动保存测试卷", Subject: "数学", DurationMin: 30, PassScore: 60,
		StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour),
		Questions: []dto.ExamQuestionInput{{QuestionID: q1.ID.Hex()}, {QuestionID: q2.ID.Hex()}},
	}, teacher)
	if err != nil {
		t.Fatalf("create exam: %v", err)
	}
	if _, err := svc.exam.Publish(context.Background(), exam.ID, "t@example.com"); err != nil {
		t.Fatalf("publish: %v", err)
	}
	rec, err := svc.StartExam(context.Background(), exam.ID, student, "李同学")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	return svc, rec, exam.ID, student, q1.ID, q2.ID
}

func TestSaveDraftResumeAndStale(t *testing.T) {
	svc, rec, _, student, q1, q2 := prepareRecordForSave(t)
	ctx := context.Background()

	// 第一次自动保存：第 1 题答 A，停在第 2 题（index=1）
	saved, err := svc.SaveDraft(ctx, rec.ID, student, []dto.AnswerInput{
		{QuestionID: q1.Hex(), Answer: "A"},
	}, 1, 1)
	if err != nil {
		t.Fatalf("SaveDraft() error = %v", err)
	}
	if saved.SaveVersion != 1 || saved.CurrentIndex != 1 {
		t.Fatalf("version=%d index=%d, want 1/1", saved.SaveVersion, saved.CurrentIndex)
	}

	// 乱序旧请求（version=1）必须被忽略，不能覆盖已保存内容
	stale, err := svc.SaveDraft(ctx, rec.ID, student, []dto.AnswerInput{
		{QuestionID: q1.Hex(), Answer: "OLD"},
		{QuestionID: q2.Hex(), Answer: "OLD2"},
	}, 0, 1)
	if err != nil {
		t.Fatalf("stale SaveDraft() error = %v", err)
	}
	if got := stale.Questions[0].UserAnswer; got != "A" {
		t.Fatalf("旧请求覆盖了新答案，q1.user_answer=%s, want A", got)
	}
	if got := stale.Questions[1].UserAnswer; got != "" {
		t.Fatalf("旧请求写入了答案，q2.user_answer=%s, want empty", got)
	}
	if stale.CurrentIndex != 1 {
		t.Fatalf("旧请求改动了当前题号 index=%d, want 1", stale.CurrentIndex)
	}

	// 更新的请求（version=2）正常写入
	newer, err := svc.SaveDraft(ctx, rec.ID, student, []dto.AnswerInput{
		{QuestionID: q1.Hex(), Answer: "B"},
		{QuestionID: q2.Hex(), Answer: "C"},
	}, 1, 2)
	if err != nil {
		t.Fatalf("SaveDraft(v2) error = %v", err)
	}
	if newer.Questions[0].UserAnswer != "B" || newer.Questions[1].UserAnswer != "C" {
		t.Fatalf("v2 未写入，answers=%s/%s", newer.Questions[0].UserAnswer, newer.Questions[1].UserAnswer)
	}

	// 模拟刷新/换设备重进：重新读取记录，答案与当前题号完整恢复
	got, err := svc.GetByID(ctx, rec.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Questions[0].UserAnswer != "B" || got.Questions[1].UserAnswer != "C" {
		t.Fatal("断点续答恢复答案失败")
	}
	if got.CurrentIndex != 1 {
		t.Fatalf("断点续答恢复题号失败 index=%d, want 1", got.CurrentIndex)
	}
	if got.Status != constants.RecordStatusInProgress {
		t.Fatalf("自动保存后状态被改动 status=%s, want in_progress", got.Status)
	}
	if got.LastSavedAt == nil {
		t.Fatal("last_saved_at 未写入")
	}

	// 交卷后不允许再保存
	if _, err := svc.Submit(ctx, rec.ID, student, []dto.AnswerInput{
		{QuestionID: q1.Hex(), Answer: "B"}, {QuestionID: q2.Hex(), Answer: "C"},
	}, 0, nil, false); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if _, err := svc.SaveDraft(ctx, rec.ID, student, nil, 0, 3); err == nil {
		t.Fatal("交卷后自动保存应失败")
	}
}

func TestSaveDraftRejectsNonOwner(t *testing.T) {
	svc, rec, _, _, q1, _ := prepareRecordForSave(t)
	other := primitive.NewObjectID()
	if _, err := svc.SaveDraft(context.Background(), rec.ID, other, []dto.AnswerInput{
		{QuestionID: q1.Hex(), Answer: "X"},
	}, 0, 1); err == nil {
		t.Fatal("非本人自动保存应失败")
	}
}

func TestSaveDraftRejectsExpiredWindow(t *testing.T) {
	svc := newTestRecordSvc()
	teacher := primitive.NewObjectID()
	student := primitive.NewObjectID()
	q1, _ := svc.exam.question.Create(context.Background(), &dto.CreateQuestionRequest{
		Type: "single", Subject: "数学", KnowledgePoints: []string{"代数"},
		Difficulty: "easy", Content: "1+1=?", Answer: "B", Score: 5,
		Options: []dto.OptionInput{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
	}, teacher)
	now := time.Now()
	// 考试窗口 2 小时前就已结束
	exam, _ := svc.exam.Create(context.Background(), &dto.CreateExamRequest{
		Title: "已结束考试", Subject: "数学", DurationMin: 30, PassScore: 60,
		StartAt: now.Add(-3 * time.Hour), EndAt: now.Add(-2 * time.Hour),
		Questions: []dto.ExamQuestionInput{{QuestionID: q1.ID.Hex()}},
	}, teacher)
	_, _ = svc.exam.Publish(context.Background(), exam.ID, "t@example.com")
	// 直接插入一条 90 分钟前开始的进行中记录（时长 30 分钟、窗口 2 小时前结束），模拟“已过期但未交卷”
	rec := &model.ExamRecord{
		ID: primitive.NewObjectID(), ExamID: exam.ID, ExamTitle: exam.Title,
		StudentID: student, StudentName: "李同学",
		Questions: []model.AttemptQuestion{{
			QuestionID: q1.ID, Type: "single", Content: "1+1=?", Score: 5,
			CorrectAnswer: "B", Result: constants.AnswerResultUnmarked,
		}},
		Status: constants.RecordStatusInProgress, StartedAt: now.Add(-90 * time.Minute),
		CreatedAt: now.Add(-90 * time.Minute), UpdatedAt: now.Add(-90 * time.Minute),
	}
	if err := svc.repo.Create(context.Background(), rec); err != nil {
		t.Fatalf("create expired record: %v", err)
	}
	if _, err := svc.SaveDraft(context.Background(), rec.ID, student, []dto.AnswerInput{
		{QuestionID: q1.ID.Hex(), Answer: "A"},
	}, 0, 1); err == nil {
		t.Fatal("超过截止时间的自动保存应失败")
	}
}
