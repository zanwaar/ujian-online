package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/service"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type fakeAttemptRepository struct {

	startContext           *repository.AttemptStartContext
	startErr               error
	activeAttempt          *entity.Attempt
	activeErr              error
	completedAttempt       *entity.Attempt
	completedErr           error
	questions              []entity.Question
	attemptCount           int64
	createdAttempt         *entity.Attempt
	createdQuestions       []entity.AttemptQuestion
	findAttempt            *entity.Attempt
	findAttemptErr         error
	findAttemptCandidateID string
	savedOption            string
	saveAnswerCandidateID  string
	submitAttempt          *entity.Attempt
	submitCalls            int
	resultAttempt          *entity.Attempt
	resultShowResult       bool
	resultErr              error
	adminDetail            *repository.AdminAttemptDetailData
	adminDetailErr         error
}

func newFakeAttemptRepository() *fakeAttemptRepository {
	return &fakeAttemptRepository{
		activeErr:    gorm.ErrRecordNotFound,
		completedErr: gorm.ErrRecordNotFound,
	}
}

func (f *fakeAttemptRepository) ListSchedules(string) ([]dto.CandidateExamSchedule, error) {
	return nil, nil
}

func (f *fakeAttemptRepository) ListAdminResults(map[string]string, int, int) (*utils.PaginationResult, error) {
	return nil, nil
}

func (f *fakeAttemptRepository) FindAdminResultDetail(string) (*repository.AdminAttemptDetailData, error) {
	return f.adminDetail, f.adminDetailErr
}

func (f *fakeAttemptRepository) FindStartContext(string, string) (*repository.AttemptStartContext, error) {
	return f.startContext, f.startErr
}

func (f *fakeAttemptRepository) FindActiveAttempt(string) (*entity.Attempt, error) {
	return f.activeAttempt, f.activeErr
}

func (f *fakeAttemptRepository) FindCompletedAttempt(string) (*entity.Attempt, error) {
	return f.completedAttempt, f.completedErr
}

func (f *fakeAttemptRepository) CountAttempts(string) (int64, error) {
	return f.attemptCount, nil
}

func (f *fakeAttemptRepository) FindQuestions(string) ([]entity.Question, error) {
	return f.questions, nil
}

func (f *fakeAttemptRepository) CreateAttemptWithSnapshot(attempt *entity.Attempt, questions []entity.AttemptQuestion) error {
	attempt.ID = "attempt-created"
	for index := range questions {
		questions[index].ID = "attempt-question-" + string(rune('1'+index))
		questions[index].AttemptID = attempt.ID
	}
	f.createdAttempt = attempt
	f.createdQuestions = questions
	f.findAttempt = &entity.Attempt{
		Base:          entity.Base{ID: attempt.ID},
		ExamID:        f.startContext.ExamID,
		AssignmentID:  attempt.AssignmentID,
		AttemptNumber: attempt.AttemptNumber,
		StartedAt:     attempt.StartedAt,
		DeadlineAt:    attempt.DeadlineAt,
		Status:        attempt.Status,
		Questions:     questions,
	}
	return nil
}

func (f *fakeAttemptRepository) FindAttemptForCandidate(_, candidateID string) (*entity.Attempt, error) {
	f.findAttemptCandidateID = candidateID
	return f.findAttempt, f.findAttemptErr
}

func (f *fakeAttemptRepository) SaveAnswer(_, candidateID, _, selectedOptionKey string, _ time.Time) error {
	f.saveAnswerCandidateID = candidateID
	f.savedOption = selectedOptionKey
	return nil
}

func (f *fakeAttemptRepository) SubmitAttempt(_, _ string, _ time.Time) (*entity.Attempt, error) {
	f.submitCalls++
	if f.submitAttempt != nil {
		f.findAttempt = f.submitAttempt
		return f.submitAttempt, nil
	}
	return f.findAttempt, nil
}

func (f *fakeAttemptRepository) FindResultForCandidate(string, string) (*entity.Attempt, bool, error) {
	return f.resultAttempt, f.resultShowResult, f.resultErr
}

func validStartContext(now time.Time) *repository.AttemptStartContext {
	return &repository.AttemptStartContext{
		AssignmentID:    "assignment-1",
		ExamID:          "exam-1",
		DurationMinutes: 10,
		SessionStartsAt: now.Add(-time.Minute),
		SessionEndsAt:   now.Add(30 * time.Minute),
		VersionID:       "version-1",
	}
}

func oneQuestion() []entity.Question {
	return []entity.Question{
		{
			Base:        entity.Base{ID: "question-1"},
			Code:        "Q1",
			ContentHTML: "<p>2 + 2 = ?</p>",
			Options: []entity.QuestionOption{
				{Base: entity.Base{ID: "option-a"}, OptionKey: "A", ContentHTML: "3"},
				{Base: entity.Base{ID: "option-b"}, OptionKey: "B", ContentHTML: "4", IsCorrect: true},
			},
		},
	}
}

func TestAttemptServiceStartCreatesServerTimedSnapshot(t *testing.T) {
	now := time.Now().UTC()
	repo := newFakeAttemptRepository()
	repo.startContext = validStartContext(now)
	repo.questions = oneQuestion()

	view, err := service.NewAttemptService(repo).Start("candidate-1", "exam-1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if view.ID != "attempt-created" || view.ExamID != "exam-1" {
		t.Fatalf("unexpected attempt identity: %+v", view)
	}
	if len(repo.createdQuestions) != 1 {
		t.Fatalf("snapshot question count = %d, want 1", len(repo.createdQuestions))
	}
	if !repo.createdAttempt.DeadlineAt.Before(repo.startContext.SessionEndsAt) {
		t.Fatal("deadline should be bounded by the session end")
	}
	if got := repo.createdQuestions[0].SnapshotContentJSON; !strings.Contains(got, `"correct_option_key":"B"`) {
		t.Fatalf("snapshot does not preserve correct option: %s", got)
	}
}

func TestAttemptServiceStartRejectsCompletedAttempt(t *testing.T) {
	repo := newFakeAttemptRepository()
	repo.startContext = validStartContext(time.Now().UTC())
	repo.completedAttempt = &entity.Attempt{Base: entity.Base{ID: "attempt-done"}, Status: "submitted"}
	repo.completedErr = nil

	_, err := service.NewAttemptService(repo).Start("candidate-1", "exam-1")
	if err == nil || !strings.Contains(err.Error(), "already been submitted") {
		t.Fatalf("Start() error = %v, want completed-attempt rejection", err)
	}
}

func TestAttemptServiceStartPropagatesActiveAttemptRepositoryError(t *testing.T) {
	repo := newFakeAttemptRepository()
	repo.startContext = validStartContext(time.Now().UTC())
	repo.activeErr = errors.New("database unavailable")

	_, err := service.NewAttemptService(repo).Start("candidate-1", "exam-1")
	if err == nil || err.Error() != "database unavailable" {
		t.Fatalf("Start() error = %v, want repository error", err)
	}
}

func TestAttemptServiceAnswerNormalizesOptionKey(t *testing.T) {
	repo := newFakeAttemptRepository()
	if err := service.NewAttemptService(repo).Answer("candidate-1", "attempt-1", "question-1", " b "); err != nil {
		t.Fatalf("Answer() error = %v", err)
	}
	if repo.savedOption != "B" {
		t.Fatalf("saved option = %q, want %q", repo.savedOption, "B")
	}
}

func TestAttemptServiceGetScopesAttemptLookupToCandidate(t *testing.T) {
	repo := newFakeAttemptRepository()
	repo.findAttemptErr = gorm.ErrRecordNotFound

	if _, err := service.NewAttemptService(repo).Get("candidate-2", "attempt-1"); err == nil {
		t.Fatal("Get() should reject an attempt that is not owned by the candidate")
	}
	if repo.findAttemptCandidateID != "candidate-2" {
		t.Fatalf("candidate ID = %q, want candidate-2", repo.findAttemptCandidateID)
	}
}

func TestAttemptServiceAnswerScopesSaveToCandidate(t *testing.T) {
	repo := newFakeAttemptRepository()
	if err := service.NewAttemptService(repo).Answer("candidate-2", "attempt-1", "question-1", "A"); err != nil {
		t.Fatalf("Answer() error = %v", err)
	}
	if repo.saveAnswerCandidateID != "candidate-2" {
		t.Fatalf("candidate ID = %q, want candidate-2", repo.saveAnswerCandidateID)
	}
}

func TestAttemptServiceGetAutoSubmitsExpiredAttempt(t *testing.T) {
	expired := time.Now().UTC().Add(-time.Minute)
	submittedAt := time.Now().UTC()
	repo := newFakeAttemptRepository()
	repo.findAttempt = &entity.Attempt{
		Base:       entity.Base{ID: "attempt-1"},
		Status:     "in_progress",
		DeadlineAt: &expired,
	}
	repo.submitAttempt = &entity.Attempt{
		Base:        entity.Base{ID: "attempt-1"},
		Status:      "submitted",
		SubmittedAt: &submittedAt,
	}

	view, err := service.NewAttemptService(repo).Get("candidate-1", "attempt-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if repo.submitCalls != 1 || view.Status != "submitted" {
		t.Fatalf("submit_calls = %d, status = %q", repo.submitCalls, view.Status)
	}
}

func TestAttemptServiceResultIncludesExamIDAndHonorsVisibility(t *testing.T) {
	submittedAt := time.Now().UTC()
	repo := newFakeAttemptRepository()
	repo.resultAttempt = &entity.Attempt{
		Base:         entity.Base{ID: "attempt-1"},
		ExamID:       "exam-1",
		Status:       "submitted",
		SubmittedAt:  &submittedAt,
		CorrectCount: 8,
		Score:        80,
	}
	repo.resultShowResult = true

	result, err := service.NewAttemptService(repo).Result("candidate-1", "attempt-1")
	if err != nil {
		t.Fatalf("Result() error = %v", err)
	}
	if result.ExamID != "exam-1" || result.CorrectCount != 8 || result.Score != 80 {
		t.Fatalf("unexpected visible result: %+v", result)
	}

	repo.resultShowResult = false
	result, err = service.NewAttemptService(repo).Result("candidate-1", "attempt-1")
	if err != nil {
		t.Fatalf("Result() hidden error = %v", err)
	}
	if result.ExamID != "exam-1" || result.CorrectCount != 0 || result.Score != 0 {
		t.Fatalf("unexpected hidden result: %+v", result)
	}
}

func TestAttemptServiceGetAdminResultDetailUsesAttemptSnapshot(t *testing.T) {
	repo := newFakeAttemptRepository()
	repo.adminDetail = &repository.AdminAttemptDetailData{
		Result: dto.AdminAttemptResult{
			AttemptID:     "attempt-1",
			CandidateName: "Siswa Satu",
			Score:         100,
		},
		Questions: []entity.AttemptQuestion{{
			Base:                entity.Base{ID: "attempt-question-1"},
			SequenceNumber:      1,
			SnapshotContentJSON: `{"code":"Q1","content_html":"<p>2 + 2 = ?</p>","correct_option_key":"B","options":[{"option_key":"A","content_html":"3"},{"option_key":"B","content_html":"4"}]}`,
		}},
		Answers: []entity.AttemptAnswer{{
			AttemptQuestionID: "attempt-question-1",
			SelectedOptionKey: "B",
			IsCorrect:         true,
		}},
	}

	detail, err := service.NewAttemptService(repo).GetAdminResultDetail("attempt-1")
	if err != nil {
		t.Fatalf("GetAdminResultDetail() error = %v", err)
	}
	if detail.AttemptID != "attempt-1" || len(detail.Questions) != 1 {
		t.Fatalf("unexpected detail: %+v", detail)
	}
	question := detail.Questions[0]
	if question.SelectedOptionKey == nil || *question.SelectedOptionKey != "B" || !question.IsCorrect {
		t.Fatalf("unexpected answer detail: %+v", question)
	}
	if len(question.Options) != 2 || !question.Options[1].IsCorrect || question.Options[0].IsCorrect {
		t.Fatalf("unexpected option correctness: %+v", question.Options)
	}
}
