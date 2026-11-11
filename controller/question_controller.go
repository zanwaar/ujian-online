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

// List godoc
// @Summary List questions
// @Tags admin-questions
// @Produce json
// @Security BearerAuth
// @Param exam_version_id query string false "Exam version ID"
// @Param search query string false "Search term"
// @Param status query string false "Question status"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions [get]
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

// Get godoc
// @Summary Get question
// @Tags admin-questions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions/{id} [get]
func (c *QuestionController) Get(ctx *gin.Context) {
	question, err := c.service.Get(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, "Question not found", http.StatusNotFound, nil)
		return
	}
	utils.SuccessResponse(ctx, "Question loaded successfully", question)
}

// Create godoc
// @Summary Create question
// @Tags admin-questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.CreateQuestionRequest true "Question data"
// @Success 201 {object} utils.Response
// @Router /v1/admin/questions [post]
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

// Update godoc
// @Summary Update question
// @Tags admin-questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Param input body dto.UpdateQuestionRequest true "Question data"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions/{id} [put]
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

// Archive godoc
// @Summary Archive question
// @Tags admin-questions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/questions/{id} [delete]
func (c *QuestionController) Archive(ctx *gin.Context) {
	if err := c.service.Archive(ctx.Param("id")); err != nil {
		utils.ErrorResponse(ctx, "Failed to archive question", http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(ctx, "Question archived successfully", nil)
}
