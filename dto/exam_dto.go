package dto

type CreateExamRequest struct {
	Code             string `json:"code" binding:"required,max=50"`
	Name             string `json:"name" binding:"required,max=100"`
	Subject          string `json:"subject" binding:"required,max=100"`
	DescriptionHTML  string `json:"description_html" binding:"required"`
	DurationMinutes  int    `json:"duration_minutes" binding:"required,min=1,max=1440"`
	ShowResult       bool   `json:"show_result"`
	ShuffleQuestions bool   `json:"shuffle_questions"`
	ShuffleOptions   bool   `json:"shuffle_options"`
}

type UpdateExamRequest struct {
	Name             string `json:"name" binding:"required,max=100"`
	Subject          string `json:"subject" binding:"required,max=100"`
	DescriptionHTML  string `json:"description_html" binding:"required"`
	DurationMinutes  int    `json:"duration_minutes" binding:"required,min=1,max=1440"`
	ShowResult       bool   `json:"show_result"`
	ShuffleQuestions bool   `json:"shuffle_questions"`
	ShuffleOptions   bool   `json:"shuffle_options"`
	Status           string `json:"status" binding:"omitempty,oneof=draft published archived"`
}
