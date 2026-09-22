package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
	"github.com/onlineexam/onlineexam/internal/util"
)

// fakeQuestionRepo 内存版题库仓储。
type fakeQuestionRepo struct {
	questions map[string]*model.Question
}

func newFakeQuestionRepo() *fakeQuestionRepo {
	return &fakeQuestionRepo{questions: make(map[string]*model.Question)}
}

func (f *fakeQuestionRepo) Create(_ context.Context, q *model.Question) error {
	f.questions[q.ID.Hex()] = q
	return nil
}
func (f *fakeQuestionRepo) CreateMany(_ context.Context, qs []*model.Question) error {
	for _, q := range qs {
		f.questions[q.ID.Hex()] = q
	}
	return nil
}
func (f *fakeQuestionRepo) Update(_ context.Context, q *model.Question) error {
	f.questions[q.ID.Hex()] = q
	return nil
}
func (f *fakeQuestionRepo) Delete(_ context.Context, id primitive.ObjectID) error {
	delete(f.questions, id.Hex())
	return nil
}
func (f *fakeQuestionRepo) FindByID(_ context.Context, id primitive.ObjectID) (*model.Question, error) {
	if q, ok := f.questions[id.Hex()]; ok {
		return q, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeQuestionRepo) FindByIDs(_ context.Context, ids []primitive.ObjectID) ([]*model.Question, error) {
	var out []*model.Question
	for _, id := range ids {
		if q, ok := f.questions[id.Hex()]; ok {
			out = append(out, q)
		}
	}
	return out, nil
}
func (f *fakeQuestionRepo) List(_ context.Context, _ bson.M, _, _ int64) ([]*model.Question, int64, error) {
	var out []*model.Question
	for _, q := range f.questions {
		out = append(out, q)
	}
	return out, int64(len(out)), nil
}
func (f *fakeQuestionRepo) RandomPick(_ context.Context, _ bson.M, limit int) ([]*model.Question, error) {
	var out []*model.Question
	for _, q := range f.questions {
		if len(out) >= limit {
			break
		}
		out = append(out, q)
	}
	return out, nil
}

func newTestQuestionSvc(repo repository.QuestionRepository) *QuestionService {
	return NewQuestionService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestQuestionCreate(t *testing.T) {
	svc := newTestQuestionSvc(newFakeQuestionRepo())
	creator := primitive.NewObjectID()
	cases := []struct {
		name    string
		req     *dto.CreateQuestionRequest
		wantErr bool
	}{
		{"单选题创建成功", &dto.CreateQuestionRequest{
			Type: "single", Subject: "计算机", KnowledgePoints: []string{"数据结构"},
			Difficulty: "easy", Content: "栈的特点是？",
			Options: []dto.OptionInput{{Key: "A", Text: "先进先出"}, {Key: "B", Text: "先进后出"}},
			Answer:  "B", Score: 5,
		}, false},
		{"非法题型", &dto.CreateQuestionRequest{
			Type: "essay", Subject: "计算机", KnowledgePoints: []string{"x"},
			Difficulty: "easy", Content: "x", Answer: "A", Score: 5,
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.req, creator)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestQuestionImport(t *testing.T) {
	svc := newTestQuestionSvc(newFakeQuestionRepo())
	rows := []util.ExcelQuestionRow{
		{Type: "single", Subject: "数学", KnowledgePoint: "代数", Difficulty: "easy", Content: "1+1=?", Options: []string{"1", "2", "3", "4"}, Answer: "B", Analysis: "", Score: 5},
		{Type: "judge", Subject: "数学", KnowledgePoint: "几何", Difficulty: "hard", Content: "圆周率是无理数", Answer: "true", Analysis: "", Score: 5},
	}
	n, err := svc.Import(context.Background(), rows, primitive.NewObjectID())
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if n != 2 {
		t.Fatalf("import count = %d, want 2", n)
	}
}
