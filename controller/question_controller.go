package controller

import (
	"net/http"
	"strings"
	"time"

	"golang-backend/dto"
	"golang-backend/service"
	"golang-backend/utils"

	"github.com/gin-gonic/gin"
)

type QuestionController struct {
	service service.QuestionService
}

func NewQuestionController(questionService service.QuestionService) *QuestionController {
	return &QuestionController{service: questionService}
}

func (c *QuestionController) List(ctx *gin.Context) {
	page, perPage := utils.GetPaginationParams(ctx)
	filters := map[string]string{
		"exam_version_id": strings.TrimSpace(ctx.Query("exam_version_id")),
		"search":          strings.TrimSpace(ctx.Query("search")),
		"status":          strings.TrimSpace(ctx.Query("status")),
	}
	startedAt := time.Now()
	result, err := c.service.List(filters, page, perPage)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to load questions", http.StatusInternalServerError, err.Error())
		return
	}
	utils.APIResponse(ctx, "Questions loaded successfully", http.StatusOK, result.Items, nil, utils.BuildMeta(result.Pagination, time.Since(startedAt).Seconds()))
}

func (c *QuestionController) Get(ctx *gin.Context) {
	question, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Question not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Question loaded successfully", question)
}

func (c *QuestionController) Create(ctx *gin.Context) {
	var request dto.CreateQuestionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	question, err := c.service.Create(request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to create question", http.StatusBadRequest, err.Error())
		return
	}
	utils.CreatedResponse(ctx, "Question created successfully", question)
}

func (c *QuestionController) Update(ctx *gin.Context) {
	var request dto.UpdateQuestionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}
	question, err := c.service.Update(ctx.Param("id"), request)
	if err != nil {
		utils.ErrorResponse(ctx, "Failed to update question", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Question updated successfully", question)
}

func (c *QuestionController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive question", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Question archived successfully", nil)
}
