package controller

import (
	"net/http"

	"golang-backend/dto"
	"golang-backend/service"
	"golang-backend/utils"

	"github.com/gin-gonic/gin"
)

type AttemptController struct {
	service service.AttemptService
}

func NewAttemptController(attemptService service.AttemptService) *AttemptController {
	return &AttemptController{service: attemptService}
}

func candidateID(ctx *gin.Context) (string, bool) {
	id, exists := ctx.Get("candidate_id")
	if !exists {
		return "", false
	}
	candidateID, ok := id.(string)
	return candidateID, ok && candidateID != ""
}

func (c *AttemptController) ListSchedules(ctx *gin.Context) {
	id, ok := candidateID(ctx)
	if !ok {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	schedules, err := c.service.ListSchedules(id)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to load exam schedules", http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Exam schedules loaded successfully", schedules)
}

func (c *AttemptController) Start(ctx *gin.Context) {
	id, ok := candidateID(ctx)
	if !ok {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	attempt, err := c.service.Start(id, ctx.Param("exam_id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to start exam", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Exam attempt ready", attempt)
}

func (c *AttemptController) Get(ctx *gin.Context) {
	id, ok := candidateID(ctx)
	if !ok {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	attempt, err := c.service.Get(id, ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Attempt not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Attempt loaded successfully", attempt)
}

func (c *AttemptController) Answer(ctx *gin.Context) {
	id, ok := candidateID(ctx)
	if !ok {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	var request dto.AnswerRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	if err := c.service.Answer(id, ctx.Param("id"), ctx.Param("question_id"), request.SelectedOptionKey); err != nil {
		utils.ErrorResponse(ctx, "Failed to save answer", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Answer saved successfully", nil)
}

func (c *AttemptController) Submit(ctx *gin.Context) {
	id, ok := candidateID(ctx)
	if !ok {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	attempt, err := c.service.Submit(id, ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to submit attempt", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Attempt submitted successfully", attempt)
}

func (c *AttemptController) Result(ctx *gin.Context) {
	id, ok := candidateID(ctx)
	if !ok {
		utils.ErrorResponse(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	result, err := c.service.Result(id, ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Result not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Attempt result loaded successfully", result)
}
