package service_test

import (
	"testing"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/service"
	"golang-backend/utils"
)

type fakeQuestionRepository struct {
	versionExists bool
	created       []entity.Question
}

func (f *fakeQuestionRepository) List(map[string]string, int, int) (*utils.PaginationResult, error) {
	return nil, nil
}

func (f *fakeQuestionRepository) FindByID(string) (*entity.Question, error) {
	return nil, nil
}

func (f *fakeQuestionRepository) ExamVersionExists(string) (bool, error) {
	return f.versionExists, nil
}

func (f *fakeQuestionRepository) CreateWithOptions(*entity.Question, []entity.QuestionOption) error {
	return nil
}

func (f *fakeQuestionRepository) CreateBatchWithOptions(questions []entity.Question) error {
	f.created = append(f.created, questions...)
	return nil
}

func (f *fakeQuestionRepository) UpdateWithOptions(*entity.Question, []entity.QuestionOption) error {
	return nil
}

func (f *fakeQuestionRepository) Delete(*entity.Question) error {
	return nil
}

func TestQuestionServiceImportCreatesSanitizedBatch(t *testing.T) {
	repo := &fakeQuestionRepository{versionExists: true}
	rows := []dto.QuestionImportRow{{
		Row: 2,
		Data: dto.CreateQuestionRequest{
			ExamVersionID: "version-1",
			Code:          "Q1",
			ContentHTML:   `<p>Berapa?</p><script>alert(1)</script>`,
			Weight:        1,
			DisplayOrder:  1,
			Options: []dto.QuestionOptionRequest{
				{OptionKey: "A", ContentHTML: "Satu", DisplayOrder: 1},
				{OptionKey: "B", ContentHTML: "Dua", IsCorrect: true, DisplayOrder: 2},
			},
		},
	}}

	result, err := service.NewQuestionService(repo).Import(rows)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.Imported != 1 || len(repo.created) != 1 {
		t.Fatalf("imported = %d, created = %d; want 1", result.Imported, len(repo.created))
	}
	if repo.created[0].ContentHTML != "<p>Berapa?</p>" || len(repo.created[0].Options) != 2 {
		t.Fatalf("unexpected created question: %+v", repo.created[0])
	}
}

func TestQuestionServiceImportRejectsInvalidRowsWithoutWriting(t *testing.T) {
	repo := &fakeQuestionRepository{versionExists: true}
	rows := []dto.QuestionImportRow{{
		Row: 4,
		Data: dto.CreateQuestionRequest{
			ExamVersionID: "version-1",
			ContentHTML:   "Pertanyaan",
			Weight:        1,
			DisplayOrder:  1,
			Options: []dto.QuestionOptionRequest{
				{OptionKey: "A", ContentHTML: "Satu", IsCorrect: true},
			},
		},
	}}

	result, err := service.NewQuestionService(repo).Import(rows)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.Imported != 0 || len(result.Errors) != 1 || len(repo.created) != 0 {
		t.Fatalf("unexpected import result: %+v, created=%d", result, len(repo.created))
	}
}

var _ repository.QuestionRepository = (*fakeQuestionRepository)(nil)
