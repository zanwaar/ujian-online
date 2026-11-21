package repository

import (
	"os"
	"sync"
	"testing"
	"time"

	"golang-backend/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type submitFixture struct {
	db           *gorm.DB
	candidateID  string
	attemptID    string
	assignmentID string
}

func openSubmitTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get test database handle: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	return db
}

func createSubmitFixture(t *testing.T, db *gorm.DB) submitFixture {
	t.Helper()

	now := time.Now().UTC()
	exam := &entity.Exam{
		Code:            "TEST-SUBMIT-" + now.Format("150405.000000000"),
		Name:            "Submit integration test",
		Subject:         "Testing",
		DescriptionHTML: "<p>Test</p>",
		DurationMinutes: 30,
		ShowResult:      true,
		Status:          "published",
	}
	if err := db.Create(exam).Error; err != nil {
		t.Fatalf("create exam: %v", err)
	}

	version := &entity.ExamVersion{
		ExamID:        exam.ID,
		VersionNumber: 1,
		Status:        "published",
		PublishedAt:   &now,
	}
	if err := db.Create(version).Error; err != nil {
		t.Fatalf("create exam version: %v", err)
	}
	if err := db.Model(exam).Update("current_version_id", version.ID).Error; err != nil {
		t.Fatalf("set current exam version: %v", err)
	}

	question := &entity.Question{
		ExamVersionID: version.ID,
		Code:          "Q1",
		ContentHTML:   "<p>Question</p>",
		Weight:        1,
		DisplayOrder:  1,
		Status:        "active",
	}
	if err := db.Create(question).Error; err != nil {
		t.Fatalf("create question: %v", err)
	}
	for index, optionKey := range []string{"A", "B"} {
		if err := db.Create(&entity.QuestionOption{
			QuestionID:   question.ID,
			OptionKey:    optionKey,
			ContentHTML:  "<p>Option</p>",
			IsCorrect:    optionKey == "A",
			DisplayOrder: index + 1,
		}).Error; err != nil {
			t.Fatalf("create question option: %v", err)
		}
	}

	session := &entity.ExamSession{
		ExamID:   exam.ID,
		Name:     "Submit test session",
		StartsAt: now.Add(-time.Hour),
		EndsAt:   now.Add(time.Hour),
		Status:   "active",
	}
	if err := db.Create(session).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}

	candidate := &entity.Candidate{
		ExamNumber:   "TEST-SUBMIT-" + now.Format("150405.000000000"),
		PasswordHash: "not-used",
		Name:         "Submit test candidate",
		Gender:       "M",
		School:       "Test school",
		ClassName:    "Test class",
		Status:       "active",
	}
	if err := db.Create(candidate).Error; err != nil {
		t.Fatalf("create candidate: %v", err)
	}

	assignment := &entity.ExamAssignment{
		CandidateID:   candidate.ID,
		ExamID:        exam.ID,
		ExamSessionID: session.ID,
		Status:        "assigned",
	}
	if err := db.Create(assignment).Error; err != nil {
		t.Fatalf("create assignment: %v", err)
	}

	startedAt := now.Add(-5 * time.Minute)
	deadlineAt := now.Add(time.Hour)
	attempt := &entity.Attempt{
		AssignmentID:  assignment.ID,
		AttemptNumber: 1,
		StartedAt:     &startedAt,
		DeadlineAt:    &deadlineAt,
		Status:        "in_progress",
	}
	if err := db.Create(attempt).Error; err != nil {
		t.Fatalf("create attempt: %v", err)
	}

	if err := db.Create(&entity.AttemptQuestion{
		AttemptID:           attempt.ID,
		QuestionID:          question.ID,
		SequenceNumber:      1,
		OptionOrderJSON:     `[{"option_key":"A"},{"option_key":"B"}]`,
		SnapshotContentJSON: `{"correct_option_key":"A","options":[{"option_key":"A"},{"option_key":"B"}]}`,
	}).Error; err != nil {
		t.Fatalf("create attempt question: %v", err)
	}

	fixture := submitFixture{
		db:           db,
		candidateID:  candidate.ID,
		attemptID:    attempt.ID,
		assignmentID: assignment.ID,
	}
	t.Cleanup(func() {
		cleanupSubmitFixture(t, fixture, exam.ID, version.ID, question.ID, session.ID, candidate.ID)
	})
	return fixture
}

func cleanupSubmitFixture(t *testing.T, fixture submitFixture, examID, versionID, questionID, sessionID, candidateID string) {
	t.Helper()
	for _, operation := range []func() error{
		func() error {
			return fixture.db.Unscoped().Where("attempt_id = ?", fixture.attemptID).Delete(&entity.AttemptAnswer{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("attempt_id = ?", fixture.attemptID).Delete(&entity.AttemptQuestion{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("id = ?", fixture.attemptID).Delete(&entity.Attempt{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("id = ?", fixture.assignmentID).Delete(&entity.ExamAssignment{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("question_id = ?", questionID).Delete(&entity.QuestionOption{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("id = ?", questionID).Delete(&entity.Question{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("id = ?", sessionID).Delete(&entity.ExamSession{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("id = ?", candidateID).Delete(&entity.Candidate{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("id = ?", versionID).Delete(&entity.ExamVersion{}).Error
		},
		func() error {
			return fixture.db.Unscoped().Where("id = ?", examID).Delete(&entity.Exam{}).Error
		},
	} {
		if err := operation(); err != nil {
			t.Errorf("cleanup submit fixture: %v", err)
		}
	}
}

func TestSubmitAttemptIsIdempotent(t *testing.T) {
	db := openSubmitTestDB(t)
	fixture := createSubmitFixture(t, db)
	repo := NewAttemptRepository(db)
	firstSubmittedAt := time.Now().UTC()

	first, err := repo.SubmitAttempt(fixture.attemptID, fixture.candidateID, firstSubmittedAt)
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}
	if first.Status != "submitted" || first.SubmittedAt == nil {
		t.Fatalf("first submit result = %#v, want submitted attempt", first)
	}

	second, err := repo.SubmitAttempt(fixture.attemptID, fixture.candidateID, firstSubmittedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("duplicate submit: %v", err)
	}
	if second.Status != "submitted" || second.SubmittedAt == nil {
		t.Fatalf("duplicate submit result = %#v, want submitted attempt", second)
	}
	if !second.SubmittedAt.Equal(*first.SubmittedAt) {
		t.Fatalf("duplicate submit changed submitted_at from %v to %v", first.SubmittedAt, second.SubmittedAt)
	}

	var persisted entity.Attempt
	if err := db.First(&persisted, "id = ?", fixture.attemptID).Error; err != nil {
		t.Fatalf("reload attempt: %v", err)
	}
	if persisted.Status != "submitted" {
		t.Fatalf("persisted status = %q, want submitted", persisted.Status)
	}
}

func TestSubmitAttemptSerializesConcurrentRequests(t *testing.T) {
	db := openSubmitTestDB(t)
	fixture := createSubmitFixture(t, db)
	repo := NewAttemptRepository(db)

	start := make(chan struct{})
	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for index := 0; index < 2; index++ {
		waitGroup.Add(1)
		go func(offset time.Duration) {
			defer waitGroup.Done()
			<-start
			_, err := repo.SubmitAttempt(fixture.attemptID, fixture.candidateID, time.Now().UTC().Add(offset))
			results <- err
		}(time.Duration(index) * time.Second)
	}
	close(start)
	waitGroup.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Fatalf("concurrent submit: %v", err)
		}
	}

	var persisted entity.Attempt
	if err := db.First(&persisted, "id = ?", fixture.attemptID).Error; err != nil {
		t.Fatalf("reload concurrent attempt: %v", err)
	}
	if persisted.Status != "submitted" || persisted.SubmittedAt == nil {
		t.Fatalf("persisted concurrent result = %#v, want one submitted attempt", persisted)
	}
}
