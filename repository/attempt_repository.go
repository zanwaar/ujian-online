package repository

import (
	"encoding/json"
	"errors"
	"time"

	"golang-backend/dto"
	"golang-backend/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AttemptStartContext struct {
	AssignmentID     string    `gorm:"column:assignment_id"`
	ExamID           string    `gorm:"column:exam_id"`
	ExamName         string    `gorm:"column:exam_name"`
	Subject          string    `gorm:"column:subject"`
	DurationMinutes  int       `gorm:"column:duration_minutes"`
	ShowResult       bool      `gorm:"column:show_result"`
	SessionID        string    `gorm:"column:session_id"`
	SessionStartsAt  time.Time `gorm:"column:session_starts_at"`
	SessionEndsAt    time.Time `gorm:"column:session_ends_at"`
	VersionID        string    `gorm:"column:version_id"`
	ShuffleQuestions bool      `gorm:"column:shuffle_questions"`
	ShuffleOptions   bool      `gorm:"column:shuffle_options"`
}

type AttemptRepository interface {
	ListSchedules(candidateID string) ([]dto.CandidateExamSchedule, error)
	FindStartContext(candidateID, examID string) (*AttemptStartContext, error)
	FindActiveAttempt(assignmentID string) (*entity.Attempt, error)
	FindCompletedAttempt(assignmentID string) (*entity.Attempt, error)
	CountAttempts(assignmentID string) (int64, error)
	FindQuestions(versionID string) ([]entity.Question, error)
	CreateAttemptWithSnapshot(attempt *entity.Attempt, questions []entity.AttemptQuestion) error
	FindAttemptForCandidate(attemptID, candidateID string) (*entity.Attempt, error)
	SaveAnswer(attemptID, candidateID, questionID, selectedOptionKey string, answeredAt time.Time) error
	SubmitAttempt(attemptID, candidateID string, submittedAt time.Time) (*entity.Attempt, error)
	FindResultForCandidate(attemptID, candidateID string) (*entity.Attempt, bool, error)
}

type attemptRepository struct {
	db *gorm.DB
}

func NewAttemptRepository(db *gorm.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

func (r *attemptRepository) ListSchedules(candidateID string) ([]dto.CandidateExamSchedule, error) {
	var schedules []dto.CandidateExamSchedule
	err := r.db.Table("exam_assignments ea").
		Select(`ea.id AS assignment_id, e.id AS exam_id, e.name AS exam_name, e.subject,
			es.id AS session_id, es.name AS session_name, es.starts_at, es.ends_at,
			e.duration_minutes, e.show_result, a.id AS attempt_id, a.status AS attempt_status, a.submitted_at`).
		Joins("JOIN exams e ON e.id = ea.exam_id AND e.deleted_at IS NULL").
		Joins("JOIN exam_sessions es ON es.id = ea.exam_session_id AND es.deleted_at IS NULL").
		Joins("LEFT JOIN attempts a ON a.assignment_id = ea.id AND a.deleted_at IS NULL").
		Where("ea.candidate_id = ? AND ea.status <> ?", candidateID, "archived").
		Order("es.starts_at ASC, a.attempt_number DESC").
		Scan(&schedules).Error
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(schedules))
	deduplicated := make([]dto.CandidateExamSchedule, 0, len(schedules))
	for _, schedule := range schedules {
		if _, exists := seen[schedule.AssignmentID]; exists {
			continue
		}
		seen[schedule.AssignmentID] = struct{}{}
		deduplicated = append(deduplicated, schedule)
	}
	return deduplicated, nil
}

func (r *attemptRepository) FindStartContext(candidateID, examID string) (*AttemptStartContext, error) {
	var context AttemptStartContext
	err := r.db.Table("exam_assignments ea").
		Select(`ea.id AS assignment_id, e.id AS exam_id, e.name AS exam_name, e.subject,
			e.duration_minutes, e.show_result, es.id AS session_id, es.starts_at AS session_starts_at,
			es.ends_at AS session_ends_at, ev.id AS version_id, ev.shuffle_questions, ev.shuffle_options`).
		Joins("JOIN exams e ON e.id = ea.exam_id AND e.deleted_at IS NULL").
		Joins("JOIN exam_sessions es ON es.id = ea.exam_session_id AND es.deleted_at IS NULL").
		Joins("JOIN exam_versions ev ON ev.id = e.current_version_id AND ev.deleted_at IS NULL").
		Where("ea.candidate_id = ? AND ea.exam_id = ? AND ea.status = ?", candidateID, examID, "assigned").
		Take(&context).Error
	return &context, err
}

func (r *attemptRepository) FindActiveAttempt(assignmentID string) (*entity.Attempt, error) {
	var attempt entity.Attempt
	err := r.db.Where("assignment_id = ? AND status = ?", assignmentID, "in_progress").First(&attempt).Error
	return &attempt, err
}

func (r *attemptRepository) FindCompletedAttempt(assignmentID string) (*entity.Attempt, error) {
	var attempt entity.Attempt
	err := r.db.Where("assignment_id = ? AND status IN ?", assignmentID, []string{"submitted", "expired"}).Order("attempt_number DESC").First(&attempt).Error
	return &attempt, err
}

func (r *attemptRepository) CountAttempts(assignmentID string) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Attempt{}).Where("assignment_id = ?", assignmentID).Count(&count).Error
	return count, err
}

func (r *attemptRepository) FindQuestions(versionID string) ([]entity.Question, error) {
	var questions []entity.Question
	err := r.db.Preload("Options", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC, option_key ASC")
	}).Where("exam_version_id = ? AND status = ?", versionID, "active").Order("display_order ASC, created_at ASC").Find(&questions).Error
	return questions, err
}

func (r *attemptRepository) CreateAttemptWithSnapshot(attempt *entity.Attempt, questions []entity.AttemptQuestion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(attempt).Error; err != nil {
			return err
		}
		for index := range questions {
			questions[index].AttemptID = attempt.ID
			if err := tx.Create(&questions[index]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *attemptRepository) FindAttemptForCandidate(attemptID, candidateID string) (*entity.Attempt, error) {
	var attempt entity.Attempt
	err := r.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sequence_number ASC")
	}).Preload("Answers").
		Table("attempts a").
		Joins("JOIN exam_assignments ea ON ea.id = a.assignment_id AND ea.deleted_at IS NULL").
		Select("a.*, ea.exam_id AS exam_id").
		Where("a.id = ? AND ea.candidate_id = ? AND a.deleted_at IS NULL", attemptID, candidateID).
		First(&attempt).Error
	return &attempt, err
}

func (r *attemptRepository) SaveAnswer(attemptID, candidateID, questionID, selectedOptionKey string, answeredAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var attempt entity.Attempt
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Table("attempts a").
			Joins("JOIN exam_assignments ea ON ea.id = a.assignment_id").
			Where("a.id = ? AND ea.candidate_id = ?", attemptID, candidateID).First(&attempt).Error; err != nil {
			return err
		}
		if attempt.Status != "in_progress" {
			return errors.New("attempt is not active")
		}
		if attempt.DeadlineAt != nil && !answeredAt.Before(*attempt.DeadlineAt) {
			return errors.New("attempt deadline has passed")
		}

		var attemptQuestion entity.AttemptQuestion
		if err := tx.Where("id = ? AND attempt_id = ?", questionID, attemptID).First(&attemptQuestion).Error; err != nil {
			return err
		}
		var snapshot struct {
			CorrectOptionKey string `json:"correct_option_key"`
			Options          []struct {
				OptionKey string `json:"option_key"`
			} `json:"options"`
		}
		if err := json.Unmarshal([]byte(attemptQuestion.SnapshotContentJSON), &snapshot); err != nil {
			return err
		}
		validOption := false
		for _, option := range snapshot.Options {
			if option.OptionKey == selectedOptionKey {
				validOption = true
				break
			}
		}
		if !validOption {
			return errors.New("selected option is invalid")
		}
		isCorrect := selectedOptionKey == snapshot.CorrectOptionKey

		var answer entity.AttemptAnswer
		err := tx.Where("attempt_question_id = ?", attemptQuestion.ID).First(&answer).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			answer = entity.AttemptAnswer{
				AttemptID:         attemptID,
				AttemptQuestionID: attemptQuestion.ID,
			}
		} else if err != nil {
			return err
		}
		answer.SelectedOptionKey = selectedOptionKey
		answer.AnsweredAt = &answeredAt
		answer.IsCorrect = isCorrect
		if answer.ID == "" {
			return tx.Create(&answer).Error
		}
		return tx.Save(&answer).Error
	})
}

func (r *attemptRepository) SubmitAttempt(attemptID, candidateID string, submittedAt time.Time) (*entity.Attempt, error) {
	var result entity.Attempt
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var attempt entity.Attempt
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Table("attempts a").
			Joins("JOIN exam_assignments ea ON ea.id = a.assignment_id").
			Where("a.id = ? AND ea.candidate_id = ?", attemptID, candidateID).First(&attempt).Error; err != nil {
			return err
		}
		if attempt.Status == "submitted" || attempt.Status == "expired" {
			result = attempt
			return nil
		}
		var totalQuestions int64
		if err := tx.Model(&entity.AttemptQuestion{}).Where("attempt_id = ?", attempt.ID).Count(&totalQuestions).Error; err != nil {
			return err
		}
		var correctCount int64
		if err := tx.Model(&entity.AttemptAnswer{}).Where("attempt_id = ? AND is_correct = ?", attempt.ID, true).Count(&correctCount).Error; err != nil {
			return err
		}
		attempt.Status = "submitted"
		attempt.SubmittedAt = &submittedAt
		attempt.CorrectCount = int(correctCount)
		if totalQuestions > 0 {
			attempt.Score = float64(correctCount) / float64(totalQuestions) * 100
		}
		if err := tx.Save(&attempt).Error; err != nil {
			return err
		}
		result = attempt
		return nil
	})
	return &result, err
}

func (r *attemptRepository) FindResultForCandidate(attemptID, candidateID string) (*entity.Attempt, bool, error) {
	var attempt entity.Attempt
	err := r.db.Table("attempts a").
		Joins("JOIN exam_assignments ea ON ea.id = a.assignment_id AND ea.deleted_at IS NULL").
		Joins("JOIN exams e ON e.id = ea.exam_id AND e.deleted_at IS NULL").
		Select("a.*, ea.exam_id AS exam_id").
		Where("a.id = ? AND ea.candidate_id = ?", attemptID, candidateID).
		First(&attempt).Error
	if err != nil {
		return nil, false, err
	}
	var showResult bool
	if err := r.db.Table("attempts a").Joins("JOIN exam_assignments ea ON ea.id = a.assignment_id").Joins("JOIN exams e ON e.id = ea.exam_id").Where("a.id = ?", attemptID).Select("e.show_result").Scan(&showResult).Error; err != nil {
		return nil, false, err
	}
	return &attempt, showResult, nil
}
