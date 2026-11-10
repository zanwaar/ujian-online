package dto

type CreateCandidateRequest struct {
	ExamNumber string `json:"exam_number" binding:"required,max=100"`
	Password   string `json:"password" binding:"required,min=6,max=100"`
	Name       string `json:"name" binding:"required,max=100"`
	Gender     string `json:"gender" binding:"required,oneof=L P"`
	School     string `json:"school" binding:"required,max=100"`
	ClassName  string `json:"class_name" binding:"required,max=50"`
}

type UpdateCandidateRequest struct {
	Password  string `json:"password" binding:"omitempty,min=6,max=100"`
	Name      string `json:"name" binding:"required,max=100"`
	Gender    string `json:"gender" binding:"required,oneof=L P"`
	School    string `json:"school" binding:"required,max=100"`
	ClassName string `json:"class_name" binding:"required,max=50"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive archived"`
}
