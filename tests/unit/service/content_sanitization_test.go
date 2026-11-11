package service_test

import (
	"testing"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/service"
	"golang-backend/utils"
)

type sanitizationExamRepository struct {
	createdExam *entity.Exam
}

func (f *sanitizationExamRepository) List(map[string]string, int, int) (*utils.PaginationResult, error) {
	return nil, nil
}

func (f *sanitizationExamRepository) FindByID(string) (*entity.Exam, error) {
	return nil, nil
}

func (f *sanitizationExamRepository) FindCurrentVersion(string) (*entity.ExamVersion, error) {
	return nil, nil
}

func (f *sanitizationExamRepository) CreateWithVersion(exam *entity.Exam, _ *entity.ExamVersion) error {
	f.createdExam = exam
	return nil
}

func (f *sanitizationExamRepository) Update(*entity.Exam) error {
	return nil
}

func (f *sanitizationExamRepository) UpdateVersion(*entity.ExamVersion) error {
	return nil
}

func (f *sanitizationExamRepository) Delete(*entity.Exam) error {
	return nil
}

var _ repository.ExamRepository = (*sanitizationExamRepository)(nil)

type sanitizationQuestionRepository struct {
	createdQuestion *entity.Question
	createdOptions  []entity.QuestionOption
}

func (f *sanitizationQuestionRepository) List(map[string]string, int, int) (*utils.PaginationResult, error) {
	return nil, nil
}

func (f *sanitizationQuestionRepository) FindByID(string) (*entity.Question, error) {
	return nil, nil
}

func (f *sanitizationQuestionRepository) ExamVersionExists(string) (bool, error) {
	return true, nil
}

func (f *sanitizationQuestionRepository) CreateWithOptions(question *entity.Question, options []entity.QuestionOption) error {
	f.createdQuestion = question
	f.createdOptions = options
	return nil
}

func (f *sanitizationQuestionRepository) UpdateWithOptions(*entity.Question, []entity.QuestionOption) error {
	return nil
}

func (f *sanitizationQuestionRepository) Delete(*entity.Question) error {
	return nil
}

var _ repository.QuestionRepository = (*sanitizationQuestionRepository)(nil)

func TestExamServiceSanitizesDescriptionBeforePersistence(t *testing.T) {
	repo := &sanitizationExamRepository{}

	_, err := service.NewExamService(repo).Create(dto.CreateExamRequest{
		Code:            "SEC-001",
		Name:            "Ujian keamanan",
		Subject:         "TIK",
		DescriptionHTML: `<p>Petunjuk <strong>aman</strong></p><script>alert(1)</script>`,
		DurationMinutes: 30,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repo.createdExam == nil {
		t.Fatal("Create() did not pass an exam to the repository")
	}
	if repo.createdExam.DescriptionHTML != `<p>Petunjuk <strong>aman</strong></p>` {
		t.Fatalf("unexpected sanitized description: %q", repo.createdExam.DescriptionHTML)
	}
}

func TestQuestionServiceSanitizesQuestionAndOptionsBeforePersistence(t *testing.T) {
	repo := &sanitizationQuestionRepository{}

	_, err := service.NewQuestionService(repo).Create(dto.CreateQuestionRequest{
		ExamVersionID: "version-1",
		ContentHTML:   `<p>Pertanyaan</p><svg onload="alert(1)"></svg>`,
		Weight:        1,
		Options: []dto.QuestionOptionRequest{
			{OptionKey: "a", ContentHTML: `<strong>Benar</strong><script>alert(1)</script>`, IsCorrect: true},
			{OptionKey: "b", ContentHTML: `<em>Salah</em><img src=x onerror=alert(2)>`},
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repo.createdQuestion == nil || len(repo.createdOptions) != 2 {
		t.Fatal("Create() did not pass question and options to the repository")
	}
	if repo.createdQuestion.ContentHTML != `<p>Pertanyaan</p>` {
		t.Fatalf("unexpected sanitized question: %q", repo.createdQuestion.ContentHTML)
	}
	if repo.createdOptions[0].ContentHTML != `<strong>Benar</strong>` {
		t.Fatalf("unexpected sanitized first option: %q", repo.createdOptions[0].ContentHTML)
	}
	if repo.createdOptions[1].ContentHTML != `<em>Salah</em>` {
		t.Fatalf("unexpected sanitized second option: %q", repo.createdOptions[1].ContentHTML)
	}
}
