package service

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"

	"golang-backend/dto"
	"golang-backend/entity"
	"golang-backend/repository"
	"golang-backend/utils"

	"gorm.io/gorm"
)

type AttemptService interface {
	ListSchedules(candidateID string) ([]dto.CandidateExamSchedule, error)
	Start(candidateID, examID string) (*dto.AttemptView, error)
	Get(candidateID, attemptID string) (*dto.AttemptView, error)
	Answer(candidateID, attemptID, questionID, selectedOptionKey string) error
	Submit(candidateID, attemptID string) (*dto.AttemptView, error)
	Result(candidateID, attemptID string) (*dto.AttemptResultView, error)
	ListAdminResults(filters map[string]string, page, perPage int) (*utils.PaginationResult, error)
	GetAdminResultDetail(attemptID string) (*dto.AdminAttemptDetail, error)
}

type attemptService struct {
	repo repository.AttemptRepository
}

func NewAttemptService(repo repository.AttemptRepository) AttemptService {
	return &attemptService{repo: repo}
}

func (s *attemptService) ListSchedules(candidateID string) ([]dto.CandidateExamSchedule, error) {
	return s.repo.ListSchedules(candidateID)
}

func (s *attemptService) ListAdminResults(filters map[string]string, page, perPage int) (*utils.PaginationResult, error) {
	return s.repo.ListAdminResults(filters, page, perPage)
}

func (s *attemptService) GetAdminResultDetail(attemptID string) (*dto.AdminAttemptDetail, error) {
	data, err := s.repo.FindAdminResultDetail(attemptID)
	if err != nil {
		return nil, err
	}

	answers := make(map[string]entity.AttemptAnswer, len(data.Answers))
	for _, answer := range data.Answers {
		answers[answer.AttemptQuestionID] = answer
	}

	detail := &dto.AdminAttemptDetail{
		AdminAttemptResult: data.Result,
		Questions:          make([]dto.AdminAttemptQuestionDetail, 0, len(data.Questions)),
	}
	for _, question := range data.Questions {
		var snapshot questionSnapshot
		if err := json.Unmarshal([]byte(question.SnapshotContentJSON), &snapshot); err != nil {
			return nil, err
		}
		questionDetail := dto.AdminAttemptQuestionDetail{
			ID:               question.ID,
			SequenceNumber:   question.SequenceNumber,
			Code:             snapshot.Code,
			ContentHTML:      snapshot.ContentHTML,
			CorrectOptionKey: snapshot.CorrectOptionKey,
			Options:          make([]dto.AdminAttemptOptionDetail, 0, len(snapshot.Options)),
		}
		for _, option := range snapshot.Options {
			questionDetail.Options = append(questionDetail.Options, dto.AdminAttemptOptionDetail{
				OptionKey:   option.OptionKey,
				ContentHTML: option.ContentHTML,
				IsCorrect:   option.OptionKey == snapshot.CorrectOptionKey,
			})
		}
		if answer, ok := answers[question.ID]; ok {
			if answer.SelectedOptionKey != "" {
				selectedOptionKey := answer.SelectedOptionKey
				questionDetail.SelectedOptionKey = &selectedOptionKey
			}
			questionDetail.AnsweredAt = answer.AnsweredAt
			questionDetail.IsCorrect = answer.IsCorrect
		}
		detail.Questions = append(detail.Questions, questionDetail)
	}
	return detail, nil
}

type snapshotOption struct {
	OptionKey   string `json:"option_key"`
	ContentHTML string `json:"content_html"`
}

type questionSnapshot struct {
	Code             string           `json:"code"`
	ContentHTML      string           `json:"content_html"`
	CorrectOptionKey string           `json:"correct_option_key"`
	Options          []snapshotOption `json:"options"`
}

func (s *attemptService) Start(candidateID, examID string) (*dto.AttemptView, error) {
	context, err := s.repo.FindStartContext(candidateID, examID)
	if err != nil {
		return nil, errors.New("exam assignment not found")
	}
	now := time.Now().UTC()
	if now.Before(context.SessionStartsAt) {
		return nil, errors.New("exam session has not started")
	}
	if !now.Before(context.SessionEndsAt) {
		return nil, errors.New("exam session has ended")
	}

	if active, err := s.repo.FindActiveAttempt(context.AssignmentID); err == nil {
		return s.toAttemptView(active)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if _, err := s.repo.FindCompletedAttempt(context.AssignmentID); err == nil {
		return nil, errors.New("exam attempt has already been submitted")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	questions, err := s.repo.FindQuestions(context.VersionID)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, errors.New("exam has no active questions")
	}
	attemptNumber, err := s.repo.CountAttempts(context.AssignmentID)
	if err != nil {
		return nil, err
	}
	deadline := now.Add(time.Duration(context.DurationMinutes) * time.Minute)
	if context.SessionEndsAt.Before(deadline) {
		deadline = context.SessionEndsAt
	}
	attempt := &entity.Attempt{
		AssignmentID:  context.AssignmentID,
		AttemptNumber: int(attemptNumber) + 1,
		StartedAt:     &now,
		DeadlineAt:    &deadline,
		Status:        "in_progress",
	}
	randomizer := rand.New(rand.NewSource(now.UnixNano()))
	if context.ShuffleQuestions {
		randomizer.Shuffle(len(questions), func(i, j int) { questions[i], questions[j] = questions[j], questions[i] })
	}
	snapshots := make([]entity.AttemptQuestion, 0, len(questions))
	for index, question := range questions {
		options := make([]snapshotOption, 0, len(question.Options))
		optionKeys := make([]string, 0, len(question.Options))
		correctOptionKey := ""
		for _, option := range question.Options {
			options = append(options, snapshotOption{OptionKey: option.OptionKey, ContentHTML: option.ContentHTML})
			optionKeys = append(optionKeys, option.OptionKey)
			if option.IsCorrect {
				correctOptionKey = option.OptionKey
			}
		}
		if context.ShuffleOptions {
			randomizer.Shuffle(len(options), func(i, j int) {
				options[i], options[j] = options[j], options[i]
				optionKeys[i], optionKeys[j] = optionKeys[j], optionKeys[i]
			})
		}
		snapshot, err := json.Marshal(questionSnapshot{
			Code:             question.Code,
			ContentHTML:      question.ContentHTML,
			CorrectOptionKey: correctOptionKey,
			Options:          options,
		})
		if err != nil {
			return nil, err
		}
		order, err := json.Marshal(optionKeys)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, entity.AttemptQuestion{
			QuestionID:          question.ID,
			SequenceNumber:      index + 1,
			OptionOrderJSON:     string(order),
			SnapshotContentJSON: string(snapshot),
		})
	}
	if err := s.repo.CreateAttemptWithSnapshot(attempt, snapshots); err != nil {
		return nil, err
	}
	return s.Get(candidateID, attempt.ID)
}

func (s *attemptService) Get(candidateID, attemptID string) (*dto.AttemptView, error) {
	attempt, err := s.repo.FindAttemptForCandidate(attemptID, candidateID)
	if err != nil {
		return nil, err
	}
	if attempt.Status == "in_progress" && attempt.DeadlineAt != nil && !time.Now().UTC().Before(*attempt.DeadlineAt) {
		attempt, err = s.repo.SubmitAttempt(attemptID, candidateID, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		attempt, err = s.repo.FindAttemptForCandidate(attemptID, candidateID)
		if err != nil {
			return nil, err
		}
	}
	return s.toAttemptView(attempt)
}

func (s *attemptService) Answer(candidateID, attemptID, questionID, selectedOptionKey string) error {
	selectedOptionKey = strings.ToUpper(strings.TrimSpace(selectedOptionKey))
	if selectedOptionKey == "" {
		return errors.New("selected option is required")
	}
	return s.repo.SaveAnswer(attemptID, candidateID, questionID, selectedOptionKey, time.Now().UTC())
}

func (s *attemptService) Submit(candidateID, attemptID string) (*dto.AttemptView, error) {
	if _, err := s.repo.SubmitAttempt(attemptID, candidateID, time.Now().UTC()); err != nil {
		return nil, err
	}
	return s.Get(candidateID, attemptID)
}

func (s *attemptService) Result(candidateID, attemptID string) (*dto.AttemptResultView, error) {
	attempt, showResult, err := s.repo.FindResultForCandidate(attemptID, candidateID)
	if err != nil {
		return nil, err
	}
	if attempt.Status != "submitted" {
		return nil, errors.New("attempt has not been submitted")
	}
	result := &dto.AttemptResultView{
		AttemptID:   attempt.ID,
		ExamID:      attempt.ExamID,
		Status:      attempt.Status,
		StartedAt:   attempt.StartedAt,
		SubmittedAt: attempt.SubmittedAt,
		ShowResult:  showResult,
	}
	if showResult {
		result.CorrectCount = attempt.CorrectCount
		result.Score = attempt.Score
	}
	return result, nil
}

func (s *attemptService) toAttemptView(attempt *entity.Attempt) (*dto.AttemptView, error) {
	view := &dto.AttemptView{
		ID:           attempt.ID,
		ExamID:       attempt.ExamID,
		Status:       attempt.Status,
		StartedAt:    attempt.StartedAt,
		DeadlineAt:   attempt.DeadlineAt,
		SubmittedAt:  attempt.SubmittedAt,
		CorrectCount: attempt.CorrectCount,
		Score:        attempt.Score,
		Questions:    make([]dto.AttemptQuestionView, 0, len(attempt.Questions)),
	}
	for _, question := range attempt.Questions {
		var snapshot questionSnapshot
		if err := json.Unmarshal([]byte(question.SnapshotContentJSON), &snapshot); err != nil {
			return nil, err
		}
		questionView := dto.AttemptQuestionView{
			ID:             question.ID,
			SequenceNumber: question.SequenceNumber,
			Code:           snapshot.Code,
			ContentHTML:    snapshot.ContentHTML,
			Options:        make([]dto.AttemptOptionView, 0, len(snapshot.Options)),
		}
		for _, option := range snapshot.Options {
			questionView.Options = append(questionView.Options, dto.AttemptOptionView{OptionKey: option.OptionKey, ContentHTML: option.ContentHTML})
		}
		for _, answer := range attempt.Answers {
			if answer.AttemptQuestionID == question.ID {
				selected := answer.SelectedOptionKey
				questionView.SelectedOptionKey = &selected
				break
			}
		}
		view.Questions = append(view.Questions, questionView)
	}
	return view, nil
}
