package dto

import "time"

type CreateExamSessionRequest struct {
	ExamID   string    `json:"exam_id" binding:"required"`
	Name     string    `json:"name" binding:"required,max=50"`
	StartsAt time.Time `json:"starts_at" binding:"required"`
	EndsAt   time.Time `json:"ends_at" binding:"required"`
}

type UpdateExamSessionRequest struct {
	Name     string    `json:"name" binding:"required,max=50"`
	StartsAt time.Time `json:"starts_at" binding:"required"`
	EndsAt   time.Time `json:"ends_at" binding:"required"`
	Status   string    `json:"status" binding:"omitempty,oneof=scheduled open closed archived"`
}
