package dto

import "time"

type CandidateExamSchedule struct {
	AssignmentID    string     `json:"assignment_id"`
	ExamID          string     `json:"exam_id"`
	ExamName        string     `json:"exam_name"`
	Subject         string     `json:"subject"`
	SessionID       string     `json:"session_id"`
	SessionName     string     `json:"session_name"`
	StartsAt        time.Time  `json:"starts_at"`
	EndsAt          time.Time  `json:"ends_at"`
	DurationMinutes int        `json:"duration_minutes"`
	ShowResult      bool       `json:"show_result"`
	AttemptID       *string    `json:"attempt_id,omitempty"`
	AttemptStatus   *string    `json:"attempt_status,omitempty"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
}

type AnswerRequest struct {
	SelectedOptionKey string `json:"selected_option_key" binding:"required,max=4"`
}

type AttemptOptionView struct {
	OptionKey   string `json:"option_key"`
	ContentHTML string `json:"content_html"`
}

type AttemptQuestionView struct {
	ID                string              `json:"id"`
	SequenceNumber    int                 `json:"sequence_number"`
	Code              string              `json:"code,omitempty"`
	ContentHTML       string              `json:"content_html"`
	Options           []AttemptOptionView `json:"options"`
	SelectedOptionKey *string             `json:"selected_option_key,omitempty"`
}

type AttemptView struct {
	ID           string                `json:"id"`
	ExamID       string                `json:"exam_id"`
	Status       string                `json:"status"`
	StartedAt    *time.Time            `json:"started_at,omitempty"`
	DeadlineAt   *time.Time            `json:"deadline_at,omitempty"`
	SubmittedAt  *time.Time            `json:"submitted_at,omitempty"`
	CorrectCount int                   `json:"correct_count"`
	Score        float64               `json:"score"`
	Questions    []AttemptQuestionView `json:"questions"`
}

type AttemptResultView struct {
	AttemptID    string     `json:"attempt_id"`
	ExamID       string     `json:"exam_id"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	SubmittedAt  *time.Time `json:"submitted_at,omitempty"`
	CorrectCount int        `json:"correct_count,omitempty"`
	Score        float64    `json:"score,omitempty"`
	ShowResult   bool       `json:"show_result"`
}

type AdminAttemptResult struct {
	AttemptID           string     `json:"attempt_id"`
	AttemptNumber       int        `json:"attempt_number"`
	CandidateID         string     `json:"candidate_id"`
	CandidateExamNumber string     `json:"candidate_exam_number"`
	CandidateName       string     `json:"candidate_name"`
	ExamID              string     `json:"exam_id"`
	ExamName            string     `json:"exam_name"`
	SessionID           string     `json:"session_id"`
	SessionName         string     `json:"session_name"`
	Status              string     `json:"status"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	SubmittedAt         *time.Time `json:"submitted_at,omitempty"`
	CorrectCount        int        `json:"correct_count"`
	Score               float64    `json:"score"`
}

type AdminAttemptOptionDetail struct {
	OptionKey   string `json:"option_key"`
	ContentHTML string `json:"content_html"`
	IsCorrect   bool   `json:"is_correct"`
}

type AdminAttemptQuestionDetail struct {
	ID                string                     `json:"id"`
	SequenceNumber    int                        `json:"sequence_number"`
	Code              string                     `json:"code,omitempty"`
	ContentHTML       string                     `json:"content_html"`
	Options           []AdminAttemptOptionDetail `json:"options"`
	CorrectOptionKey  string                     `json:"correct_option_key,omitempty"`
	SelectedOptionKey *string                    `json:"selected_option_key,omitempty"`
	AnsweredAt        *time.Time                 `json:"answered_at,omitempty"`
	IsCorrect         bool                       `json:"is_correct"`
}

type AdminAttemptDetail struct {
	AdminAttemptResult
	Questions []AdminAttemptQuestionDetail `json:"questions"`
}
