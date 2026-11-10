package dto

type CandidateLoginRequest struct {
	ExamNumber string `json:"exam_number" binding:"required"`
	Password   string `json:"password" binding:"required"`
}
