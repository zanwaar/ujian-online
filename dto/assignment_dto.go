package dto

type CreateAssignmentRequest struct {
	CandidateID   string `json:"candidate_id" binding:"required"`
	ExamID        string `json:"exam_id" binding:"required"`
	ExamSessionID string `json:"exam_session_id" binding:"required"`
}
