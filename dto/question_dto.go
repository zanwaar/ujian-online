package dto

type QuestionOptionRequest struct {
	OptionKey    string `json:"option_key" binding:"required,max=4"`
	ContentHTML  string `json:"content_html" binding:"required"`
	IsCorrect    bool   `json:"is_correct"`
	DisplayOrder int    `json:"display_order"`
}

type CreateQuestionRequest struct {
	ExamVersionID string                  `json:"exam_version_id" binding:"required"`
	Code          string                  `json:"code" binding:"omitempty,max=50"`
	ContentHTML   string                  `json:"content_html" binding:"required"`
	Weight        float64                 `json:"weight" binding:"required,gt=0"`
	DisplayOrder  int                     `json:"display_order"`
	Options       []QuestionOptionRequest `json:"options" binding:"required,min=2,max=5"`
}

type UpdateQuestionRequest struct {
	Code         string                  `json:"code" binding:"omitempty,max=50"`
	ContentHTML  string                  `json:"content_html" binding:"required"`
	Weight       float64                 `json:"weight" binding:"required,gt=0"`
	DisplayOrder int                     `json:"display_order"`
	Status       string                  `json:"status" binding:"omitempty,oneof=active archived"`
	Options      []QuestionOptionRequest `json:"options" binding:"required,min=2,max=5"`
}
