package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
)

// fakeWrongBookRepo 内存版错题本仓储。
type fakeWrongBookRepo struct {
	items map[string]*model.WrongBook
}

func newFakeWrongBookRepo() *fakeWrongBookRepo {
	return &fakeWrongBookRepo{items: make(map[string]*model.WrongBook)}
}

func (f *fakeWrongBookRepo) Create(_ context.Context, w *model.WrongBook) error {
	f.items[w.ID.Hex()] = w
	return nil
}
func (f *fakeWrongBookRepo) Update(_ context.Context, w *model.WrongBook) error {
	f.items[w.ID.Hex()] = w
	return nil
}
func (f *fakeWrongBookRepo) Delete(_ context.Context, id primitive.ObjectID) error {
	delete(f.items, id.Hex())
	return nil
}
func (f *fakeWrongBookRepo) FindByID(_ context.Context, id primitive.ObjectID) (*model.WrongBook, error) {
	if w, ok := f.items[id.Hex()]; ok {
		return w, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeWrongBookRepo) FindByStudentAndQuestion(_ context.Context, studentID, questionID primitive.ObjectID) (*model.WrongBook, error) {
	for _, w := range f.items {
		if w.StudentID == studentID && w.QuestionID == questionID {
			return w, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeWrongBookRepo) List(_ context.Context, _ bson.M, _, _ int64) ([]*model.WrongBook, int64, error) {
	var out []*model.WrongBook
	for _, w := range f.items {
		out = append(out, w)
	}
	return out, int64(len(out)), nil
}

func TestWrongBookAddAndResolve(t *testing.T) {
	questionRepo := newFakeQuestionRepo()
	questionSvc := NewQuestionService(questionRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	recordRepo := newFakeRecordRepo()
	recordSvc := NewExamRecordService(recordRepo, newTestExamSvc(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	svc := NewWrongBookService(newFakeWrongBookRepo(), questionSvc, recordSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

	student := primitive.NewObjectID()
	q, err := questionSvc.Create(context.Background(), &dto.CreateQuestionRequest{
		Type: "single", Subject: "数学", KnowledgePoints: []string{"代数"},
		Difficulty: "easy", Content: "1+1=?", Answer: "B", Score: 5,
		Options: []dto.OptionInput{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
	}, student)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}
	entry, err := svc.Add(context.Background(), student, q.ID, primitive.NilObjectID, primitive.NilObjectID, "复习")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if entry.Status != constants.WrongBookStatusActive {
		t.Fatalf("status = %s", entry.Status)
	}
	// 幂等
	if _, err := svc.Add(context.Background(), student, q.ID, primitive.NilObjectID, primitive.NilObjectID, "复习"); err != nil {
		t.Fatalf("duplicate Add() error = %v", err)
	}
	updated, err := svc.Update(context.Background(), entry.ID, student, constants.WrongBookStatusResolved, "")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Status != constants.WrongBookStatusResolved {
		t.Fatalf("status = %s", updated.Status)
	}
}
